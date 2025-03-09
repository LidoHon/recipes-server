package requests

type AddRecipeRequest struct {
	Input struct {
		Title           string              `json:"title" validate:"required"`
		Description     string              `json:"description" validate:"required"`
		PreparationTime int                 `json:"preparation_time" validate:"required"`
		FeaturedImage   string              `json:"featured_image,omitempty"`
		UserId          int                 `json:"user_id" `
		CategoryId      int                 `json:"category_id" validate:"required"`
		Ingredients     []IngredientRequest `json:"ingredients" validate:"required,dive"`
		Steps           []StepRequest       `json:"steps" validate:"required,dive"`
		Price           int                 `json:"price" validate:"required"`
	} `json:"input"`
}

type IngredientRequest struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name" validate:"required"`
	Quantity string `json:"quantity,omitempty"`
}

type StepRequest struct {
	ID          int    `json:"id,omitempty"`
	StepNumber  int    `json:"step_number" validate:"required"`
	Instruction string `json:"instruction" validate:"required"`
}

type DeleteRecipeRequest struct {
	Input struct {
		RecipeId int `json:"id" validate:"required"`
		UserId   int `json:"user_id" validate:"required"`
	} `json:"input"`
}

type UpdateRecipeRequest struct {
	Input struct {
		ID              int                 `json:"id" validate:"required"`
		UserId          int                 `json:"user_id" validate:"required"`
		Title           string              `json:"title" validate:"required"`
		Description     string              `json:"description" validate:"required"`
		PreparationTime int                 `json:"preparation_time" validate:"required"`
		FeaturedImage   string              `json:"featured_image,omitempty"`
		CategoryId      int                 `json:"category_id" validate:"required"`
		Ingredients     []IngredientRequest `json:"ingredients" validate:"required,dive"`
		Steps           []StepRequest       `json:"steps" validate:"required,dive"`
		Price           int                 `json:"price" validate:"required"`
	} `json:"input"`
}

type ImageUpLoadRequest struct {
	Input struct {
		RecipeId int `json:"recipe_id" validate:"required"`
	} `json:"input"`
}

type BuyRecipeRequest struct {
	Input struct {
		ID           int    `json:"id" `
		BuyerId      int    `json:"buyer_id" `
		RecipeId     int    `json:"recipe_id" `
		PurchaseDate string `json:"purchase_date" `
		Price        int    `json:"price" `
		SellerId     int    `json:"seller_id" `
	}`json:"input"`
}

type PaymentProcessRequest struct {
	Input struct {
		TxRef string `json:"tx_ref"`
		Id    int    `json:"id"`
	} `json:"input"`
}

