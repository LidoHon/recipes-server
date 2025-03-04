package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LidoHon/recipes-server/libs"
	"github.com/LidoHon/recipes-server/requests"
	"github.com/LidoHon/recipes-server/response"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/graphql"
)

// create recipe
func AddRecipe() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var request requests.AddRecipeRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		if err := validate.Struct(request); err != nil {
			log.Printf("Validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Validation failed",
				"details": err.Error(),
			})
			return
		}

		log.Println("Received recipe request: ", request)

		// Check if recipe already exists (same title)
		var query struct {
			Recipe []struct {
				ID    graphql.Int    `graphql:"id"`
				Title graphql.String `graphql:"title"`
			} `graphql:"recipes(where: {title: {_eq: $title}})"`
		}

		queryVars := map[string]interface{}{
			"title": graphql.String(request.Title),
		}

		if err := client.Query(ctx, &query, queryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query recipes", "details": err.Error()})
			return
		}

		if len(query.Recipe) != 0 {
			log.Println("Recipe already exists with the same title")
			c.JSON(http.StatusBadRequest, gin.H{"message": "Recipe already exists"})
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		featuredImage, ok := imageUrl.(string)
		if !ok {
			featuredImage = ""
		}

		// Create new recipe
		var mutation struct {
			CreateRecipe struct {
				ID              graphql.Int    `graphql:"id"`
				Title           graphql.String `graphql:"title"`
				Description     graphql.String `graphql:"description"`
				PreparationTime graphql.Int    `graphql:"preparation_time"`
				FeaturedImage   graphql.String `graphql:"featured_image"`
				UserId          graphql.Int    `graphql:"user_id"`
				CategoryId      graphql.Int    `graphql:"category_id"`
				Price           graphql.Int    `graphql:"price"`
			} `graphql:"insert_recipes_one(object: {title: $title, description: $description, preparation_time: $preparation_time, featured_image: $featured_image, category_id: $category_id, price: $price, user_id: $user_id})"`
		}

		mutationVars := map[string]interface{}{
			"title":            graphql.String(request.Title),
			"description":      graphql.String(request.Description),
			"preparation_time": graphql.Int(request.PreparationTime),
			"featured_image":   graphql.String(featuredImage),
			"category_id":      graphql.Int(request.CategoryId),
			"price":            graphql.Int(request.Price),
			"user_id":          graphql.Int(request.UserId),
		}

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create recipe", "details": err.Error()})
			return
		}

		// Insert ingredients
		for _, ingredient := range request.Ingredients {
			var ingredientMutation struct {
				CreateIngredient struct {
					ID       graphql.Int    `graphql:"id"`
					Name     graphql.String `graphql:"name"`
					Quantity graphql.String `graphql:"quantity"`
					RecipeId graphql.Int    `graphql:"recipe_id"`
				} `graphql:"insert_ingredients_one(object: {name: $name, quantity: $quantity, recipe_id: $recipe_id})"`
			}

			ingredientVars := map[string]interface{}{
				"name":      graphql.String(ingredient.Name),
				"quantity":  graphql.String(ingredient.Quantity),
				"recipe_id": graphql.Int(mutation.CreateRecipe.ID),
			}

			if err := client.Mutate(ctx, &ingredientMutation, ingredientVars); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create ingredient", "details": err.Error()})
				return
			}
		}

		// Insert steps
		for _, step := range request.Steps {
			var stepMutation struct {
				CreateStep struct {
					ID          graphql.Int    `graphql:"id"`
					StepNumber  graphql.Int    `graphql:"step_number"`
					Instruction graphql.String `graphql:"instruction"`
					RecipeId    graphql.Int    `graphql:"recipe_id"`
				} `graphql:"insert_steps_one(object: {step_number: $step_number, instruction: $instruction, recipe_id: $recipe_id})"`
			}

			stepVars := map[string]interface{}{
				"step_number": graphql.Int(step.StepNumber),
				"instruction": graphql.String(step.Instruction),
				"recipe_id":   graphql.Int(mutation.CreateRecipe.ID),
			}

			if err := client.Mutate(ctx, &stepMutation, stepVars); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create step", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, response.AddRecipeResponse{Message: "Recipe created successfully"})
	}
}

// delete recipe
func DeleteRecipe() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Define a struct to match the request body
		var request requests.DeleteRecipeRequest

		// Bind JSON request to struct
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON request", "details": err.Error()})
			return
		}

		log.Printf("Request payload: %+v", request)

		// Validate ID
		if request.RecipeId <= 0 {
			log.Printf("Invalid recipe ID: %d", request.RecipeId)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid recipe ID"})
			return
		}

		// Query the database for the recipe
		var query struct {
			Recipes []struct {
				ID     graphql.Int `graphql:"id"`
				UserId graphql.Int `graphql:"user_id"`
			} `graphql:"recipes(where: {id: {_eq: $id}})"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(request.RecipeId),
		}

		if err := client.Query(ctx, &query, queryVars); err != nil {
			log.Printf("Failed to query recipe data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query recipe data", "details": err.Error()})
			return
		}

		if len(query.Recipes) == 0 {
			log.Printf("Recipe not found: %d", request.RecipeId)
			c.JSON(http.StatusNotFound, gin.H{"message": "Recipe not found"})
			return
		}

		userID := request.UserId

		// Authorization check
		if query.Recipes[0].UserId != graphql.Int(userID) {
			log.Println("Unauthorized deletion attempt userid", query.Recipes[0].UserId, userID)
			log.Printf("Unauthorized deletion attempt: userID=%d, recipeUserID=%d", userID, query.Recipes[0].UserId)
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not allowed to delete this recipe"})
			return
		}

		// Perform deletion
		var mutation struct {
			DeleteRecipe struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_recipes_by_pk(id: $id)"`
		}

		mutationVars := map[string]interface{}{
			"id": graphql.Int(request.RecipeId),
		}

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			log.Printf("Failed to delete recipe: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete recipe", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted successfully"})
	}
}

