package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	// "strconv"
	"time"

	"github.com/LidoHon/recipes-server/helpers"
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
		body, _ := c.GetRawData()
		log.Println("Raw request body:", string(body))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		log.Println("incoming request body", request)
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request input ", "details": err.Error()})
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
			"title": graphql.String(request.Input.Title),
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
			"title":            graphql.String(request.Input.Title),
			"description":      graphql.String(request.Input.Description),
			"preparation_time": graphql.Int(request.Input.PreparationTime),
			"featured_image":   graphql.String(featuredImage),
			"category_id":      graphql.Int(request.Input.CategoryId),
			"price":            graphql.Int(request.Input.Price),
			"user_id":          graphql.Int(request.Input.UserId),
		}

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create recipe", "details": err.Error()})
			return
		}

		// Insert ingredients
		for _, ingredient := range request.Input.Ingredients {
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
		for _, step := range request.Input.Steps {
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

		c.JSON(http.StatusOK, response.AddRecipeResponseOutput{
			ID:      int(mutation.CreateRecipe.ID),
			Title:   string(mutation.CreateRecipe.Title),
			Message: "Recipe created successfully",
		})
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
		if request.Input.RecipeId <= 0 {
			log.Printf("Invalid recipe ID: %d", request.Input.RecipeId)
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
			"id": graphql.Int(request.Input.RecipeId),
		}

		if err := client.Query(ctx, &query, queryVars); err != nil {
			log.Printf("Failed to query recipe data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query recipe data", "details": err.Error()})
			return
		}

		if len(query.Recipes) == 0 {
			log.Printf("Recipe not found: %d", request.Input.RecipeId)
			c.JSON(http.StatusNotFound, gin.H{"message": "Recipe not found"})
			return
		}

		userID := request.Input.UserId

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
			"id": graphql.Int(request.Input.RecipeId),
		}

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			log.Printf("Failed to delete recipe: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete recipe", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response.RemoveRecipeOutput{
			Message: "recipe deleted successfully",
		})
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

		log.Println("Received update recipe request id: ", request.Input.ID)

		var query struct {
			Recipe []struct {
				ID     graphql.Int    `graphql:"id"`
				Title  graphql.String `graphql:"title"`
				UserId graphql.Int    `graphql:"user_id"`
				FeaturedImage graphql.String `graphql:"featured_image"`
			} `graphql:"recipes(where: {id: {_eq: $id}})"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(request.Input.ID),
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

		if query.Recipe[0].UserId != graphql.Int(request.Input.UserId) {
			c.JSON(http.StatusForbidden, gin.H{"message": "You are not allowed to update this recipe"})
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		featuredImage, ok := imageUrl.(string)
		if !ok || featuredImage == "" {
			featuredImage = string(query.Recipe[0].FeaturedImage)
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
			"id":               graphql.Int(request.Input.ID),
			"title":            graphql.String(request.Input.Title),
			"description":      graphql.String(request.Input.Description),
			"preparation_time": graphql.Int(request.Input.PreparationTime),
			"featured_image":   graphql.String(featuredImage),
			"category_id":      graphql.Int(request.Input.CategoryId),
			"user_id":          graphql.Int(request.Input.UserId),
			"price":            graphql.Int(request.Input.Price),
		}
		log.Println("mutationVarsssss", mutationVars)

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update recipe", "details": err.Error()})
			return
		}

		// // Update ingredients
		// for _, ingredient := range request.Input.Ingredients {
		// 	var ingredientMutation struct {
		// 		UpdateIngredient struct {
		// 			ID       graphql.Int    `graphql:"id"`
		// 			Name     graphql.String `graphql:"name"`
		// 			Quantity graphql.String `graphql:"quantity"`
		// 			RecipeId graphql.String `graphql:"recipe_id"`
		// 		} `graphql:"update_ingredients_by_pk(pk_columns: {id: $id}, _set: {name: $name, quantity: $quantity, recipe_id: $recipe_id})"`
		// 	}

		// 	ingredientVars := map[string]interface{}{
		// 		"id":        graphql.Int(ingredient.ID),
		// 		"name":      graphql.String(ingredient.Name),
		// 		"quantity":  graphql.String(ingredient.Quantity),
		// 		"recipe_id": graphql.Int(mutation.UpdateRecipe.ID),
		// 	}

		// 	log.Println("ingredientVars from update_ingredients_by_pk", ingredientVars)

		// 	if err := client.Mutate(ctx, &ingredientMutation, ingredientVars); err != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update ingredient", "details": err.Error()})
		// 		return
		// 	}
		// }

		// // Update steps
		// for _, step := range request.Input.Steps {
		// 	var stepMutation struct {
		// 		UpdateStep struct {
		// 			ID          graphql.Int    `graphql:"id"`
		// 			StepNumber  graphql.Int    `graphql:"step_number"`
		// 			Instruction graphql.String `graphql:"instruction"`
		// 			RecipeId    graphql.Int    `graphql:"recipe_id"`
		// 		} `graphql:"update_steps_by_pk(pk_columns: {id: $id}, _set: {step_number: $step_number, instruction: $instruction, recipe_id: $recipe_id})"`
		// 	}

		// 	stepVars := map[string]interface{}{
		// 		"id":          graphql.Int(step.ID),
		// 		"step_number": graphql.Int(step.StepNumber),
		// 		"instruction": graphql.String(step.Instruction),
		// 		"recipe_id":   graphql.Int(mutation.UpdateRecipe.ID),
		// 	}

		// 	if err := client.Mutate(ctx, &stepMutation, stepVars); err != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update step", "details": err.Error()})
		// 		return
		// 	}
		// }

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

func UploadImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var req requests.ImageUpLoadRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		if validationError := validate.Struct(req); validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": validationError.Error()})
			return
		}

		imageUrls, exists := c.Get("imageUrls")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Image URLs not found"})
			return
		}

		imageUrlsArray, ok := imageUrls.([]string)
		if !ok || len(imageUrlsArray) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid image URLs"})
			return
		}

		featuredImageIndex := req.Input.FeaturedImageIndex
		if featuredImageIndex < 0 || featuredImageIndex >=len(imageUrlsArray){
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid featured image index"})
			return
		}

		featuredImageUrl := imageUrlsArray[featuredImageIndex]

		// 1.  update recipe table to set feature image url
		var updateRecipeMutation struct {
			UpdateRecipe struct{
				ID graphql.Int `graphql:"id"`
			}`graphql:" update_recipes_by_pk(pk_columns: {id: $recipeId}, _set: {featured_image: $featuredImageUrl})"`
			
		}
		updateMutationVars := map[string]interface{}{
			"recipeId": graphql.Int(req.Input.RecipeId),
			"featuredImageUrl": graphql.String(featuredImageUrl),
		}

		if err := client.Mutate(ctx, &updateRecipeMutation, updateMutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update recipe", "details": err.Error()})
			return
		}

// 2. now we insert recipe images to recipe images table

		var uploadResponse []response.ImageUploadResponse

		for i, imageUrl := range imageUrlsArray {
			isFeatured := i == featuredImageIndex

			var insertImageMutation struct {
				InsertRecipeImage struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"insert_recipe_images_one(object: {image_url: $imageUrl, recipe_id: $recipeId, is_featured: $isFeatured})"`
			}

			mutationVars := map[string]interface{}{
				"imageUrl": graphql.String(imageUrl),
				"recipeId": graphql.Int(req.Input.RecipeId),
				"isFeatured": graphql.Boolean(isFeatured),
			}

			if err := client.Mutate(ctx, &insertImageMutation, mutationVars); err != nil {
				log.Println("Error inserting an image:", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload image", "details": err.Error()})
				return
			}

			uploadResponse = append(uploadResponse, response.ImageUploadResponse{
				Url: graphql.String(strconv.Itoa(int(insertImageMutation.InsertRecipeImage.ID))),
			})
		}

		c.JSON(http.StatusOK, uploadResponse)
	}
}

func UpdateImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var request requests.ImageUpLoadRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		validationError := validate.Struct(request)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": validationError.Error()})
			return
		}

		imageUrls, exists := c.Get("imageUrls")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Image URLs not found"})
			return
		}

		imageUrlsArray, ok := imageUrls.([]string)
		if !ok || len(imageUrlsArray) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid image URLs"})
			return
		}

		if len(imageUrlsArray) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "No image URLs provided"})
			return
		}

		var mutation struct {
			DeleteRecipeImage struct {
				AffectedRows graphql.Int `graphql:"affected_rows"`
			} `graohql:"delete_recipe_images(where: {id: {_eq: $recipeImageId}})"`
		}
		MutationVars := map[string]interface{}{
			"recipeImageId": graphql.Int(request.Input.RecipeId),
		}

		err := client.Mutate(ctx, &mutation, MutationVars)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete image", "details": err.Error()})
			return
		}

		var uploadResponse []response.ImageUploadResponse

		for _, imageUrl := range imageUrlsArray {

			var mutation struct {
				InsertRecipeImage struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"insert_recipe_images_one(object: {image_url: $imageUrl, recipe_id: $recipeId})"`
			}
			mutationVars := map[string]interface{}{
				"imageUrl": graphql.String(imageUrl),
				"recipeId": graphql.Int(request.Input.RecipeId),
			}
			if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
				log.Println("Error inserting an image:", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload image", "details": err.Error()})
				return
			}

			uploadResponse = append(uploadResponse, response.ImageUploadResponse{
				Url: graphql.String(strconv.Itoa(int(mutation.InsertRecipeImage.ID))),
			})
		}
		c.JSON(http.StatusOK, uploadResponse)

	}
}

func BuyRecipe() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var req requests.BuyRecipeRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Println("comming request from frontend", req)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid json input", "details": err.Error()})
			println(err.Error())
			return
		}
		if validationError := validate.Struct(req); validationError != nil {
			log.Println("validation failed:", validationError.Error())
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": validationError.Error()})
			return
		}

		var userQuery struct {
			User struct {
				Name  string `graphql:"username"`
				Phone string `graphql:"phone"`
				Email string `graphql:"email"`
			} `graphql:"users_by_pk(id: $id)"`
		}

		userQueryVars := map[string]interface{}{
			"id": graphql.Int(req.Input.BuyerId),
		}

		if err := client.Query(ctx, &userQuery, userQueryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user data aye", "details": err.Error()})
			return
		}
		log.Println("user data after query", userQuery.User)

		var RecipeQuery struct {
			Recipes []struct {
				ID    int `graphql:"id"`
				Price int `graphql:"price"`
				User  struct {
					ID int `graphql:"id"`
				}
			} `graphql:"recipes(where: {id: {_eq: $recipeId}})"`
		}
		var recipeQueryVars = map[string]interface{}{
			"recipeId": graphql.Int(req.Input.RecipeId),
		}
		if err := client.Query(ctx, &RecipeQuery, recipeQueryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query recipe data", "details": err.Error()})
			return
		}
		log.Println("recipe data after query", RecipeQuery.Recipes)
		if len(RecipeQuery.Recipes) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Recipe not found"})
			return
		}

		Price := RecipeQuery.Recipes[0].Price

		// check if the user already bought the recipe
		var soldRecipeQuery struct {
			SoldRecipes []struct {
				ID       int `graphql:"id"`
				BuyerId  int `graphql:"buyer_id"`
				RecipeId int `graphql:"recipe_id"`
				SellerId int `graphql:"seller_id"`
			} `graphql:"sold_recipes(where: {buyer_id: {_eq: $buyer_id}, recipe_id: {_eq: $recipe_id}})"`
		}
		var soldRecipeQueryVars = map[string]interface{}{
			"buyer_id":  graphql.Int(req.Input.BuyerId),
			"recipe_id": graphql.Int(req.Input.RecipeId),
		}
		if err := client.Query(ctx, &soldRecipeQuery, soldRecipeQueryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query sold recipe data", "details": err.Error()})
			return
		}
		log.Println("sold recipe data after query", soldRecipeQuery.SoldRecipes)
		if len(soldRecipeQuery.SoldRecipes) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "you already bought the recipe"})
			return
		}

		var muation struct {
			BuyRecipe struct {
				ID       graphql.Int `graphql:"id"`
				BuyerId  graphql.Int `graphql:"buyer_id"`
				RecipeId graphql.Int `graphql:"recipe_id"`
				Price    graphql.Int `graphql:"price"`
				SellerId graphql.Int `graphql:"seller_id"`
			} `graphql:"insert_sold_recipes_one(object: {buyer_id: $buyer_id, price: $price, recipe_id: $recipe_id, seller_id: $seller_id})"`
		}
		var mutationVars = map[string]interface{}{
			"buyer_id":  graphql.Int(req.Input.BuyerId),
			"recipe_id": graphql.Int(req.Input.RecipeId),
			"price":     graphql.Int(Price),
			"seller_id": graphql.Int(RecipeQuery.Recipes[0].User.ID),
		}
		log.Println("a user with id " + strconv.Itoa(int(req.Input.BuyerId)) + " is trying to buy a recipe " + strconv.Itoa(int(req.Input.RecipeId)))
		if err := client.Mutate(ctx, &muation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to buy recipe", "details": err.Error()})
			return
		}

		var chapaForm helpers.PaymentRequest
		chapaForm.Amount = Price

		chapaForm.PhoneNumber = userQuery.User.Phone
		chapaForm.Email = userQuery.User.Email
		chapaForm.FirstName = userQuery.User.Name
		chapaForm.Currency = "ETB"
		chapaForm.TxRef = fmt.Sprintf("buy-recipe-%d", muation.BuyRecipe.ID)
		chapaForm.ReturnURL = os.Getenv("CHAPA_REDIRECT_URL")
		chapaForm.CallbackURL = os.Getenv("CHAPA_CALLBACK_URL")
		chapaForm.CustomizationTitle = "Buying a recipe"
		chapaForm.CustomizationDesc = "Buying a recipe"

		log.Println("recipe bought successfully", muation.BuyRecipe)

		paymentResponse, err := helpers.InitPayment(&chapaForm)
		fmt.Println("chapa form to get the user detail..", chapaForm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to buy recipe", "details": err.Error()})
			return
		}
		log.Println("payment initiated successfully", paymentResponse)
		log.Println("ChapaResponseeeee: ", paymentResponse.ChapaResponse)

		var paymentID graphql.Int
		var checkoutUrl graphql.String

		if paymentResponse.Status {
			data, ok := paymentResponse.ChapaResponse["data"].(map[string]interface{})
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve checkout_url from ChapaResponse"})
				return
			}

			checkoutURL, ok := data["checkout_url"].(string)
			if !ok {
				log.Panicln("checkout url not found in ChapaResponse", data)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to retrieve checkout_url from data"})
				return
			}

			log.Println("checkout url: ", checkoutURL)

			paymentMutation := struct {
				InsertPayment struct {
					ID            graphql.Int `json:"id"`
					SoldRecipeId  graphql.Int `graphql:"sold_recipe_id"`
					TxRef         string      `graphql:"tx_ref"`
					CheckoutURL   string      `graphql:"checkout_url"`
					Amount        graphql.Int `graphql:"amount"`
					Currency      string      `graphql:"currency"`
					PaymentMethod string      `graphql:"payment_method"`
					Status        string      `graphql:"payment_status"`
					BuyerId       graphql.Int `graphql:"buyer_id"`
					RecipeId      graphql.Int `graphql:"recipe_id"`
				} `graphql:"insert_payments_one(object: {sold_recipe_id: $soldRecipeId, tx_ref: $txRef, checkout_url: $checkoutUrl, amount: $amount, currency: $currency, payment_method: $paymentMethod, payment_status: $status, buyer_id: $buyerId, recipe_id: $recipeId})"`
			}{}

			paymentVars := map[string]interface{}{
				"soldRecipeId":  graphql.Int(muation.BuyRecipe.ID),
				"txRef":         graphql.String(chapaForm.TxRef),
				"checkoutUrl":   graphql.String(checkoutURL),
				"amount":        graphql.Int(Price),
				"currency":      graphql.String(chapaForm.Currency),
				"paymentMethod": graphql.String("chapa"),
				"status":        graphql.String("pending"),
				"buyerId":       graphql.Int(req.Input.BuyerId),
				"recipeId":      graphql.Int(req.Input.RecipeId),
			}

			if err := client.Mutate(ctx, &paymentMutation, paymentVars); err != nil {
				log.Println("Failed to insert payment record:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to insert payment record", "details": err.Error()})
				return
			}
			log.Println("payment inserted successfully", paymentMutation.InsertPayment)
			paymentID = paymentMutation.InsertPayment.ID
			checkoutUrl = graphql.String(paymentMutation.InsertPayment.CheckoutURL)
			log.Println("checccccclkout url", paymentMutation.InsertPayment.CheckoutURL)
			log.Println("Payment record inserted successfully and checkout url:.", checkoutURL)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "paymnent initation failed", "details": paymentResponse.Message})
		}
		res := response.BuyRecipeOutput{
			Message:     "recipe bought successfully",
			PaymentId:   graphql.Int(paymentID),
			CheckOutUrl: checkoutUrl,
			BuyerId:     graphql.Int(muation.BuyRecipe.BuyerId),
			RecipeId:    graphql.Int(muation.BuyRecipe.RecipeId),
			Price:       graphql.Int(Price),
			SellerId:    graphql.Int(muation.BuyRecipe.SellerId),
		}
		log.Println("Returning response:", res)
		log.Printf("Response object before sending: %+v\n", res)

		c.JSON(http.StatusOK, res)
	}
}

func ProcessPayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var req requests.PaymentProcessRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}
		log.Println("tx_ref_id:", req.Input.Id)

		isVerified, err := helpers.VerifyPayment(req.Input.TxRef)
		if err != nil || !isVerified {
			log.Println("Payment verification failed:", err)
			c.JSON(http.StatusOK, gin.H{"message": "payment verification failed", "details": err.Error()})
			return
		}

		type UpdatePaymentMutation struct {
			Status graphql.String `graphql:"payment_status"`
			TxRef  graphql.String `graphql:"tx_ref"`
		}

		var mutation struct {
			UpdatePayment struct {
				Returning []UpdatePaymentMutation `graphql:"returning"`
			} `graphql:"update_payments(where: {id: {_eq: $id}}, _set: {payment_status: \"paid\"})"`
		}

		mutationVars := map[string]interface{}{
			"id": graphql.Int(req.Input.Id),
		}

		err = client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("Failed to update payment status:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update payment status"})
			return
		}

		log.Println("Payment status updated successfully for Payment ID:", req.Input.Id)
		log.Println("Payment status updated successfully for Payment txref:", req.Input.TxRef)
		res := response.ProcessPaymentOutput{
			Message: "payment processed successfully and status is updated",
			Status:  "success",
		}
		c.JSON(http.StatusOK, res)
	}
}

func PaymentCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		// Extract query parameters
		txRef := c.Query("trx_ref")
		status := c.Query("status")

		log.Println("Received payment callback - TxRef:", txRef, "Status:", status)

		if status != "success" {
			log.Println("Payment failed or not completed:", status)
			c.JSON(http.StatusOK, gin.H{"message": "payment not successful"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		isVerified, err := helpers.VerifyPayment(txRef)
		if err != nil || !isVerified {
			log.Println("Payment verification failed from backend callback:", err)
			c.JSON(http.StatusOK, gin.H{"message": "payment verification failed", "details": err.Error()})
			return
		}

		type UpdatePaymentMutation struct {
			Status graphql.String `graphql:"payment_status"`
			TxRef  graphql.String `graphql:"tx_ref"`
		}

		var mutation struct {
			UpdatePayment struct {
				Returning []UpdatePaymentMutation `graphql:"returning"`
			} `graphql:"update_payments(where: {tx_ref: {_eq: $txRef}}, _set: {payment_status: \"paid\"})"`
		}
		mutationVars := map[string]interface{}{
			"txRef": graphql.String(txRef),
		}

		// Execute the mutation to update the payment status
		err = client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("Failed to update payment status:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update payment status"})
			return
		}

		// Log success and return a response
		log.Println("Payment status updated successfully for TxRef:", txRef)
		c.JSON(http.StatusOK, gin.H{"message": "payment processed successfully"})
	}
}
