package routes

import (
	"github.com/LidoHon/recipes-server/controllers"
	"github.com/gin-gonic/gin"
)

func RecipeRoutes(router *gin.Engine) {
	RecipeRoutes := router.Group("/api/recipes")
	{
		RecipeRoutes.POST("/create", controllers.AddRecipe())
		RecipeRoutes.DELETE("/delete", controllers.DeleteRecipe())
		RecipeRoutes.PUT("/update", controllers.UpdateRecipe())
		router.GET("/api/recipes", controllers.GetAllRecipes()) // Get all recipes
		router.GET("/api/recipes/filter", controllers.FilterRecipes()) // Filter recipes
	}
}
