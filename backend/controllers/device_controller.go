package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"qa-automation-system/backend/models"
)

// DeviceController handles device-related operations
type DeviceController struct {
	DB *gorm.DB
}

// NewDeviceController creates a new device controller
func NewDeviceController(db *gorm.DB) *DeviceController {
	return &DeviceController{DB: db}
}

// Create handles the creation of a new device
func (c *DeviceController) Create(ctx *gin.Context) {
	var device models.Device
	if err := ctx.ShouldBindJSON(&device); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Create(&device).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, device)
}

// GetAll retrieves all devices
func (c *DeviceController) GetAll(ctx *gin.Context) {
	var devices []models.Device
	if err := c.DB.Find(&devices).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, devices)
}

// GetByID retrieves a device by ID
func (c *DeviceController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var device models.Device
	if err := c.DB.First(&device, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	ctx.JSON(http.StatusOK, device)
}

// Update handles updating a device
func (c *DeviceController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var device models.Device
	if err := c.DB.First(&device, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&device); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.DB.Save(&device).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, device)
}

// Delete handles deleting a device
func (c *DeviceController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := c.DB.Delete(&models.Device{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
} 