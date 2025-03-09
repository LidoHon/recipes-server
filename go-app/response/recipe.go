package response

import "github.com/shurcooL/graphql"

type AddRecipeResponseOutput struct {
	Title   string `json:"title"`
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type UpdateRecipeResponse struct {
	Message string `json:"message"`
}
type RemoveRecipeOutput struct {
	Message string `json:"message"`
}

type ImageUploadResponse struct {
	Url graphql.String `json:"url"`
}

type BuyRecipeOutput struct {
	PaymentId   graphql.Int    `json:"payment_id"`
	CheckOutUrl graphql.String `json:"checkout_url"`
	Message     string         `json:"message"`
	BuyerId     graphql.Int    `json:"buyer_id"`
	SellerId     graphql.Int   `json:"seller_id"`
	RecipeId    graphql.Int    `json:"recipe_id"`
	Price       graphql.Int    `json:"price"`
}

type ProcessPaymentOutput struct {
	Message string         `json:"message"`
	Status  graphql.String `json:"status"`
}