package testrunner

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"gorm.io/gorm"
	"qa-automation-system/backend/config"
	"qa-automation-system/backend/models"
)

// BrowserStackRunner handles browser automation using BrowserStack
type BrowserStackRunner struct {
	driver selenium.WebDriver
	config *BrowserStackConfig
	db     *gorm.DB
}

// BrowserStackConfig holds BrowserStack configuration
type BrowserStackConfig struct {
	Username    string
	AccessKey   string
	Browsers    map[string]map[string]interface{}
	BaseURL     string
	ProjectName string
	BuildName   string
}

// TestResult represents a test execution result
type TestResult struct {
	Feature   string
	Site      string
	Browser   string
	Device    string
	Status    string
	ErrorLog  string
	Screenshot string
	Timestamp time.Time
}

// NewBrowserStackRunner creates a new BrowserStack runner instance
func NewBrowserStackRunner() *BrowserStackRunner {
	return &BrowserStackRunner{
		config: &BrowserStackConfig{
			Username:  os.Getenv("BROWSERSTACK_USERNAME"),
			AccessKey: os.Getenv("BROWSERSTACK_ACCESS_KEY"),
			Browsers: map[string]map[string]interface{}{
				"chrome": {
					"browserName": "Chrome",
					"browserVersion": "latest",
					"os": "Windows",
					"osVersion": "10",
				},
				"firefox": {
					"browserName": "Firefox",
					"browserVersion": "latest",
					"os": "Windows",
					"osVersion": "10",
				},
				"edge": {
					"browserName": "Edge",
					"browserVersion": "latest",
					"os": "Windows",
					"osVersion": "10",
				},
				"safari": {
					"browserName": "Safari",
					"browserVersion": "latest",
					"os": "OS X",
					"osVersion": "Big Sur",
				},
			},
			BaseURL:     "https://hub.browserstack.com/wd/hub",
			ProjectName: "QA Automation System",
			BuildName:   "Test Run " + time.Now().Format("2006-01-02 15:04:05"),
		},
	}
}

// Initialize sets up the browser session
func (r *BrowserStackRunner) Initialize(browserType string) error {
	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	r.db = db

	// Set browser-specific capabilities
	if browserCapabilities, ok := r.config.Browsers[browserType]; ok {
		caps := selenium.Capabilities{
			"bstack:options": map[string]interface{}{
				"userName":    r.config.Username,
				"accessKey":   r.config.AccessKey,
				"projectName": r.config.ProjectName,
				"buildName":   r.config.BuildName,
				"sessionName": fmt.Sprintf("%s Test", browserType),
			},
		}

		// Add browser-specific capabilities
		for k, v := range browserCapabilities {
			caps[k] = v
		}

		// Initialize WebDriver
		driver, err := selenium.NewRemote(caps, r.config.BaseURL)
		if err != nil {
			return fmt.Errorf("failed to initialize WebDriver: %v", err)
		}
		r.driver = driver
		return nil
	}
	return fmt.Errorf("unsupported browser type: %s", browserType)
}

// Close closes the WebDriver session
func (r *BrowserStackRunner) Close() error {
	if r.driver != nil {
		if err := r.driver.Quit(); err != nil {
			return fmt.Errorf("failed to quit WebDriver: %v", err)
		}
	}
	return nil
}

