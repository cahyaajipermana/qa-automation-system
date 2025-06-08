package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"gorm.io/gorm"
	"qa-automation-system/backend/config"
	"qa-automation-system/backend/models"
)

// TestStep represents a single test step
type TestStep struct {
	Action   string
	Selector string
	Value    string
}

// TestCase represents a complete test case
type TestCase struct {
	Feature string
	Site    string
	URL     string
	Steps   []TestStep
}

// TestResult represents the result of a test
type TestResult struct {
	ID         string    `json:"id"`
	Feature    string    `json:"feature"`
	Site       string    `json:"site"`
	Browser    string    `json:"browser"`
	Device     string    `json:"device"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
	ErrorLog   string    `json:"errorLog,omitempty"`
	Screenshot string    `json:"screenshot,omitempty"`
}

// BrowserStackRunner handles test execution using BrowserStack
type BrowserStackRunner struct {
	config      *config.BrowserStackConfig
	wd          selenium.WebDriver
	testResults []TestResult
	db          *gorm.DB
}

// NewBrowserStackRunner creates a new BrowserStack runner
func NewBrowserStackRunner() *BrowserStackRunner {
	db, err := config.InitDB()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize database: %v\n", err)
	}
	
	return &BrowserStackRunner{
		config:      config.NewBrowserStackConfig(),
		testResults: make([]TestResult, 0),
		db:          db,
	}
}

// StoreTestResult stores the test result in the database
func (r *BrowserStackRunner) StoreTestResult(result *TestResult) error {
	if r.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Create result record
	dbResult := models.Result{
		Feature:    result.Feature,
		Site:       result.Site,
		Browser:    result.Browser,
		Device:     result.Device,
		Status:     result.Status,
		Timestamp:  result.Timestamp,
		ErrorLog:   result.ErrorLog,
		Screenshot: result.Screenshot,
	}

	if err := r.db.Create(&dbResult).Error; err != nil {
		return fmt.Errorf("failed to store test result: %v", err)
	}

	// Store result details
	for _, step := range r.testResults {
		detail := models.ResultDetail{
			ResultID:   dbResult.ID,
			Step:       step.Feature,
			Status:     step.Status,
			ErrorLog:   step.ErrorLog,
			Screenshot: step.Screenshot,
			Timestamp:  step.Timestamp,
		}

		if err := r.db.Create(&detail).Error; err != nil {
			return fmt.Errorf("failed to store result detail: %v", err)
		}
	}

	return nil
}

// Initialize sets up the browser for testing
func (r *BrowserStackRunner) Initialize(browserType string) error {
	capabilities := make(map[string]interface{})
	
	// Set BrowserStack credentials
	capabilities["browserstack.user"] = r.config.Username
	capabilities["browserstack.key"] = r.config.AccessKey

	// Set base capabilities
	for k, v := range r.config.Capabilities {
		capabilities[k] = v
	}

	// Set browser-specific capabilities
	if browserCapabilities, ok := r.config.Browsers[browserType]; ok {
		for k, v := range browserCapabilities {
			capabilities[k] = v
		}
	}

	// Connect to BrowserStack
	wd, err := selenium.NewRemote(capabilities, fmt.Sprintf("http://hub.browserstack.com/wd/hub"))
	if err != nil {
		return fmt.Errorf("failed to create remote session: %v", err)
	}

	r.wd = wd
	return nil
}

// LoginToSentiLive performs login to senti.live
func (r *BrowserStackRunner) LoginToSentiLive(email, password string) error {
	// Navigate to login page
	if err := r.wd.Get("https://senti.live/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %v", err)
	}

	// Wait for the page to load
	time.Sleep(2 * time.Second)

	// Find and fill email input
	emailInput, err := r.wd.FindElement(selenium.ByCSSSelector, "input[type='email']")
	if err != nil {
		return fmt.Errorf("failed to find email input: %v", err)
	}
	if err := emailInput.SendKeys(email); err != nil {
		return fmt.Errorf("failed to enter email: %v", err)
	}

	// Find and fill password input
	passwordInput, err := r.wd.FindElement(selenium.ByCSSSelector, "input[type='password']")
	if err != nil {
		return fmt.Errorf("failed to find password input: %v", err)
	}
	if err := passwordInput.SendKeys(password); err != nil {
		return fmt.Errorf("failed to enter password: %v", err)
	}

	// Find and click login button
	loginButton, err := r.findElementByText("Login")
	if err != nil {
		return fmt.Errorf("failed to find login button: %v", err)
	}
	if err := loginButton.Click(); err != nil {
		return fmt.Errorf("failed to click login button: %v", err)
	}

	// Wait for login to complete
	time.Sleep(3 * time.Second)

	// Verify login success by checking for chat page elements
	_, err = r.wd.FindElement(selenium.ByCSSSelector, ".chat-container")
	if err != nil {
		return fmt.Errorf("login verification failed: %v", err)
	}

	return nil
}

// NavigateToChatPage navigates to the chat page
func (r *BrowserStackRunner) NavigateToChatPage() error {
	if err := r.wd.Get("https://senti.live/chat"); err != nil {
		return fmt.Errorf("failed to navigate to chat page: %v", err)
	}

	// Wait for chat page to load
	time.Sleep(2 * time.Second)

	// Verify we're on the chat page
	_, err := r.wd.FindElement(selenium.ByCSSSelector, ".chat-container")
	if err != nil {
		return fmt.Errorf("failed to verify chat page: %v", err)
	}

	return nil
}

// ExecuteStep executes a single test step
func (r *BrowserStackRunner) ExecuteStep(step TestStep) error {
	switch step.Action {
	case "navigate":
		return r.wd.Get(step.Value)
	case "click":
		element, err := r.findElementByText(step.Value)
		if err != nil {
			return fmt.Errorf("failed to find element with text '%s': %v", step.Value, err)
		}
		return element.Click()
	case "type":
		element, err := r.wd.FindElement(selenium.ByCSSSelector, step.Selector)
		if err != nil {
			return fmt.Errorf("failed to find element: %v", err)
		}
		return element.SendKeys(step.Value)
	case "wait":
		duration, err := time.ParseDuration(step.Value)
		if err != nil {
			return fmt.Errorf("invalid duration: %v", err)
		}
		time.Sleep(duration)
		return nil
	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}
}

// findElementByText finds an element by its inner text content
func (r *BrowserStackRunner) findElementByText(text string) (selenium.WebElement, error) {
	// Try different selectors to find the element
	selectors := []string{
		fmt.Sprintf("//*[contains(text(), '%s')]", text),
		fmt.Sprintf("//*[normalize-space(text())='%s']", text),
		fmt.Sprintf("//button[contains(text(), '%s')]", text),
		fmt.Sprintf("//a[contains(text(), '%s')]", text),
		fmt.Sprintf("//div[contains(text(), '%s')]", text),
		fmt.Sprintf("//span[contains(text(), '%s')]", text),
	}

	for _, selector := range selectors {
		element, err := r.wd.FindElement(selenium.ByXPATH, selector)
		if err == nil {
			return element, nil
		}
	}

	// If no element found with text, try finding by innerHTML
	js := fmt.Sprintf(`
		function findElementByInnerHTML(text) {
			const elements = document.getElementsByTagName('*');
			for (let element of elements) {
				if (element.innerHTML.includes(text)) {
					return element;
				}
			}
			return null;
		}
		return findElementByInnerHTML('%s');
	`, text)

	result, err := r.wd.ExecuteScript(js, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute JavaScript: %v", err)
	}

	if result == nil {
		return nil, fmt.Errorf("no element found with text '%s'", text)
	}

	// Convert the JavaScript element to a WebElement
	element, err := r.wd.FindElement(selenium.ByCSSSelector, fmt.Sprintf("*:contains('%s')", text))
	if err != nil {
		return nil, fmt.Errorf("failed to convert JavaScript element to WebElement: %v", err)
	}

	return element, nil
}

// TakeScreenshot takes a screenshot and saves it
func (r *BrowserStackRunner) TakeScreenshot() (string, error) {
	screenshot, err := r.wd.Screenshot()
	if err != nil {
		return "", fmt.Errorf("failed to take screenshot: %v", err)
	}

	// Create screenshots directory if it doesn't exist
	screenshotsDir := "screenshots"
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s/screenshot_%s.png", screenshotsDir, timestamp)
	if err := os.WriteFile(filename, screenshot, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %v", err)
	}

	return filename, nil
}

// LogTestStep logs a test step with timestamp
func (r *BrowserStackRunner) LogTestStep(step string) error {
	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Generate log filename with date
	today := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("%s/test_%s.log", logsDir, today)

	// Format log entry with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, step)

	// Append to log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	return nil
}

// Close closes the browser session
func (r *BrowserStackRunner) Close() error {
	if r.wd != nil {
		return r.wd.Quit()
	}
	return nil
}

// RunTestInBackground runs the test in the background
func RunTestInBackground(resultID uint, email, password string) {
	go func() {
		// Create a new BrowserStack runner
		runner := NewBrowserStackRunner()

		// Initialize the runner with Chrome browser
		if err := runner.Initialize("chrome"); err != nil {
			logError(resultID, fmt.Sprintf("Failed to initialize runner: %v", err))
			return
		}
		defer runner.Close()

		// Create test result
		testResult := &TestResult{
			Feature:   "Login",
			Site:      "senti.live",
			Browser:   "chrome",
			Device:    "desktop",
			Status:    "running",
			Timestamp: time.Now(),
		}

		// Log test start
		if err := runner.LogTestStep("Test started - Initializing browser"); err != nil {
			log.Printf("Warning: Failed to log test start: %v", err)
		}

		// Take screenshot before login
		beforeLoginScreenshot, err := runner.TakeScreenshot()
		if err != nil {
			log.Printf("Warning: Failed to take before login screenshot: %v", err)
		} else {
			log.Printf("Before login screenshot saved: %s", beforeLoginScreenshot)
			if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken before login: %s", beforeLoginScreenshot)); err != nil {
				log.Printf("Warning: Failed to log screenshot: %v", err)
			}
		}

		// Perform login
		log.Println("Attempting to login to senti.live...")
		if err := runner.LogTestStep("Attempting to login to senti.live"); err != nil {
			log.Printf("Warning: Failed to log login attempt: %v", err)
		}

		if err := runner.LoginToSentiLive(email, password); err != nil {
			logMsg := fmt.Sprintf("Login failed: %v", err)
			logError(resultID, logMsg)
			if err := runner.LogTestStep(logMsg); err != nil {
				log.Printf("Warning: Failed to log login failure: %v", err)
			}
			testResult.Status = "failed"
			testResult.ErrorLog = logMsg
			testResult.Screenshot = beforeLoginScreenshot
			if err := runner.StoreTestResult(testResult); err != nil {
				log.Printf("Warning: Failed to store test result: %v", err)
			}
			return
		}
		
		if err := runner.LogTestStep("Login successful"); err != nil {
			log.Printf("Warning: Failed to log successful login: %v", err)
		}
		log.Println("Login successful!")

		// Take screenshot after login
		afterLoginScreenshot, err := runner.TakeScreenshot()
		if err != nil {
			log.Printf("Warning: Failed to take after login screenshot: %v", err)
		} else {
			log.Printf("After login screenshot saved: %s", afterLoginScreenshot)
			if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken after login: %s", afterLoginScreenshot)); err != nil {
				log.Printf("Warning: Failed to log screenshot: %v", err)
			}
		}

		// Navigate to chat page
		log.Println("Navigating to chat page...")
		if err := runner.LogTestStep("Navigating to chat page"); err != nil {
			log.Printf("Warning: Failed to log navigation attempt: %v", err)
		}

		if err := runner.NavigateToChatPage(); err != nil {
			logMsg := fmt.Sprintf("Failed to navigate to chat page: %v", err)
			logError(resultID, logMsg)
			if err := runner.LogTestStep(logMsg); err != nil {
				log.Printf("Warning: Failed to log navigation failure: %v", err)
			}
			testResult.Status = "failed"
			testResult.ErrorLog = logMsg
			testResult.Screenshot = afterLoginScreenshot
			if err := runner.StoreTestResult(testResult); err != nil {
				log.Printf("Warning: Failed to store test result: %v", err)
			}
			return
		}

		if err := runner.LogTestStep("Successfully navigated to chat page"); err != nil {
			log.Printf("Warning: Failed to log successful navigation: %v", err)
		}
		log.Println("Successfully navigated to chat page!")

		// Take screenshot of chat page
		chatPageScreenshot, err := runner.TakeScreenshot()
		if err != nil {
			log.Printf("Warning: Failed to take chat page screenshot: %v", err)
		} else {
			log.Printf("Chat page screenshot saved: %s", chatPageScreenshot)
			if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken of chat page: %s", chatPageScreenshot)); err != nil {
				log.Printf("Warning: Failed to log screenshot: %v", err)
			}
		}

		// Keep the session alive for a while to verify everything is working
		log.Println("Keeping session alive for 30 seconds...")
		if err := runner.LogTestStep("Keeping session alive for 30 seconds"); err != nil {
			log.Printf("Warning: Failed to log session wait: %v", err)
		}
		time.Sleep(30 * time.Second)

		// Update and store final test result
		testResult.Status = "passed"
		testResult.Screenshot = chatPageScreenshot
		if err := runner.StoreTestResult(testResult); err != nil {
			log.Printf("Warning: Failed to store test result: %v", err)
		}

		if err := runner.LogTestStep("Test completed successfully"); err != nil {
			log.Printf("Warning: Failed to log test completion: %v", err)
		}
		log.Println("Test completed successfully!")
	}()
}

// logError updates the result status and error log in the database
func logError(resultID uint, errorMsg string) {
	db, err := config.InitDB()
	if err != nil {
		log.Printf("Failed to initialize database for error logging: %v", err)
		return
	}

	var result models.Result
	if err := db.First(&result, resultID).Error; err != nil {
		log.Printf("Failed to find result for error logging: %v", err)
		return
	}

	result.Status = "failed"
	result.ErrorLog = errorMsg
	if err := db.Save(&result).Error; err != nil {
		log.Printf("Failed to update result with error: %v", err)
	}
} 