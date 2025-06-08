package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"qa-automation-system/backend/models"
)

// FeatureController handles feature-related operations
type FeatureController struct {
	DB *gorm.DB
}

// NewFeatureController creates a new feature controller
func NewFeatureController(db *gorm.DB) *FeatureController {
	return &FeatureController{DB: db}
}

// Create handles the creation of a new feature
func (c *FeatureController) Create(ctx *gin.Context) {
	var feature models.Feature
	if err := ctx.ShouldBindJSON(&feature); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Create(&feature).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, feature)
}

// GetAll retrieves all features
func (c *FeatureController) GetAll(ctx *gin.Context) {
	var features []models.Feature
	if err := c.DB.Find(&features).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, features)
}

// GetByID retrieves a feature by ID
func (c *FeatureController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var feature models.Feature
	if err := c.DB.First(&feature, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	ctx.JSON(http.StatusOK, feature)
}

// Update handles updating a feature
func (c *FeatureController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var feature models.Feature
	if err := c.DB.First(&feature, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&feature); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Save(&feature).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, feature)
}

// Delete handles deleting a feature
func (c *FeatureController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.DB.Delete(&models.Feature{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Feature deleted successfully"})
} 