// TakeScreenshot captures the current screen
func (r *BrowserStackRunner) TakeScreenshot() (string, error) {
	if r.driver == nil {
		return "", fmt.Errorf("driver not initialized")
	}

	// Create screenshots directory if it doesn't exist
	if err := os.MkdirAll("screenshots", 0755); err != nil {
		return "", fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	// Take screenshot
	screenshot, err := r.driver.Screenshot()
	if err != nil {
		return "", fmt.Errorf("failed to take screenshot: %v", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("screenshots/screenshot_%s.png", time.Now().Format("20060102_150405"))

	// Save screenshot to file
	if err := os.WriteFile(filename, screenshot, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %v", err)
	}

	return filename, nil
}

// LogTestStep logs a test step
func (r *BrowserStackRunner) LogTestStep(step string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// // Generate log filename with timestamp
	// filename := fmt.Sprintf("logs/test_%s.log", time.Now().Format("20060102_150405"))

	// // Append log entry
	// logEntry := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), step)
	// if err := os.WriteFile(filename, []byte(logEntry), 0644); err != nil {
	// 	return fmt.Errorf("failed to write log: %v", err)
	// }

	return nil
}

// StoreTestResult stores the test result in the database
func (r *BrowserStackRunner) StoreTestResult(result *TestResult) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create result record
	dbResult := models.Result{
		Feature:  models.Feature{Name: result.Feature},
		Site:     models.Site{Name: result.Site},
		Browser:  result.Browser,
		Device:   models.Device{Name: result.Device},
		Status:   result.Status,
		ErrorLog: result.ErrorLog,
	}

	if err := r.db.Create(&dbResult).Error; err != nil {
		return fmt.Errorf("failed to create result: %v", err)
	}

	// Create result detail record
	if result.Screenshot != "" {
		resultDetail := models.ResultDetail{
			ResultID:   dbResult.ID,
			Screenshot: result.Screenshot,
		}

		if err := r.db.Create(&resultDetail).Error; err != nil {
			return fmt.Errorf("failed to create result detail: %v", err)
		}
	}

	return nil
}

// LoginToSentiLive performs login to site
func (r *BrowserStackRunner) LoginToSentiLive(siteName string, email, password string) error {
	if r.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	emailSelector := "#input-7"
	passwordSelector := "#input-9"

	if siteName == "hothinge.com" {
		emailSelector = "#input-19"
		passwordSelector = "#input-21"
	}

	// Find and fill email field
	emailField, err := r.driver.FindElement(selenium.ByCSSSelector, emailSelector)
	if err != nil {
		return fmt.Errorf("failed to find email field: %v", err)
	}
	if err := emailField.Clear(); err != nil {
		return fmt.Errorf("failed to clear email field: %v", err)
	}
	if err := emailField.SendKeys(email); err != nil {
		return fmt.Errorf("failed to enter email: %v", err)
	}

	// Find and fill password field
	passwordField, err := r.driver.FindElement(selenium.ByCSSSelector, passwordSelector)
	if err != nil {
		return fmt.Errorf("failed to find password field: %v", err)
	}
	if err := passwordField.Clear(); err != nil {
		return fmt.Errorf("failed to clear password field: %v", err)
	}
	if err := passwordField.SendKeys(password); err != nil {
		return fmt.Errorf("failed to enter password: %v", err)
	}

	// Find and click submit button
	submitButton, err := r.driver.FindElement(selenium.ByCSSSelector, "#btn-register")
	if err != nil {
		return fmt.Errorf("failed to find submit button: %v", err)
	}
	if err := submitButton.Click(); err != nil {
		return fmt.Errorf("failed to click submit button: %v", err)
	}

	// Wait for login to complete and page to be ready
	time.Sleep(10 * time.Second)

	// Verify login success by checking if we're still on the login page
	currentURL, err := r.driver.CurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get current URL: %v", err)
	}
	if currentURL == "https://" + siteName + "/login" {
		return fmt.Errorf("login failed: still on login page")
	}

	return nil
}

// NavigateToLoginPage navigates to the chat page
func (r *BrowserStackRunner) NavigateToLoginPage(siteName string) error {
	if r.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	// Navigate to login page
	if err := r.driver.Get("https://" + siteName + "/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %v", err)
	}

	// Wait for the page to load
	time.Sleep(2 * time.Second)

	// Click the Login Button
	loginButton, err := r.driver.FindElement(selenium.ByCSSSelector, ".login-text")
	if err != nil {
		return fmt.Errorf("failed to find 'Or, Login' button: %v", err)
	}
	if err := loginButton.Click(); err != nil {
		return fmt.Errorf("failed to click 'Or, Login' button: %v", err)
	}

	// Wait for login page to be ready
	time.Sleep(2 * time.Second)

	// Verify we're on the login page
	currentURL, err := r.driver.CurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get current URL: %v", err)
	}
	if currentURL != "https://" + siteName + "/login" {
		return fmt.Errorf("navigation failed: not on login page, current URL: %s", currentURL)
	}

	return nil
}

