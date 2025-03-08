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