// update recipe

func UpdateRecipe() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var request requests.UpdateRecipeRequest
		log.Println("request", request)
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		if err := validate.Struct(request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		log.Println("Received update recipe request id: ", request.ID)

		var query struct {
			Recipe []struct {
				ID     graphql.Int    `graphql:"id"`
				Title  graphql.String `graphql:"title"`
				UserId graphql.Int    `graphql:"user_id"`
			} `graphql:"recipes(where: {id: {_eq: $id}})"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(request.ID),
		}

		log.Println("queryVars id", queryVars)

		if err := client.Query(ctx, &query, queryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query recipe", "details": err.Error()})
			return
		}

		if len(query.Recipe) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Recipe not found"})
			return
		}

		log.Println("query.Recipe[0].userid", query.Recipe[0].UserId)

		if query.Recipe[0].UserId != graphql.Int(request.UserId) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not allowed to update this recipe"})
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		featuredImage, ok := imageUrl.(string)
		if !ok {
			featuredImage = ""
		}

		var mutation struct {
			UpdateRecipe struct {
				ID              graphql.Int    `graphql:"id"`
				Title           graphql.String `graphql:"title"`
				Description     graphql.String `graphql:"description"`
				PreparationTime graphql.Int    `graphql:"preparation_time"`
				FeaturedImage   graphql.String `graphql:"featured_image"`
				CategoryId      graphql.Int    `graphql:"category_id"`
				UserId          graphql.Int    `graphql:"user_id"`
				Price           graphql.Int    `graphql:"price"`
			} `graphql:"update_recipes_by_pk(pk_columns: {id: $id}, _set: {title: $title, description: $description, preparation_time: $preparation_time, featured_image: $featured_image, category_id: $category_id, price: $price, user_id: $user_id})"`
		}

		mutationVars := map[string]interface{}{
			"id":               graphql.Int(request.ID),
			"title":            graphql.String(request.Title),
			"description":      graphql.String(request.Description),
			"preparation_time": graphql.Int(request.PreparationTime),
			"featured_image":   graphql.String(featuredImage),
			"category_id":      graphql.Int(request.CategoryId),
			"user_id":          graphql.Int(request.UserId),
			"price":            graphql.Int(request.Price),
		}
		log.Println("mutationVarsssss", mutationVars)

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update recipe", "details": err.Error()})
			return
		}

		// Update ingredients
		for _, ingredient := range request.Ingredients {
			var ingredientMutation struct {
				UpdateIngredient struct {
					ID       graphql.Int    `graphql:"id"`
					Name     graphql.String `graphql:"name"`
					Quantity graphql.String `graphql:"quantity"`
				} `graphql:"update_ingredients_by_pk(pk_columns: {id: $id}, _set: {name: $name, quantity: $quantity, recipe_id: $recipe_id})"`
			}

			ingredientVars := map[string]interface{}{
				"id":		graphql.Int(ingredient.ID),
				"name":      graphql.String(ingredient.Name),
				"quantity":  graphql.String(ingredient.Quantity),
				"recipe_id": graphql.Int(mutation.UpdateRecipe.ID),
			}

			log.Println("ingredientVars", ingredientVars)

			if err := client.Mutate(ctx, &ingredientMutation, ingredientVars); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update ingredient", "details": err.Error()})
				return
			}
		}

		// Update steps
		for _, step := range request.Steps {
			var stepMutation struct {
				UpdateStep struct {
					ID          graphql.Int    `graphql:"id"`
					StepNumber  graphql.Int    `graphql:"step_number"`
					Instruction graphql.String `graphql:"instruction"`
					RecipeId    graphql.Int    `graphql:"recipe_id"`
				} `graphql:"update_steps_by_pk(pk_columns: {id: $id}, _set: {step_number: $step_number, instruction: $instruction, recipe_id: $recipe_id})"`
			}

			stepVars := map[string]interface{}{
				"id":		  graphql.Int(step.ID),
				"step_number": graphql.Int(step.StepNumber),
				"instruction": graphql.String(step.Instruction),
				"recipe_id":   graphql.Int(mutation.UpdateRecipe.ID),
			}

			if err := client.Mutate(ctx, &stepMutation, stepVars); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update step", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, response.UpdateRecipeResponse{Message: "Recipe updated successfully"})
	}
}


func GetAllRecipes() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Define the query to fetch all recipes
		var query struct {
			Recipes []struct {
				ID              graphql.Int    `graphql:"id"`
				Title           graphql.String `graphql:"title"`
				Description     graphql.String `graphql:"description"`
				PreparationTime graphql.Int    `graphql:"preparation_time"`
				FeaturedImage   graphql.String `graphql:"featured_image"`
				UserId          graphql.Int    `graphql:"user_id"`
				CategoryId      graphql.Int    `graphql:"category_id"`
				Price           graphql.Int    `graphql:"price"`
			} `graphql:"recipes"`
		}

		// Execute the query
		if err := client.Query(ctx, &query, nil); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch recipes", "details": err.Error()})
			return
		}

		// Return the list of recipes
		c.JSON(http.StatusOK, query.Recipes)
	}
}