// NavigateToChatPage navigates to the chat page
func (r *BrowserStackRunner) NavigateToChatPage(siteName string) error {
	if r.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	// Navigate to chat page
	if err := r.driver.Get("https://" + siteName + "/chat"); err != nil {
		return fmt.Errorf("failed to navigate to chat page: %v", err)
	}

	// Wait for chat page to load and be ready
	time.Sleep(3 * time.Second)

	// Verify we're on the chat page
	currentURL, err := r.driver.CurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get current URL: %v", err)
	}
	if currentURL != "https://" + siteName + "/chat" {
		return fmt.Errorf("navigation failed: not on chat page, current URL: %s", currentURL)
	}

	return nil
}

// Navigate to Open Chat
func (r *BrowserStackRunner) NavigateToOpenChat(siteName string) error {
	if r.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	chatRestID := os.Getenv("SENTI_CHAT_REST_ID")

	if siteName == "hothinge.com" {
		chatRestID = os.Getenv("HOTHINGE_CHAT_REST_ID")
	}

	if siteName == "shorts.senti.live" {
		chatRestID = os.Getenv("SHORTS_SENTI_CHAT_REST_ID")
	}

	// Navigate to chat rest page
	if err := r.driver.Get("https://" + siteName + "/chat-rest/" + chatRestID); err != nil {
		return fmt.Errorf("failed to navigate to chat rest page: %v", err)
	}

	// Wait for the open chat page to load
	time.Sleep(5 * time.Second)

	// Verify we're on the chat rest page
	currentURL, err := r.driver.CurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get current URL: %v", err)
	}
	if currentURL != "https://" + siteName + "/chat-rest/" + chatRestID {
		return fmt.Errorf("navigation failed: not on chat rest page, current URL: %s", currentURL)
	}

	return nil
}

// Sending Message to Chat
func (r *BrowserStackRunner) SendingMessageToChat(siteName string) error {
	if r.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	chatRestID := os.Getenv("SENTI_CHAT_REST_ID")

	if siteName == "hothinge.com" {
		chatRestID = os.Getenv("HOTHINGE_CHAT_REST_ID")
	}

	if siteName == "shorts.senti.live" {
		chatRestID = os.Getenv("SHORTS_SENTI_CHAT_REST_ID")
	}

	// Find and fill message field
	emailField, err := r.driver.FindElement(selenium.ByCSSSelector, ".v-field__input")
	if err != nil {
		return fmt.Errorf("failed to find message field: %v", err)
	}
	if err := emailField.Clear(); err != nil {
		return fmt.Errorf("failed to clear message field: %v", err)
	}
	if err := emailField.SendKeys("Chat send on " + time.Now().Format("2006-01-02 15:04:05")); err != nil {
		return fmt.Errorf("failed to enter message: %v", err)
	}

	// Find and click send button
	submitButton, err := r.driver.FindElement(selenium.ByCSSSelector, ".mdi-send")
	if err != nil {
		return fmt.Errorf("failed to find send button: %v", err)
	}
	if err := submitButton.Click(); err != nil {
		return fmt.Errorf("failed to click send button: %v", err)
	}

	// Wait for send chat to complete and page to be ready
	time.Sleep(20 * time.Second)

	// Verify we're on the chat rest page
	currentURL, err := r.driver.CurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get current URL: %v", err)
	}
	if currentURL != "https://" + siteName + "/chat-rest/" + chatRestID {
		return fmt.Errorf("navigation failed: not on chat rest page, current URL: %s", currentURL)
	}

	return nil
}

