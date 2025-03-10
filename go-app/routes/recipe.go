package routes

import (
	"github.com/LidoHon/recipes-server/controllers"
	"github.com/LidoHon/recipes-server/middlewares"
	"github.com/gin-gonic/gin"
)

func RecipeRoutes(router *gin.Engine) {
	RecipeRoutes := router.Group("/api/recipes")
	{
		RecipeRoutes.POST("/create", middlewares.ImageUpload(), controllers.AddRecipe())
		RecipeRoutes.DELETE("/delete", controllers.DeleteRecipe())
		RecipeRoutes.PUT("/update", middlewares.ImageUpload(), controllers.UpdateRecipe())
		RecipeRoutes.GET("/", controllers.GetAllRecipes())
		RecipeRoutes.POST("/uploadImg", middlewares.ImageUpload(), controllers.UploadImage())
		RecipeRoutes.POST("/updateImg", middlewares.ImageUpload(), controllers.UpdateImage())
		RecipeRoutes.POST("/buy-recipe", controllers.BuyRecipe())
		RecipeRoutes.PUT("/verify-payment", controllers.ProcessPayment())
		RecipeRoutes.GET("/callback", controllers.PaymentCallback())
	}
}
