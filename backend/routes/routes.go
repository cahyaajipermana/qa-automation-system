package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"qa-automation-system/backend/controllers"
	"time"
	"github.com/gin-contrib/cors"
)

// SetupRouter configures all the routes for the application
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize controllers
	siteController := controllers.NewSiteController(db)
	deviceController := controllers.NewDeviceController(db)
	featureController := controllers.NewFeatureController(db)
	resultController := controllers.NewResultController(db)

	// API routes
	api := router.Group("/api")
	{
		// Sites routes
		sites := api.Group("/sites")
		{
			sites.POST("", siteController.Create)
			sites.GET("", siteController.GetAll)
			sites.GET("/:id", siteController.GetByID)
			sites.PUT("/:id", siteController.Update)
			sites.DELETE("/:id", siteController.Delete)
		}

		// Devices routes
		devices := api.Group("/devices")
		{
			devices.POST("", deviceController.Create)
			devices.GET("", deviceController.GetAll)
			devices.GET("/:id", deviceController.GetByID)
			devices.PUT("/:id", deviceController.Update)
			devices.DELETE("/:id", deviceController.Delete)
		}

		// Features routes
		features := api.Group("/features")
		{
			features.POST("", featureController.Create)
			features.GET("", featureController.GetAll)
			features.GET("/:id", featureController.GetByID)
			features.PUT("/:id", featureController.Update)
			features.DELETE("/:id", featureController.Delete)
		}

		// Results routes
		results := api.Group("/results")
		{
			results.GET("", resultController.GetResults)
			results.GET("/export", resultController.ExportResults)
			results.GET("/:id", resultController.GetByID)
			results.POST("", resultController.Create)
			results.PUT("/:id", resultController.Update)
			results.DELETE("/:id", resultController.Delete)
			results.GET("/:id/details", resultController.GetResultDetails)
			results.POST("/:id/details", resultController.CreateResultDetail)
			results.DELETE("/:id/details/:detail_id", resultController.DeleteResultDetail)
		}
	}

	return router
} 