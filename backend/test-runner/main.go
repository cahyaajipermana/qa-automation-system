package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// Get credentials from environment variables
	email := os.Getenv("SENTI_EMAIL")
	password := os.Getenv("SENTI_PASSWORD")

	if email == "" || password == "" {
		log.Fatal("SENTI_EMAIL and SENTI_PASSWORD environment variables must be set")
	}

	// Create a new BrowserStack runner
	runner := NewBrowserStackRunner()

	// Initialize the runner with Chrome browser
	if err := runner.Initialize("chrome"); err != nil {
		log.Fatalf("Failed to initialize runner: %v", err)
	}
	defer runner.Close()

	// Create test result
	testResult := &TestResult{
		Feature:   "Login",
		Site:      "senti.live",
		Browser:   "chrome",
		Device:    "desktop",
		Status:    "processing",
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
		log.Fatal(logMsg)
		if err := runner.LogTestStep(logMsg); err != nil {
			log.Printf("Warning: Failed to log login failure: %v", err)
		}
		testResult.Status = "failed"
		testResult.ErrorLog = logMsg
		testResult.Screenshot = beforeLoginScreenshot
		if err := runner.StoreTestResult(testResult); err != nil {
			log.Printf("Warning: Failed to store test result: %v", err)
		}
		os.Exit(1)
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
		log.Fatal(logMsg)
		if err := runner.LogTestStep(logMsg); err != nil {
			log.Printf("Warning: Failed to log navigation failure: %v", err)
		}
		testResult.Status = "failed"
		testResult.ErrorLog = logMsg
		testResult.Screenshot = afterLoginScreenshot
		if err := runner.StoreTestResult(testResult); err != nil {
			log.Printf("Warning: Failed to store test result: %v", err)
		}
		os.Exit(1)
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
} 