// RunTestInBackground runs the test in the background for multiple browsers
func RunTestInBackground(siteID, deviceID, featureID uint, email, password string) {
	// Define browsers to test
	browsers := []string{"chrome", "firefox", "edge", "safari"}

	// Initialize database connection
	db, err := config.InitDB()
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return
	}

	var site models.Site
	if err := db.First(&site, siteID).Error; err != nil {
		log.Printf("Site not found")
		return
	}

	var device models.Device
	if err := db.First(&device, deviceID).Error; err != nil {
		log.Printf("Device not found")
		return
	}

	var feature models.Feature
	if err := db.First(&feature, featureID).Error; err != nil {
		log.Printf("Feature not found")
		return
	}

	for _, browser := range browsers {
		go func(browserType string) {
			startTime := time.Now()

			// Create initial result record
			result := models.Result{
				SiteID:    siteID,
				DeviceID:  deviceID,
				FeatureID: featureID,
				Browser:   browserType,
				Status:    "processing",
			}

			if err := db.Create(&result).Error; err != nil {
				log.Printf("Failed to create result for %s: %v", browserType, err)
				return
			}

			// Create a new BrowserStack runner
			runner := NewBrowserStackRunner()

			// Initialize the runner with specified browser
			if err := runner.Initialize(browserType); err != nil {
				logError(result.ID, fmt.Sprintf("Failed to initialize %s runner: %v", browserType, err))
				return
			}
			defer runner.Close()

			// Create test result
			testResult := &TestResult{
				Feature:   feature.Name,
				Site:      site.Name,
				Browser:   browserType,
				Device:    device.Name,
				Status:    "processing",
			}

			// Log test start
			if err := runner.LogTestStep(fmt.Sprintf("Test started for %s - Initializing browser", browserType)); err != nil {
				log.Printf("Warning: Failed to log test start for %s: %v", browserType, err)
			}

			// Take screenshot before login
			beforeLoginScreenshot, err := runner.TakeScreenshot()
			if err != nil {
				log.Printf("Warning: Failed to take before login screenshot for %s: %v", browserType, err)
			} else {
				log.Printf("Before login screenshot saved for %s: %s", browserType, beforeLoginScreenshot)
				if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken before login: %s", beforeLoginScreenshot)); err != nil {
					log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
				}
				// Store screenshot in result details
				// resultDetail := models.ResultDetail{
				// 	ResultID:   result.ID,
				// 	Screenshot: beforeLoginScreenshot,
				// 	Description: "Screenshot before login",
				// }
				// if err := db.Create(&resultDetail).Error; err != nil {
				// 	log.Printf("Warning: Failed to store before login screenshot for %s: %v", browserType, err)
				// }
			}

			// Navigate to login page
			log.Printf("Navigating to login page using %s...", browserType)
			if err := runner.LogTestStep(fmt.Sprintf("Navigating to login page using %s", browserType)); err != nil {
				log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
			}

			if err := runner.NavigateToLoginPage(site.Name); err != nil {
				logMsg := fmt.Sprintf("Failed to navigate to login page using %s: %v", browserType, err)
				logError(result.ID, logMsg)
				if err := runner.LogTestStep(logMsg); err != nil {
					log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
				}
				testResult.Status = "failed"
				testResult.ErrorLog = logMsg
				testResult.Screenshot = beforeLoginScreenshot
				if err := runner.StoreTestResult(testResult); err != nil {
					log.Printf("Warning: Failed to store test result for %s: %v", browserType, err)
				}
				return
			}

			if err := runner.LogTestStep(fmt.Sprintf("Successfully navigated to login page using %s", browserType)); err != nil {
				log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
			}
			log.Printf("Successfully navigated to login page using %s!", browserType)

			// Take screenshot of login page
			loginPageScreenshot, err := runner.TakeScreenshot()
			if err != nil {
				log.Printf("Warning: Failed to take login page screenshot for %s: %v", browserType, err)
			} else {
				log.Printf("Login page screenshot saved for %s: %s", browserType, loginPageScreenshot)
				if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken of login page: %s", loginPageScreenshot)); err != nil {
					log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
				}
				// Store screenshot in result details
				resultDetail := models.ResultDetail{
					ResultID:   result.ID,
					Screenshot: loginPageScreenshot,
					Description: "Screenshot of login page after successful navigation",
				}
				if err := db.Create(&resultDetail).Error; err != nil {
					log.Printf("Warning: Failed to store login page screenshot for %s: %v", browserType, err)
				}
			}

			// Perform login
			log.Printf("Attempting to login to " + site.Name + " using %s...", browserType)
			if err := runner.LogTestStep(fmt.Sprintf("Attempting to login to " + site.Name + " using %s", browserType)); err != nil {
				log.Printf("Warning: Failed to log login attempt for %s: %v", browserType, err)
			}

			if err := runner.LoginToSentiLive(site.Name, email, password); err != nil {
				logMsg := fmt.Sprintf("Login failed for %s: %v", browserType, err)
				logError(result.ID, logMsg)
				if err := runner.LogTestStep(logMsg); err != nil {
					log.Printf("Warning: Failed to log login failure for %s: %v", browserType, err)
				}
				testResult.Status = "failed"
				testResult.ErrorLog = logMsg
				testResult.Screenshot = beforeLoginScreenshot
				if err := runner.StoreTestResult(testResult); err != nil {
					log.Printf("Warning: Failed to store test result for %s: %v", browserType, err)
				}
				return
			}
			
			if err := runner.LogTestStep(fmt.Sprintf("Login successful for %s", browserType)); err != nil {
				log.Printf("Warning: Failed to log successful login for %s: %v", browserType, err)
			}
			log.Printf("Login successful for %s!", browserType)

			// Take screenshot after login
			afterLoginScreenshot, err := runner.TakeScreenshot()
			if err != nil {
				log.Printf("Warning: Failed to take after login screenshot for %s: %v", browserType, err)
			} else {
				log.Printf("After login screenshot saved for %s: %s", browserType, afterLoginScreenshot)
				if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken after login: %s", afterLoginScreenshot)); err != nil {
					log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
				}
				// Store screenshot in result details
				resultDetail := models.ResultDetail{
					ResultID:   result.ID,
					Screenshot: afterLoginScreenshot,
					Description: "Screenshot after successful login",
				}
				if err := db.Create(&resultDetail).Error; err != nil {
					log.Printf("Warning: Failed to store after login screenshot for %s: %v", browserType, err)
				}
			}

			// Test Chat Functionality
			if feature.Name == "Chat Functionality" {
				// Navigate to chat page
				log.Printf("Navigating to chat page using %s...", browserType)
				if err := runner.LogTestStep(fmt.Sprintf("Navigating to chat page using %s", browserType)); err != nil {
					log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
				}

				if err := runner.NavigateToChatPage(site.Name); err != nil {
					logMsg := fmt.Sprintf("Failed to navigate to chat page using %s: %v", browserType, err)
					logError(result.ID, logMsg)
					if err := runner.LogTestStep(logMsg); err != nil {
						log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
					}
					testResult.Status = "failed"
					testResult.ErrorLog = logMsg
					testResult.Screenshot = afterLoginScreenshot
					if err := runner.StoreTestResult(testResult); err != nil {
						log.Printf("Warning: Failed to store test result for %s: %v", browserType, err)
					}
					return
				}

				if err := runner.LogTestStep(fmt.Sprintf("Successfully navigated to chat page using %s", browserType)); err != nil {
					log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
				}
				log.Printf("Successfully navigated to chat page using %s!", browserType)

				// Take screenshot of chat page
				chatPageScreenshot, err := runner.TakeScreenshot()
				if err != nil {
					log.Printf("Warning: Failed to take chat page screenshot for %s: %v", browserType, err)
				} else {
					log.Printf("Chat page screenshot saved for %s: %s", browserType, chatPageScreenshot)
					if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken of chat page: %s", chatPageScreenshot)); err != nil {
						log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
					}
					// Store screenshot in result details
					resultDetail := models.ResultDetail{
						ResultID:   result.ID,
						Screenshot: chatPageScreenshot,
						Description: "Screenshot of chat page after successful navigation",
					}
					if err := db.Create(&resultDetail).Error; err != nil {
						log.Printf("Warning: Failed to store chat page screenshot for %s: %v", browserType, err)
					}
				}

				// Navigate to open chat
				log.Printf("Navigating to open chat using %s...", browserType)
				if err := runner.LogTestStep(fmt.Sprintf("Navigating to open chat using %s", browserType)); err != nil {
					log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
				}

				if err := runner.NavigateToOpenChat(site.Name); err != nil {
					logMsg := fmt.Sprintf("Failed to navigate to open chat using %s: %v", browserType, err)
					logError(result.ID, logMsg)
					if err := runner.LogTestStep(logMsg); err != nil {
						log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
					}
					testResult.Status = "failed"
					testResult.ErrorLog = logMsg
					testResult.Screenshot = chatPageScreenshot
					if err := runner.StoreTestResult(testResult); err != nil {
						log.Printf("Warning: Failed to store test result for %s: %v", browserType, err)
					}
					return
				}

				if err := runner.LogTestStep(fmt.Sprintf("Successfully navigated to open chat using %s", browserType)); err != nil {
					log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
				}
				log.Printf("Successfully navigated to open chat using %s!", browserType)

				// Take screenshot of open chat
				openChatPageScreenshot, err := runner.TakeScreenshot()
				if err != nil {
					log.Printf("Warning: Failed to take open chat screenshot for %s: %v", browserType, err)
				} else {
					log.Printf("Open chat screenshot saved for %s: %s", browserType, openChatPageScreenshot)
					if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken of open chat: %s", openChatPageScreenshot)); err != nil {
						log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
					}
					// Store screenshot in result details
					resultDetail := models.ResultDetail{
						ResultID:   result.ID,
						Screenshot: openChatPageScreenshot,
						Description: "Screenshot of open chat after successful navigation",
					}
					if err := db.Create(&resultDetail).Error; err != nil {
						log.Printf("Warning: Failed to store open chat screenshot for %s: %v", browserType, err)
					}
				}

				// Navigate to send message to chat
				log.Printf("Navigating to send message to chat using %s...", browserType)
				if err := runner.LogTestStep(fmt.Sprintf("Navigating to send message to chat using %s", browserType)); err != nil {
					log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
				}

				if err := runner.SendingMessageToChat(site.Name); err != nil {
					logMsg := fmt.Sprintf("Failed to navigate to send message to chat using %s: %v", browserType, err)
					logError(result.ID, logMsg)
					if err := runner.LogTestStep(logMsg); err != nil {
						log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
					}
					testResult.Status = "failed"
					testResult.ErrorLog = logMsg
					testResult.Screenshot = chatPageScreenshot
					if err := runner.StoreTestResult(testResult); err != nil {
						log.Printf("Warning: Failed to store test result for %s: %v", browserType, err)
					}
					return
				}

				if err := runner.LogTestStep(fmt.Sprintf("Successfully navigated to send message to chat using %s", browserType)); err != nil {
					log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
				}
				log.Printf("Successfully navigated to send message to chat using %s!", browserType)

				// Take screenshot of sending message to chat
				sendChatPageScreenshot, err := runner.TakeScreenshot()
				if err != nil {
					log.Printf("Warning: Failed to take send message to chat screenshot for %s: %v", browserType, err)
				} else {
					log.Printf("Send message to chat screenshot saved for %s: %s", browserType, sendChatPageScreenshot)
					if err := runner.LogTestStep(fmt.Sprintf("Screenshot taken of sending message to chat: %s", sendChatPageScreenshot)); err != nil {
						log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
					}
					// Store screenshot in result details
					resultDetail := models.ResultDetail{
						ResultID:   result.ID,
						Screenshot: sendChatPageScreenshot,
						Description: "Screenshot of sending message to chat after successful navigation",
					}
					if err := db.Create(&resultDetail).Error; err != nil {
						log.Printf("Warning: Failed to store send message to chat screenshot for %s: %v", browserType, err)
					}
				}
			} else {
				// the rest function has not done yet
				logMsg := fmt.Sprintf("%s feature has not been implemented yet", feature.Name)
				logError(result.ID, logMsg)
				if err := runner.LogTestStep(logMsg); err != nil {
					log.Printf("%s feature has not been implemented yet", feature.Name)
				}
				testResult.Status = "failed"
				testResult.ErrorLog = logMsg
				if err := runner.StoreTestResult(testResult); err != nil {
					log.Printf("Warning: %s feature has not been implemented yet", feature.Name)
				}
				return
			}

			// Keep the session alive for a while to verify everything is working
			log.Printf("Keeping %s session alive for 10 seconds...", browserType)
			if err := runner.LogTestStep(fmt.Sprintf("Keeping %s session alive for 10 seconds", browserType)); err != nil {
				log.Printf("Warning: Failed to log session wait for %s: %v", browserType, err)
			}
			time.Sleep(5 * time.Second)

			// Calculate duration
			duration := time.Since(startTime)

			// Update result status to passed
			if err := db.Model(&result).Updates(map[string]interface{}{
				"status": "passed",
				"duration": duration.Seconds(),
			}).Error; err != nil {
				log.Printf("Warning: Failed to update result status for %s: %v", browserType, err)
			}

			if err := runner.LogTestStep(fmt.Sprintf("Test completed successfully for %s in %v", browserType, duration)); err != nil {
				log.Printf("Warning: Failed to log test completion for %s: %v", browserType, err)
			}
			log.Printf("Test completed successfully for %s in %v!", browserType, duration)
		}(browser)
	}
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