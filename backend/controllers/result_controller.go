package controllers

import (
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
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