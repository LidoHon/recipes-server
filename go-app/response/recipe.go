package response

type AddRecipeResponseOutput struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type UpdateRecipeResponse struct {
	Message string `json:"message"`
}
type RemoveRecipeOutput struct {
	Message string `json:"message"`
}
