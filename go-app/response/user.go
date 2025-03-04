package response

import "github.com/shurcooL/graphql"

type SignupResponse struct {
	Data struct {
		Signup SignedUpUserOutput `json:"signup"`
	} `json:"data"`
}

type SignedUpUserOutput struct {
	ID           int    `json:"id"`
	UserName     string `json:"userName"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	Role         string `json:"role"`
}

type UserResponse struct {
	ID           graphql.Int    `json:"id"`
	Name         graphql.String `json:"name"`
	Email        graphql.String `json:"email"`
	Token        graphql.String `json:"token"`
	RefreshToken graphql.String `json:"refreshToken"`
	Role         graphql.String `json:"role"`
}

type LoginResponse struct {
	User *UserResponse `json:"user"`
}

type ResetRequestOutput struct {
	ID      graphql.Int    `json:"id"`
	Message graphql.String `json:"message"`
}

type ResetedPasswordOutput struct {
	Message graphql.String `json:"message"`
}

type UpdatePasswordResponse struct {
	Message graphql.String `json:"message"`
}

type UpdateResponce struct {
	Message string `json:"message"`
}

type DeleteResponse struct {
	Message graphql.String `json:"message"`
}

type DeleteUserWithEmailResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type UpdateProfileResponce struct {
	Message graphql.String `json:"message"`
}

type VerifyEmailResponse struct {
	Message string `json:"message"`
}

type AllUserResponse struct {
	ID       int    `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
}
type SingleUserResponse struct {
	ID       int    `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
}
