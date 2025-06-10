package controllers

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"qa-automation-system/backend/config"
	"qa-automation-system/backend/models"
	"qa-automation-system/backend/pkg/testrunner"
)

var db *gorm.DB

// Initialize the database connection
func init() {
	var err error
	db, err = config.InitDB()
	if err != nil {
		panic("Failed to initialize database: " + err.Error())
	}
}

// ResultController handles result-related operations
type ResultController struct {
	DB *gorm.DB
}

// NewResultController creates a new result controller
func NewResultController(db *gorm.DB) *ResultController {
	return &ResultController{DB: db}
}

// Create handles the creation of a new result
func (c *ResultController) Create(ctx *gin.Context) {
	var payload struct {
		SiteID    uint `json:"site_id" binding:"required"`
		DeviceID  uint `json:"device_id" binding:"required"`
		FeatureID uint `json:"feature_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get credentials from environment variables
	email := os.Getenv("SENTI_EMAIL")
	password := os.Getenv("SENTI_PASSWORD")

	// Run the test in the background with the payload
	go testrunner.RunTestInBackground(payload.SiteID, payload.DeviceID, payload.FeatureID, email, password)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Test started in background",
		"payload": payload,
	})
}

// GetAll retrieves all results
func (c *ResultController) GetAll(ctx *gin.Context) {
	var results []models.Result
	if err := c.DB.Preload("Site").Preload("Device").Preload("Feature").Preload("Details").Order("id DESC").Find(&results).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, results)
}

// GetByID retrieves a result by ID
func (c *ResultController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var result models.Result
	if err := c.DB.Preload("Site").Preload("Device").Preload("Feature").Preload("Details").First(&result, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Update handles updating a result
func (c *ResultController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var result models.Result
	if err := c.DB.First(&result, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&result); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Save(&result).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// Delete handles deleting a result
func (c *ResultController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.DB.Delete(&models.Result{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Result deleted successfully"})
}

// CreateDetail handles the creation of a new result detail
func (c *ResultController) CreateDetail(ctx *gin.Context) {
	var detail models.ResultDetail
	if err := ctx.ShouldBindJSON(&detail); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Create(&detail).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, detail)
}

// GetDetailsByResultID retrieves all details for a specific result
func (c *ResultController) GetDetailsByResultID(ctx *gin.Context) {
	resultID, err := strconv.ParseUint(ctx.Param("result_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Result ID"})
		return
	}

	var details []models.ResultDetail
	if err := c.DB.Where("result_id = ?", resultID).Find(&details).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, details)
}

// UpdateDetail handles updating a result detail
func (c *ResultController) UpdateDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var detail models.ResultDetail
	if err := c.DB.First(&detail, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Result detail not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&detail); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Save(&detail).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, detail)
}

// DeleteDetail handles deleting a result detail
func (c *ResultController) DeleteDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.DB.Delete(&models.ResultDetail{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Result detail deleted successfully"})
}

// GetResults handles GET request to fetch all results with pagination
func (rc *ResultController) GetResults(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Get filter parameters
	siteID := c.Query("site_id")
	deviceID := c.Query("device_id")
	featureID := c.Query("feature_id")
	status := c.Query("status")

	var results []models.Result
	var total int64
	query := rc.DB.Model(&models.Result{})

	// Apply filters if they exist
	if siteID != "" {
		query = query.Where("site_id = ?", siteID)
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if featureID != "" {
		query = query.Where("feature_id = ?", featureID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count with filters
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count results"})
		return
	}

	// Get paginated results with filters
	if err := query.
		Preload("Site").
		Preload("Device").
		Preload("Feature").
		Preload("Details").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{
			"total": total,
			"page": page,
			"limit": limit,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// GetResultDetails handles GET request to fetch details for a specific result
func (rc *ResultController) GetResultDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result ID"})
		return
	}

	var details []models.ResultDetail
	if err := db.Where("result_id = ?", uint(id)).Find(&details).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch result details"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// CreateResultDetail handles POST request to create a new result detail
func (rc *ResultController) CreateResultDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result ID"})
		return
	}

	var detail models.ResultDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	detail.ResultID = uint(id)

	if err := db.Create(&detail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create result detail"})
		return
	}

	c.JSON(http.StatusCreated, detail)
}

// DeleteResultDetail handles DELETE request to remove a result detail
func (rc *ResultController) DeleteResultDetail(c *gin.Context) {
	resultIDStr := c.Param("id")
	detailIDStr := c.Param("detail_id")

	resultID, err := strconv.ParseUint(resultIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid result ID"})
		return
	}

	detailID, err := strconv.ParseUint(detailIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid detail ID"})
		return
	}

	if err := db.Where("result_id = ? AND id = ?", uint(resultID), uint(detailID)).Delete(&models.ResultDetail{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete result detail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Result detail deleted successfully"})
}

// ExportResults exports test results to Excel
func (rc *ResultController) ExportResults(c *gin.Context) {
	var results []models.Result
	if err := rc.DB.Preload("Site").Preload("Device").Preload("Feature").Order("created_at DESC").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Found %d results to export\n", len(results))

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet
	sheet := "Test Results"
	index, err := f.NewSheet(sheet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Delete default sheet

	// Create headers
	headers := []string{
		"Created At",
		"Status",
		"Site Name",
		"Browser",
		"Device Name",
		"Feature Name",
		"Duration (s)",
		"Error Log",
	}

	// Set headers with style
	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(sheet, cell, header); err != nil {
			fmt.Printf("Error setting header %s: %v\n", header, err)
		}
		if err := f.SetCellStyle(sheet, cell, cell, style); err != nil {
			fmt.Printf("Error setting style for header %s: %v\n", header, err)
		}
	}

	// Add data
	for i, result := range results {
		row := i + 2
		fmt.Printf("Processing result %d: ID=%d, Status=%s\n", i+1, result.ID, result.Status)

		// Format created_at
		createdAt := result.CreatedAt.Format("2006-01-02 15:04:05")
		
		// Get related data
		siteName := "N/A"
		if result.SiteID != 0 {
			siteName = result.Site.Name
		}

		deviceName := "N/A"
		if result.DeviceID != 0 {
			deviceName = result.Device.Name
		}

		featureName := "N/A"
		if result.FeatureID != 0 {
			featureName = result.Feature.Name
		}

		// Set values
		cells := map[string]interface{}{
			fmt.Sprintf("A%d", row): createdAt,
			fmt.Sprintf("B%d", row): result.Status,
			fmt.Sprintf("C%d", row): siteName,
			fmt.Sprintf("D%d", row): result.Browser,
			fmt.Sprintf("E%d", row): deviceName,
			fmt.Sprintf("F%d", row): featureName,
			fmt.Sprintf("G%d", row): result.Duration,
			fmt.Sprintf("H%d", row): result.ErrorLog,
		}

		for cell, value := range cells {
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				fmt.Printf("Error setting cell %s: %v\n", cell, err)
			}
		}
	}

	// Set column widths
	widths := map[string]float64{
		"A": 20, // Created At
		"B": 10, // Status
		"C": 20, // Site Name
		"D": 15, // Browser
		"E": 20, // Device Name
		"F": 20, // Feature Name
		"G": 15, // Duration
		"H": 50, // Error Log
	}

	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			fmt.Printf("Error setting column width for %s: %v\n", col, err)
		}
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("test_results_%s.xlsx", time.Now().Format("20060102_150405"))

	// Set response headers
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Write file to response
	if err := f.Write(c.Writer); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Export completed successfully")
} 