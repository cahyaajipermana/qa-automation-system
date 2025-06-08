package testrunner

import (
	"fmt"
	"time"
	"github.com/tebeka/selenium"
	"qa-automation-system/backend/config"
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
}

// NewBrowserStackRunner creates a new BrowserStack runner
func NewBrowserStackRunner() *BrowserStackRunner {
	return &BrowserStackRunner{
		config:      config.NewBrowserStackConfig(),
		testResults: make([]TestResult, 0),
	}
}

// Initialize sets up the browser for testing
func (r *BrowserStackRunner) Initialize(browserType string) error {
	capabilities := make(map[string]interface{})

	// cahyaajipermana_SgK6Bo
	// vH3hWWMrHTXfH4GppTUX
	
	// Set BrowserStack credentials
	capabilities["browserstack.user"] = "cahyaajipermana_SgK6Bo"
	capabilities["browserstack.key"] = "vH3hWWMrHTXfH4GppTUX"

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

// InitializeDevice sets up a mobile device for testing
func (r *BrowserStackRunner) InitializeDevice(deviceType string) error {
	capabilities := make(map[string]interface{})
	
	// Set BrowserStack credentials
	capabilities["browserstack.user"] = "cahyaajipermana_SgK6Bo"
	capabilities["browserstack.key"] = "vH3hWWMrHTXfH4GppTUX"

	// Set base capabilities
	for k, v := range r.config.Capabilities {
		capabilities[k] = v
	}

	// Set device-specific capabilities
	if deviceCapabilities, ok := r.config.Devices[deviceType]; ok {
		for k, v := range deviceCapabilities {
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

// RunTest executes a test case
func (r *BrowserStackRunner) RunTest(testCase TestCase, browserType string) (*TestResult, error) {
	result := &TestResult{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Feature:   testCase.Feature,
		Site:      testCase.Site,
		Browser:   browserType,
		Timestamp: time.Now(),
		Status:    "pass",
	}

	// Initialize browser
	if err := r.Initialize(browserType); err != nil {
		result.Status = "fail"
		result.ErrorLog = fmt.Sprintf("Failed to initialize browser: %v", err)
		return result, err
	}
	defer r.wd.Quit()

	// Navigate to the test URL
	if err := r.wd.Get(testCase.URL); err != nil {
		result.Status = "fail"
		result.ErrorLog = fmt.Sprintf("Failed to navigate to URL: %v", err)
		return result, err
	}

	// Execute test steps
	for _, step := range testCase.Steps {
		if err := r.executeStep(step); err != nil {
			result.Status = "fail"
			result.ErrorLog = fmt.Sprintf("Step failed: %v", err)
			return result, err
		}
	}

	// Take screenshot
	screenshot, err := r.wd.Screenshot()
	if err == nil {
		result.Screenshot = fmt.Sprintf("data:image/png;base64,%s", screenshot)
	}

	return result, nil
}

// executeStep executes a single test step
func (r *BrowserStackRunner) executeStep(step TestStep) error {
	switch step.Action {
	case "navigate":
		return r.wd.Get(step.Value)
	case "click_text":
		// Try different methods to find and click the element
		element, err := r.findElementByText(step.Value)
		if err != nil {
			return fmt.Errorf("failed to find element with text '%s': %v", step.Value, err)
		}
		return element.Click()
	case "click":
		elem, err := r.wd.FindElement(selenium.ByCSSSelector, step.Selector)
		if err != nil {
			return fmt.Errorf("failed to find element: %v", err)
		}
		return elem.Click()
	case "input":
		elem, err := r.wd.FindElement(selenium.ByCSSSelector, step.Selector)
		if err != nil {
			return fmt.Errorf("failed to find element: %v", err)
		}
		return elem.SendKeys(step.Value)
	case "wait":
		time.Sleep(2 * time.Second)
	case "assert":
		elem, err := r.wd.FindElement(selenium.ByCSSSelector, step.Selector)
		if err != nil {
			return fmt.Errorf("failed to find element: %v", err)
		}
		text, err := elem.Text()
		if err != nil {
			return fmt.Errorf("failed to get element text: %v", err)
		}
		if text != step.Value {
			return fmt.Errorf("assertion failed: expected %s, got %s", step.Value, text)
		}
	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}
	return nil
}

// getBrowserName returns the current browser name
func (r *BrowserStackRunner) getBrowserName() string {
	if r.wd == nil {
		return "unknown"
	}
	caps, err := r.wd.Capabilities()
	if err != nil {
		return "unknown"
	}
	if browser, ok := caps["browserName"].(string); ok {
		return browser
	}
	return "unknown"
}

// getDeviceName returns the current device name
func (r *BrowserStackRunner) getDeviceName() string {
	if r.wd == nil {
		return "desktop"
	}
	caps, err := r.wd.Capabilities()
	if err != nil {
		return "desktop"
	}
	if bstackOpts, ok := caps["bstack:options"].(map[string]interface{}); ok {
		if device, ok := bstackOpts["deviceName"].(string); ok {
			return device
		}
	}
	return "desktop"
}

// Cleanup closes the browser session
func (r *BrowserStackRunner) Cleanup() error {
	if r.wd != nil {
		return r.wd.Quit()
	}
	return nil
}

// GetTestResults returns all test results
func (r *BrowserStackRunner) GetTestResults() []TestResult {
	return r.testResults
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