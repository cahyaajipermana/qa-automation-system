package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"qa-automation-system/backend/models"
)

// SiteController handles site-related operations
type SiteController struct {
	DB *gorm.DB
}

// NewSiteController creates a new site controller
func NewSiteController(db *gorm.DB) *SiteController {
	return &SiteController{DB: db}
}

// Create handles the creation of a new site
func (c *SiteController) Create(ctx *gin.Context) {
	var site models.Site
	if err := ctx.ShouldBindJSON(&site); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Create(&site).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, site)
}

// GetAll retrieves all sites
func (c *SiteController) GetAll(ctx *gin.Context) {
	var sites []models.Site
	if err := c.DB.Find(&sites).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, sites)
}

// GetByID retrieves a site by ID
func (c *SiteController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var site models.Site
	if err := c.DB.First(&site, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Site not found"})
		return
	}

	ctx.JSON(http.StatusOK, site)
}

// Update handles updating a site
func (c *SiteController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var site models.Site
	if err := c.DB.First(&site, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Site not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&site); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Save(&site).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, site)
}

// Delete handles deleting a site
func (c *SiteController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.DB.Delete(&models.Site{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Site deleted successfully"})
} 