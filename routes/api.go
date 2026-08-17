package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"sequAcc/controllers"
	_ "sequAcc/docs" // Import generated docs
)

// SetupRoutes configures the application routes
func SetupRoutes(r *gin.Engine) {
	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		achievements := api.Group("/achievements")
		{
			achievements.POST("/shining-star", controllers.CreateShiningStar)
			achievements.GET("/shining-star/:user_id", controllers.GetUserShiningStars)
			achievements.GET("/shining-star", controllers.GetAllShiningStars)
		}
	}
}
