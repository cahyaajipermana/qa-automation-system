package testrunner

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

		// Maximize browser window
		if err := driver.MaximizeWindow(""); err != nil {
			return fmt.Errorf("failed to maximize window: %v", err)
		}

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

	log.Printf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), step)

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

// RunTestInBackground runs the test in the background for multiple browsers
func RunTestInBackground(siteID, deviceID, featureID uint, email, password string) {
	// Define browsers to test
	browsers := []string{"chrome"/*, "firefox", "edge", "safari"*/}

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
				runner.logError(result.ID, time.Since(startTime), fmt.Sprintf("Failed to initialize %s runner: %v", browserType, err))
				return
			}
			defer runner.Close()

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
			}

			// Navigate to login page
			log.Printf("Navigating to login page using %s...", browserType)
			if err := runner.LogTestStep(fmt.Sprintf("Navigating to login page using %s", browserType)); err != nil {
				log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
			}

			if err := runner.NavigateToLoginPage(site.Name); err != nil {
				logMsg := fmt.Sprintf("Failed to navigate to login page using %s: %v", browserType, err)
				runner.logError(result.ID, time.Since(startTime), logMsg)
				if err := runner.LogTestStep(logMsg); err != nil {
					log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
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
					Description: "Screenshot of login page",
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

			if err := runner.LoginHandler(site.Name, email, password); err != nil {
				logMsg := fmt.Sprintf("Login failed for %s: %v", browserType, err)
				runner.logError(result.ID, time.Since(startTime), logMsg)
				if err := runner.LogTestStep(logMsg); err != nil {
					log.Printf("Warning: Failed to log login failure for %s: %v", browserType, err)
				}
				return
			}
			
			if err := runner.LogTestStep(fmt.Sprintf("Login successful for %s", browserType)); err != nil {
				log.Printf("Warning: Failed to log successful login for %s: %v", browserType, err)
			}
			log.Printf("Login successful for %s!", browserType)

			// Take screenshot after login
			// Home page screenshot
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
				if err := runner.ChatFunctionality(db, site, device, feature, browserType, result.ID, startTime); err != nil {
					log.Printf("Warning: Failed to test chat functionality for Result ID %d: %v", result.ID, err)
					return
				}
			} else if feature.Name == "Scrolling Home Page" {
				if err := runner.ScrollingHomePage(db, site, device, feature, browserType, result.ID, startTime); err != nil {
					log.Printf("Warning: Failed to test scrolling home page for Result ID %d: %v", result.ID, err)
					return
				}
			} else {
				// the rest function has not done yet
				logMsg := fmt.Sprintf("%s feature has not been implemented yet", feature.Name)
				runner.logError(result.ID, time.Since(startTime), logMsg)
				if err := runner.LogTestStep(logMsg); err != nil {
					log.Printf("%s feature has not been implemented yet", feature.Name)
				}
				return
			}

			// Keep the session alive for a while to verify everything is working
			log.Printf("Keeping %s session alive for 5 seconds...", browserType)
			if err := runner.LogTestStep(fmt.Sprintf("Keeping %s session alive for 5 seconds", browserType)); err != nil {
				log.Printf("Warning: Failed to log session wait for %s: %v", browserType, err)
			}
			time.Sleep(5 * time.Second)

			// Calculate duration
			duration := time.Since(startTime)

			// savedVideoPath, err := runner.SaveVideo(result.ID)
			// if err != nil || savedVideoPath == "" {
			// 	log.Printf("Warning: Failed to save video for Result ID %d: %v", resultID, err)
			// }

			// Update result status to passed
			if err := db.Model(&result).Updates(map[string]interface{}{
				"status": "passed",
				"duration": duration.Seconds(),
				// "video_path": savedVideoPath,
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

// LoginHandler performs login to site
func (r *BrowserStackRunner) LoginHandler(siteName string, email, password string) error {
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
		return fmt.Errorf("failed to find .login-text button: %v", err)
	}
	if err := loginButton.Click(); err != nil {
		return fmt.Errorf("failed to click .login-text button: %v", err)
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

// NavigateToHomePage navigates to the chat page
func (r *BrowserStackRunner) NavigateToHomePage(siteName string) error {
	if r.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	// Navigate to home page
	if err := r.driver.Get("https://" + siteName); err != nil {
		return fmt.Errorf("failed to navigate to home page: %v", err)
	}

	// Wait for home page to load and be ready
	time.Sleep(3 * time.Second)

	// Verify we're on the home page
	currentURL, err := r.driver.CurrentURL()
	if err != nil {
		return fmt.Errorf("failed to get current URL: %v", err)
	}
	if currentURL != "https://" + siteName {
		return fmt.Errorf("navigation failed: not on home page, current URL: %s", currentURL)
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

	chatRestID := ""

	if siteName == "senti.live" {
		chatRestID = os.Getenv("SENTI_CHAT_REST_ID")
	}

	if siteName == "shorts.senti.live" {
		chatRestID = os.Getenv("SHORTS_SENTI_CHAT_REST_ID")
	}

	if siteName == "hothinge.com" {
		chatRestID = os.Getenv("HOTHINGE_CHAT_REST_ID")
	}

	if siteName == "viblys.com" {
		chatRestID = os.Getenv("VIBLYS_CHAT_REST_ID")
	}

	if chatRestID == "" {
		return fmt.Errorf("chat rest ID not found for site: %s", siteName)
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

	chatRestID := ""

	if siteName == "senti.live" {
		chatRestID = os.Getenv("SENTI_CHAT_REST_ID")
	}

	if siteName == "shorts.senti.live" {
		chatRestID = os.Getenv("SHORTS_SENTI_CHAT_REST_ID")
	}

	if siteName == "hothinge.com" {
		chatRestID = os.Getenv("HOTHINGE_CHAT_REST_ID")
	}

	if siteName == "viblys.com" {
		chatRestID = os.Getenv("VIBLYS_CHAT_REST_ID")
	}

	if chatRestID == "" {
		return fmt.Errorf("chat rest ID not found for site: %s", siteName)
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

func (r *BrowserStackRunner) ChatFunctionality(db *gorm.DB, site models.Site, device models.Device, feature models.Feature, browserType string, resultID uint, startTime time.Time) error {
	// Navigate to chat page
	log.Printf("Navigating to chat page using %s...", browserType)
	if err := r.LogTestStep(fmt.Sprintf("Navigating to chat page using %s", browserType)); err != nil {
		log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
	}

	if err := r.NavigateToChatPage(site.Name); err != nil {
		logMsg := fmt.Sprintf("Failed to navigate to chat page using %s: %v", browserType, err)
		r.logError(resultID, time.Since(startTime), logMsg)
		if err := r.LogTestStep(logMsg); err != nil {
			log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
		}
		return err
	}

	if err := r.LogTestStep(fmt.Sprintf("Successfully navigated to chat page using %s", browserType)); err != nil {
		log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
	}
	log.Printf("Successfully navigated to chat page using %s!", browserType)

	// Take screenshot of chat page
	chatPageScreenshot, err := r.TakeScreenshot()
	if err != nil {
		log.Printf("Warning: Failed to take chat page screenshot for %s: %v", browserType, err)
	} else {
		log.Printf("Chat page screenshot saved for %s: %s", browserType, chatPageScreenshot)
		if err := r.LogTestStep(fmt.Sprintf("Screenshot taken of chat page: %s", chatPageScreenshot)); err != nil {
			log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
		}
		// Store screenshot in result details
		resultDetail := models.ResultDetail{
			ResultID:   resultID,
			Screenshot: chatPageScreenshot,
			Description: "Screenshot of chat page",
		}
		if err := db.Create(&resultDetail).Error; err != nil {
			log.Printf("Warning: Failed to store chat page screenshot for %s: %v", browserType, err)
		}
	}

	// Navigate to open chat
	log.Printf("Navigating to open chat using %s...", browserType)
	if err := r.LogTestStep(fmt.Sprintf("Navigating to open chat using %s", browserType)); err != nil {
		log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
	}

	if err := r.NavigateToOpenChat(site.Name); err != nil {
		logMsg := fmt.Sprintf("Failed to navigate to open chat using %s: %v", browserType, err)
		r.logError(resultID, time.Since(startTime), logMsg)
		if err := r.LogTestStep(logMsg); err != nil {
			log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
		}
		return err
	}

	if err := r.LogTestStep(fmt.Sprintf("Successfully navigated to open chat using %s", browserType)); err != nil {
		log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
	}
	log.Printf("Successfully navigated to open chat using %s!", browserType)

	// Take screenshot of open chat
	openChatPageScreenshot, err := r.TakeScreenshot()
	if err != nil {
		log.Printf("Warning: Failed to take open chat screenshot for %s: %v", browserType, err)
	} else {
		log.Printf("Open chat screenshot saved for %s: %s", browserType, openChatPageScreenshot)
		if err := r.LogTestStep(fmt.Sprintf("Screenshot taken of open chat: %s", openChatPageScreenshot)); err != nil {
			log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
		}
		// Store screenshot in result details
		resultDetail := models.ResultDetail{
			ResultID:   resultID,
			Screenshot: openChatPageScreenshot,
			Description: "Screenshot of open chat",
		}
		if err := db.Create(&resultDetail).Error; err != nil {
			log.Printf("Warning: Failed to store open chat screenshot for %s: %v", browserType, err)
		}
	}

	// Navigate to send message to chat
	log.Printf("Navigating to send message to chat using %s...", browserType)
	if err := r.LogTestStep(fmt.Sprintf("Navigating to send message to chat using %s", browserType)); err != nil {
		log.Printf("Warning: Failed to log navigation attempt for %s: %v", browserType, err)
	}

	if err := r.SendingMessageToChat(site.Name); err != nil {
		logMsg := fmt.Sprintf("Failed to navigate to send message to chat using %s: %v", browserType, err)
		r.logError(resultID, time.Since(startTime), logMsg)
		if err := r.LogTestStep(logMsg); err != nil {
			log.Printf("Warning: Failed to log navigation failure for %s: %v", browserType, err)
		}
		return err
	}

	if err := r.LogTestStep(fmt.Sprintf("Successfully navigated to send message to chat using %s", browserType)); err != nil {
		log.Printf("Warning: Failed to log successful navigation for %s: %v", browserType, err)
	}
	log.Printf("Successfully navigated to send message to chat using %s!", browserType)

	// Take screenshot of sending message to chat
	sendChatPageScreenshot, err := r.TakeScreenshot()
	if err != nil {
		log.Printf("Warning: Failed to take send message to chat screenshot for %s: %v", browserType, err)
	} else {
		log.Printf("Send message to chat screenshot saved for %s: %s", browserType, sendChatPageScreenshot)
		if err := r.LogTestStep(fmt.Sprintf("Screenshot taken of sending message to chat: %s", sendChatPageScreenshot)); err != nil {
			log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
		}
		// Store screenshot in result details
		resultDetail := models.ResultDetail{
			ResultID:   resultID,
			Screenshot: sendChatPageScreenshot,
			Description: "Screenshot of sending message to chat",
		}
		if err := db.Create(&resultDetail).Error; err != nil {
			log.Printf("Warning: Failed to store send message to chat screenshot for %s: %v", browserType, err)
		}
	}

	return nil
}

func (r *BrowserStackRunner) ScrollingHomePage(db *gorm.DB, site models.Site, device models.Device, feature models.Feature, browserType string, resultID uint, startTime time.Time) error {
	wheelCssSelector := ""

	if site.Name == "senti.live" {
		wheelCssSelector = "root-observed"
	}

	if site.Name == "hothinge.com" {
		wheelCssSelector = "layout-page"
	}

	// If current site is shorts.senti.live or viblys.com
	// Check the Pause and Play Video action
	if site.Name == "shorts.senti.live" || site.Name == "viblys.com" {
		wheelCssSelector = "video-feed"
		// After login, by default it will redirect to home page
		// And the video will automatically play
		// So we need to click on the Video to pause the video
		// Click the Video Element
		r.pauseVideo(db, resultID, browserType, startTime)

		// And then click on the Play Video button to play the video
		// Click the Play Video Button
		r.playVideo(db, resultID, browserType, startTime)

		// Simulate wheel event with deltaY of 150
		wheelScript := simulateWheelEvent(150, wheelCssSelector)
		if _, err := r.driver.ExecuteScript(wheelScript, nil); err != nil {
			r.logError(resultID, time.Since(startTime), fmt.Sprintf("Failed to simulate scroll event: %v", err))
			return err
		}
	} else {
		// Simulate scroll event with top 1000
		scrollScript := simulateScrollEvent(site.Name, 1000, wheelCssSelector)
		if _, err := r.driver.ExecuteScript(scrollScript, nil); err != nil {
			r.logError(resultID, time.Since(startTime), fmt.Sprintf("Failed to simulate scroll event: %v", err))
			return err
		}
	}

	// Wait for scroll to complete
	time.Sleep(1 * time.Second)

	// Take screenshot after scroll event
	afterScrollScreenshot, err := r.TakeScreenshot()
	if err != nil {
		log.Printf("Warning: Failed to take after-scroll screenshot for %s: %v", browserType, err)
	} else {
		log.Printf("After-scroll screenshot saved for %s: %s", browserType, afterScrollScreenshot)
		if err := r.LogTestStep(fmt.Sprintf("Screenshot taken after scroll event: %s", afterScrollScreenshot)); err != nil {
			log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
		}
		// Store screenshot in result details
		resultDetail := models.ResultDetail{
			ResultID:    resultID,
			Screenshot:  afterScrollScreenshot,
			Description: "Screenshot after scroll event",
		}
		if err := db.Create(&resultDetail).Error; err != nil {
			log.Printf("Warning: Failed to store after-scroll screenshot for %s: %v", browserType, err)
		}
	}

	return nil
}

func (r *BrowserStackRunner) pauseVideo(db *gorm.DB, resultID uint, browserType string, startTime time.Time) error {
	videoElement, err := r.driver.FindElement(selenium.ByCSSSelector, ".video-player")
	if err != nil {
		return fmt.Errorf("failed to find .video-player element: %v", err)
	}
	if err := videoElement.Click(); err != nil {
		return fmt.Errorf("failed to click .video-player element: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Take screenshot of paused video
	pausedVideoScreenshot, err := r.TakeScreenshot()
	if err != nil {
		log.Printf("Warning: Failed to take paused video screenshot for %s: %v", browserType, err)
	} else {
		log.Printf("Paused video screenshot saved for %s: %s", browserType, pausedVideoScreenshot)
		if err := r.LogTestStep(fmt.Sprintf("Screenshot taken of paused video: %s", pausedVideoScreenshot)); err != nil {
			log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
		}
		// Store screenshot in result details
		resultDetail := models.ResultDetail{
			ResultID:    resultID,
			Screenshot:  pausedVideoScreenshot,
			Description: "Screenshot of paused video",
		}
		if err := db.Create(&resultDetail).Error; err != nil {
			log.Printf("Warning: Failed to store paused video screenshot for %s: %v", browserType, err)
		}
	}

	return nil
}

func (r *BrowserStackRunner) playVideo(db *gorm.DB, resultID uint, browserType string, startTime time.Time) error {
	playVideoButton, err := r.driver.FindElement(selenium.ByCSSSelector, ".play-button-overlay")
	if err != nil {
		return fmt.Errorf("failed to find .play-button-overlay button: %v", err)
	}
	if err := playVideoButton.Click(); err != nil {
		return fmt.Errorf("failed to click .play-button-overlay button: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Take screenshot of played video
	playedVideoScreenshot, err := r.TakeScreenshot()
	if err != nil {
		log.Printf("Warning: Failed to take played video screenshot for %s: %v", browserType, err)
	} else {
		log.Printf("Played video screenshot saved for %s: %s", browserType, playedVideoScreenshot)
		if err := r.LogTestStep(fmt.Sprintf("Screenshot taken of played video: %s", playedVideoScreenshot)); err != nil {
			log.Printf("Warning: Failed to log screenshot for %s: %v", browserType, err)
		}
		// Store screenshot in result details
		resultDetail := models.ResultDetail{
			ResultID:    resultID,
			Screenshot:  playedVideoScreenshot,
			Description: "Screenshot of played video",
		}
		if err := db.Create(&resultDetail).Error; err != nil {
			log.Printf("Warning: Failed to store played video screenshot for %s: %v", browserType, err)
		}
	}
	
	return nil
}

func simulateScrollEvent(siteName string, deltaY int, wheelCssSelector string) string {
	if siteName == "senti.live" {
		return fmt.Sprintf(`
			document.getElementsByClassName('%s')[0].scrollTo({
				top: %d,
				behavior: 'smooth'
			});
		`, wheelCssSelector, deltaY)
	}

	// default scroll by browser window
	return fmt.Sprintf(`
		window.scrollTo({
			top: %d,
			behavior: 'smooth'
		});
	`, deltaY)
}

func simulateWheelEvent(deltaY int, wheelCssSelector string) string {
	return fmt.Sprintf(`
		let wheelEvent = new WheelEvent('wheel', {
			deltaY: %d,
			deltaMode: 1
		});
		document.getElementsByClassName('%s')[0].dispatchEvent(wheelEvent);	
	`, deltaY, wheelCssSelector)
}

// logError updates the result status and error log in the database
func (r *BrowserStackRunner) logError(resultID uint, duration time.Duration, errorMsg string) {
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

	// savedVideoPath, err := r.SaveVideo(resultID)
	// if err != nil || savedVideoPath == "" {
	// 	log.Printf("Warning: Failed to save video for Result ID %d: %v", resultID, err)
	// }

	result.Status = "failed"
	result.Duration = duration.Seconds()
	// result.VideoPath = savedVideoPath
	result.ErrorLog = errorMsg
	if err := db.Save(&result).Error; err != nil {
		log.Printf("Failed to update result with error: %v", err)
	}
}

// Save Video to Videos Folder
func (r *BrowserStackRunner) SaveVideo(resultID uint) (videoPath string, err error) {
	// Get session ID for video download
	sessionID := r.driver.SessionID()
	if sessionID == "" {
		return "", fmt.Errorf("failed to get session ID: %v", err)
	}

	log.Printf("BrowserStack Session ID: %s", sessionID)

	// Create videos directory if it doesn't exist
	videosDir := "videos"
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create videos directory: %v", err)
	}

	// Download video from BrowserStack
	videoURL := fmt.Sprintf("https://api.browserstack.com/app-automate/sessions/%s/video", sessionID)
	videoPath = filepath.Join(videosDir, fmt.Sprintf("test_%d_%s.mp4", resultID, time.Now().Format("20060102_150405")))

	// Download video file
	resp, err := http.Get(videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download video: status code %d", resp.StatusCode)
	}

	// Create video file
	videoFile, err := os.Create(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to create video file: %v", err)
	}
	defer videoFile.Close()

	// Copy video content to file
	if _, err := io.Copy(videoFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save video: %v", err)
	}

	return videoPath, nil
} 