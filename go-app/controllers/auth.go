package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/LidoHon/recipes-server/helpers"
	"github.com/LidoHon/recipes-server/libs"
	"github.com/LidoHon/recipes-server/requests"
	"github.com/LidoHon/recipes-server/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shurcooL/graphql"
)

var validate = validator.New()

// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user with the provided details
// @Tags users Registration
// @Accept json
// @Produce json
// @Param input body requests.RegisterRequest true "User registration details"
// @Success 200 {object} response.SignedUpUserOutput
// @Failure 400 {object} gin.H "Invalid input data"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /users/register [post]
func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Declare a variable to hold the input data from the json req body
		var request requests.RegisterRequest

		// Bind the JSON request body to the request struct
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("Validation error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input data"})
			return
		}

		// Validate the input data
		if err := validate.Struct(request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Validation failed", "details": err.Error()})
			return
		}

		// Check if the user already exists
		var query struct {
			User []struct {
				ID       graphql.Int    `graphql:"id"`
				Name     graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Password graphql.String `graphql:"password"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}

		variable := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}

		if err := client.Query(ctx, &query, variable); err != nil {
			log.Println("Error querying existing user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query the existing user", "details": err.Error()})
			return
		}

		if len(query.User) != 0 {
			log.Println("User already exists")
			c.JSON(http.StatusBadRequest, gin.H{"message": "User already exists"})
			return

		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		profilePictureURL, ok := imageUrl.(string)
		if !ok {
			profilePictureURL = ""
		}
		// hash the password
		password := helpers.HashPassword(request.Input.Password)

		var mutation struct {
			CreateUser struct {
				ID       graphql.Int    `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"insert_users_one(object: {username: $userName, email: $email, password: $password, profile: $profile, role:$role, phone:$phone})"`
		}
		mutationVariables := map[string]interface{}{
			"userName": graphql.String(request.Input.UserName),
			"email":    graphql.String(request.Input.Email),
			"password": graphql.String(password),
			"phone":    graphql.String(request.Input.Phone),
			"profile":  graphql.String(profilePictureURL),
			"role":     graphql.String(request.Input.Role + "user"),
		}
		err := client.Mutate(context.Background(), &mutation, mutationVariables)
		if err != nil {
			log.Println("Failed to register user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to register user"})
			return
		}
		// query newly created user
		var regesteredUserQuery struct {
			User []struct {
				ID       graphql.Int    `graphql:"id"`
				Name     graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Password graphql.String `graphql:"password"`
				Role     graphql.String `graphql:"role"`
				TokenId  graphql.String `graphql:"tokenId"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}

		regesteredUserVariable := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}
		err = client.Query(context.Background(), &regesteredUserQuery, regesteredUserVariable)
		if err != nil {
			log.Println("Error querying the registered user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query the registered user"})
			return
		}
		if len(regesteredUserQuery.User) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "user not found after registration maybe he flead to another table:)"})
		}
		user := regesteredUserQuery.User[0]
		// generate email verification token
		emailVerficationToken, err := helpers.GenerateToken()

		if err != nil {
			log.Println("Error generating email verification token:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate email verification token", "detail": err.Error()})
			return
		}

		var VerficationEmailMutation struct {
			InsertedToken struct {
				ID    graphql.Int    `graphql:"id"`
				Token graphql.String `graphql:"token"`
			} `graphql:"insert_email_verification_tokens_one(object: {token: $token, user_id: $userId})"`
		}

		// "insert_email_verification_tokens_one(object: {token: $token, user_id: $userId})"

		verficatonEmailVariable := map[string]interface{}{
			"token":  graphql.String(emailVerficationToken),
			"userId": graphql.Int(user.ID),
		}

		err = client.Mutate(context.Background(), &VerficationEmailMutation, verficatonEmailVariable)
		if err != nil {
			log.Printf("Error inserting email verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to register verfication token"})
			return
		}

		var UpdateUserTokenMutation struct {
			UpdatedUser struct {
				ID      graphql.Int `graphql:"id"`
				TokenID graphql.Int `graphql:"tokenId"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {tokenId: $tokenId})"`
		}

		updatedUserVariables := map[string]interface{}{
			"userId":  graphql.Int(user.ID),
			"tokenId": VerficationEmailMutation.InsertedToken.ID,
		}

		err = client.Mutate(context.Background(), &UpdateUserTokenMutation, updatedUserVariables)

		if err != nil {
			log.Printf("error updating user with tokenId: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user tokenid", "details": err.Error()})
			return
		}
		log.Printf("User %d successfully updated with tokenId %d", user.ID, UpdateUserTokenMutation.UpdatedUser.TokenID)

		verificationLink := os.Getenv("RETURN_URL") + "?verification_token=" + emailVerficationToken + "&user_id=" + strconv.Itoa(int(user.ID))

		emailForm := helpers.EmailData{
			Name:    string(user.Name),
			Email:   string(user.Email),
			Link:    verificationLink,
			Subject: "Verifying your email",
		}

		res, errString := helpers.SendEmail(
			[]string{emailForm.Email},
			"verifyEmail.html",
			emailForm,
		)
		if !res {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email verification email, please contact support.", "details": errString})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(string(user.Email), string(user.Name), string(user.Role), string(user.TokenId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to  generate jwt token", "details": err.Error()})
			return

		}

		response := response.SignedUpUserOutput{
			ID:           int(user.ID),
			UserName:     string(user.Name),
			Email:        string(user.Email),
			Token:        token,
			RefreshToken: refreshToken,
			Role:         string(user.Role),
		}
		c.JSON(http.StatusOK, response)

	}

}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verifies a user's email by checking the verification token and updating the user's status.
// @Tags User Verification
// @Accept json
// @Produce json
// @Param request body requests.EmailVerifyRequest true "User email verification request"
// @Success 200 {object} response.VerifyEmailResponse "Email verified successfully"
// @Failure 400 {object} gin.H "Invalid input data"
// @Failure 401 {object} gin.H "Invalid or expired verification token"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /users/verify-email [post]
func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.EmailVerifyRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		// Validate the request body
		validationError := validate.Struct(request)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
			return

		}

		// Fetch the token from the database based on the provided
		var query struct {
			Tokens []struct {
				ID        graphql.Int    `graphql:"id"`
				Token     graphql.String `graphql:"token"`
				UserId    graphql.Int    `graphql:"user_id"`
				ExpiresAt graphql.String `graphql:"expires_at"`
			} `graphql:"email_verification_tokens(where: {token: {_eq: $token}})"`
		}
		variables := map[string]interface{}{
			"token": graphql.String(request.Input.VerificationToken),
		}
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query the verification token", "details": err.Error()})
			return
		}
		// Check if token is found
		if len(query.Tokens) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid verification token"})
			return
		}

		//  validate if the token is expired
		expirationTime, err := time.Parse(time.RFC3339, string(query.Tokens[0].ExpiresAt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to parse token expiration date", "details": err.Error()})
			return
		}

		if time.Now().After(expirationTime) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "expired verification token"})
			return
		}

		var mutation struct {
			UpdateUser struct {
				ID       graphql.ID     `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
				IsVerified graphql.Boolean `graphql:"is_email_verified"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $id}, _set: {is_email_verified: $status})"`
		}

		mutationVariables := map[string]interface{}{
			"id":     graphql.Int(request.Input.UserId),
			"status": graphql.Boolean(true),
		}

		err = client.Mutate(context.Background(), &mutation, mutationVariables)
		if err != nil {
			log.Printf("Error updating user email verification status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user email verification status", "details": err.Error()})
			return
		}

		// delete the token after it has been used
		var deleteMutation struct {
			DeleteToken struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_email_verification_tokens_by_pk(id: $id)"`
		}
		deleteVariables := map[string]interface{}{
			"id": query.Tokens[0].ID,
		}
		err = client.Mutate(context.Background(), &deleteMutation, deleteVariables)
		if err != nil {
			log.Printf("Error deleting email verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete email verification token", "details": err.Error()})
			return
		}

		// return success message

		res := response.VerifyEmailResponse{
			Message: string("Email verified successfully"),
		}

		c.JSON(http.StatusOK, res)
	}
}

// Login handles user authentication
// @Summary User login
// @Description Authenticates a user with email and password, returning access and refresh tokens
// @Tags User Login
// @Accept json
// @Produce json
// @Param request body requests.LoginRequest true "Login request body"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} gin.H "Invalid input or bad request"
// @Failure 401 {object} gin.H "Invalid credentials or unverified email"
// @Failure 500 {object} gin.H "Failed to query user or generate tokens"
// @Router /users/login [post]
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.LoginRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		var query struct {
			User []struct {
				ID              graphql.Int     `graphql:"id"`
				Name            graphql.String  `graphql:"username"`
				Email           graphql.String  `graphql:"email"`
				Password        graphql.String  `graphql:"password"`
				Role            graphql.String  `graphql:"role"`
				TokenId         graphql.Int     `graphql:"tokenId"`
				IsEmailVerified graphql.Boolean `graphql:"is_email_verified"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}

		variables := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}

		if err := client.Query(context.Background(), &query, variables); err != nil {
			log.Printf("failed to query the user with email %s: %v", request.Input.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user"})
			return
		}

		if len(query.User) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
			return

		}
		user := query.User[0]

		// if !user.IsEmailVerified {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"message": "You need to verify your account first to login"})
		// 	return
		// }

		if valid, msg := helpers.VerifyPassword(request.Input.Password, string(user.Password)); !valid {
			c.JSON(http.StatusBadRequest, gin.H{"message": msg})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(string(user.Email), string(user.Name), string(user.Role), string(user.TokenId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to generate tokens",
				"details": err.Error(),
			})
			return
		}

		response := response.LoginResponse{
			User: &response.UserResponse{
				ID:           user.ID,
				Token:        graphql.String(token),
				RefreshToken: graphql.String(refreshToken),
				Email:        graphql.String(user.Email),
				Name:         graphql.String(user.Name),
				Role:         graphql.String(user.Role),
			},
		}

		c.JSON(http.StatusOK, response)

	}
}

// ResetPassword  request godoc
// @Summary sending Reset password request
// @Description Send a password reset token to the user's email address if they have a verified email.
// @Tags Password Reset Request
// @Accept json
// @Produce json
// @Param request body requests.PasswordResetRequest true "Password Reset Request"
// @Success 200 {object} response.ResetRequestOutput "Password reset request success"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 401 {object} gin.H "Unauthorized - Invalid credentials or unverified email"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /users/reset-password [post]
func ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.PasswordResetRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("invalid input:%v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		// fetch user by email
		var query struct {
			User []struct {
				ID              graphql.Int     `graphql:"id"`
				Name            graphql.String  `graphql:"username"`
				Email           graphql.String  `graphql:"email"`
				Password        graphql.String  `graphql:"password"`
				Role            graphql.String  `graphql:"role"`
				TokenId         graphql.Int     `graphql:"tokenId"`
				IsEmailVerified graphql.Boolean `graphql:"is_email_verified"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}
		queryVars := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			log.Printf("failed to query a user data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query user data", "details": err.Error()})
			return
		}

		if len(query.User) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
			return
		}

		user := query.User[0]
		if !user.IsEmailVerified {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Please verify your email first to reset your password"})
			return
		}

		token, err := helpers.GenerateToken()
		if err != nil {
			log.Printf("Failed to generate token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate token", "details": err.Error()})
			return
		}
		// Insert reset token into the database
		var mutation struct {
			InsertedToken struct {
				ID    graphql.Int    `graphql:"id"`
				Token graphql.String `graphql:"token"`
			} `graphql:"insert_email_verification_tokens_one(object: {token: $token, user_id: $userId})"`
		}

		mutationVars := map[string]interface{}{
			"token":  graphql.String(token),
			"userId": graphql.Int(user.ID),
		}

		if err := client.Mutate(context.Background(), &mutation, mutationVars); err != nil {
			log.Printf("failed to register reset token")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to register rewste token", "details": err.Error()})
			return
		}

		// update the user table to add tokenid
		var UpdateUserTokenMutation struct {
			UpdatedUser struct {
				ID      graphql.Int `graphql:"id"`
				TokenID graphql.Int `graphql:"tokenId"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {tokenId: $tokenId})"`
		}

		updatedUserVariables := map[string]interface{}{
			"userId":  graphql.Int(user.ID),
			"tokenId": mutation.InsertedToken.ID,
		}

		err = client.Mutate(context.Background(), &UpdateUserTokenMutation, updatedUserVariables)

		if err != nil {
			log.Printf("error updating user with tokenId: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user tokenid", "details": err.Error()})
			return
		}
		log.Printf("User %d successfully updated with tokenId %d", user.ID, UpdateUserTokenMutation.UpdatedUser.TokenID)

		verificationLink := os.Getenv("RESET_PASS_URL") + "/password-reset?token=" + token + "&id=" + strconv.Itoa(int(user.ID))

		// Send password reset email
		emailData := helpers.EmailData{
			Name:    string(user.Name),
			Email:   string(user.Email),
			Link:    verificationLink,
			Subject: "Reset your password",
		}

		if success, errString := helpers.SendEmail(
			[]string{emailData.Email},
			"passReset.html",
			emailData,
		); !success {
			log.Printf("Failed to send password reset email: %v", errString)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send password reset email", "details": errString})
			return
		}

		response := response.ResetRequestOutput{
			ID:      mutation.InsertedToken.ID,
			Message: "we have sent you an email to reset your password, please go and check your email to reset your password",
		}

		c.JSON(http.StatusOK, response)

	}
}

// Reset Password godoc
// @Summary Reset password
// @Description Send a password reset token to the user's email address if they have a verified email.
// @Tags Reset Password
// @Accept json
// @Produce json
// @Param request body requests.UpdatePasswordRequest true "Password Reset Request"
// @Success 200 {object} response.UpdatePasswordResponse "Password reset request success"
// @Failure 400 {object} gin.H "Invalid input"
// @Failure 401 {object} gin.H "Unauthorized - Invalid credentials or unverified email"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /users/update-password [post]
func UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.UpdatePasswordRequest

		if err := c.ShouldBind(&request); err != nil {
			log.Printf("error binding request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		log.Printf("Incoming request: %+v", request)
		if validationError := validate.Struct(request); validationError != nil {
			log.Printf("Validation error: %v", validationError)
			c.JSON(http.StatusBadRequest, gin.H{"message": "validation failed", "details": validationError.Error()})
			return
		}

		var query struct {
			Tokens []struct {
				ID     graphql.Int    `graphql:"id"`
				Token  graphql.String `graphql:"token"`
				UserId graphql.Int    `graphql:"user_id"`
			} `graphql:"email_verification_tokens(where: {token: {_eq: $token}})"`
		}
		queryVars := map[string]interface{}{
			"token": graphql.String(request.Input.Token),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			log.Printf("failed to query token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query token", "details": err.Error()})
			return
		}

		if len(query.Tokens) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid tokens"})
			return
		}

		password := helpers.HashPassword(request.Input.Password)

		var mutation struct {
			UpdateUser struct {
				ID       graphql.Int    `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $id}, _set: {password: $password})"`
		}

		mutationVars := map[string]interface{}{
			"id":       graphql.Int(request.Input.UserId),
			"password": graphql.String(password),
		}

		err := client.Mutate(context.Background(), &mutation, mutationVars)
		if err != nil {
			log.Printf("failed to update the password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update the password", "details": err.Error()})
			return
		}
		emailForm := helpers.EmailData{
			Name:    string(mutation.UpdateUser.UserName),
			Email:   string(mutation.UpdateUser.Email),
			Subject: "Password reseted successfully!",
		}

		sucess, errorString := helpers.SendEmail(
			[]string{emailForm.Email}, "resetPasswordSuccess.html", emailForm,
		)
		if !sucess {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to send password reset email", "details": errorString})
			return
		}

		response := response.UpdatePasswordResponse{
			Message: "your password has been reseted successfully",
		}
		log.Println(response)

		c.JSON(http.StatusOK, response)

		// delete the token after it has been used
		var deleteMutation struct {
			DeleteToken struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_email_verification_tokens_by_pk(id: $id)"`
		}
		deleteVariables := map[string]interface{}{
			"id": query.Tokens[0].ID,
		}
		err = client.Mutate(context.Background(), &deleteMutation, deleteVariables)
		if err != nil {
			log.Printf("Error deleting email verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete email verification token", "details": err.Error()})
			return
		}

	}
}

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var eventPayload requests.EventPayload

		if err := c.ShouldBindJSON(&eventPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		if eventPayload.Event.Op == "DELETE" && eventPayload.Event.Data.Old != nil {
			oldUserData := eventPayload.Event.Data.Old

			emailData := helpers.EmailData{
				Name:    oldUserData.UserName,
				Email:   oldUserData.Email,
				Subject: "Account Deletion Confirmation",
			}

			success, errorString := helpers.SendEmail(
				[]string{emailData.Email},
				"deletedAccount.html",
				emailData,
			)

			if !success {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send account deletion email", "details": errorString})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "Account removal email has been sent", "email": oldUserData.Email})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"message": "Something went wrong", "details": "Invalid input"})
	}
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile details of a user, including username, phone, profile picture, and role.
// @Tags update user profile
// @Accept json
// @Produce json
// @Param input body requests.UpdateRequest true "User profile update details"
// @Success 200 {object} response.UpdateResponce "Profile updated successfully"
// @Failure 400 {object} gin.H "Invalid input data"
// @Failure 422 {object} gin.H "Validation error"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /users/update-profile [put]
func UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var req requests.UpdateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("error from jsonbind %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}

		if validationErr := validate.Struct(req); validationErr != nil {
			fmt.Println("Validation Error:", validationErr.Error())
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"updateProfile": validationErr.Error(),
			},
			)
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		proPicUrl, ok := imageUrl.(string)
		if !ok {
			proPicUrl = ""
		}

		roleToUse := req.Input.Role
		if roleToUse == "" {
			var query struct {
				Users []struct {
					Role *string `graphql:"role"`
				} `graphql:"users(where: {id: {_eq: $userId}})"`
			}

			queryVars := map[string]interface{}{
				"userId": graphql.Int(req.Input.UserId),
			}

			err := client.Query(context.Background(), &query, queryVars)
			if err != nil {
				log.Println("Error fetching existing role:", err)
			}

			log.Println("Query response to check role:", query.Users)

			if len(query.Users) > 0 && query.Users[0].Role != nil {
				roleToUse = *query.Users[0].Role
				log.Println("Fetched Role from DB:", roleToUse)
			} else {
				roleToUse = "user"
			}
		}

		log.Println("Final Role to Use:", roleToUse)

		if proPicUrl == "" {
			var mutation struct {
				UpdateProfile struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {username: $userName, phone: $Phone, role: $Role})"`
			}

			mutationVars := map[string]interface{}{
				"userId":   graphql.Int(req.Input.UserId),
				"userName": graphql.String(req.Input.UserName),
				"Phone":    graphql.String(req.Input.Phone),
				"Role":     graphql.String(roleToUse),
			}

			err := client.Mutate(context.Background(), &mutation, mutationVars)
			if err != nil {
				log.Println("Failed to update user profile:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user profile", "details": err.Error()})
				return
			}

			res := response.UpdateResponce{
				Message: "Profile updated successfully",
			}
			c.JSON(http.StatusOK, res)
		} else {
			var mutation struct {
				UpdateProfile struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"update_users_by_pk(pk_columns: {id: userId}, _set: {username: $userName, phone: $Phone, profile: $Profile, role: $Role})"`
			}

			mutationVars := map[string]interface{}{
				"userId":   graphql.Int(req.Input.UserId),
				"userName": graphql.String(req.Input.UserName),
				"Phone":    graphql.String(req.Input.Phone),
				"Profile":  graphql.String(proPicUrl),
				"Role":     graphql.String(req.Input.Role),
			}

			err := client.Mutate(context.Background(), &mutation, mutationVars)
			if err != nil {
				log.Println("Failed to update user profile:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user profile"})
				return
			}
			res := response.UpdateResponce{
				Message: "Profile updated successfully",
			}

			c.JSON(http.StatusOK, res)
		}

	}

}

// DeleteUserById deletes a user by their ID.
// @Summary Delete a user
// @Description Deletes a user from the system using their user ID. If the deletion is successful, an email is sent to confirm the account deletion. If the email fails to send, the deletion is rolled back.
// @Tags delete user
// @Accept json
// @Produce json
// @Param request body requests.DeleteUserWithIdInput true "User ID to delete"
// @Success 200 {object} response.DeleteUserWithEmailResponse "User deleted and email sent successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Failed to delete user, rollback attempted if necessary"
// @Router /users/delete [delete]
func DeleteUserById() gin.HandlerFunc {
	return func(c *gin.Context) {

		client := libs.SetupGraphqlClient()

		var input requests.DeleteUserWithIdInput
		if err := c.ShouldBindBodyWithJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		var query struct {
			User struct {
				ID       graphql.Int    `graphql:"id"`
				Email    graphql.String `graphql:"email"`
				UserName graphql.String `graphql:"username"`
			} `graphql:"users_by_pk(id: $id)"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(input.UserID),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user data", "details": err.Error()})
			return
		}

		var mutation struct {
			DeleteUser struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_users_by_pk(id: $id)"`
		}

		mutationVars := map[string]interface{}{
			"id": graphql.Int(input.UserID),
		}

		if err := client.Mutate(context.Background(), &mutation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete user", "details": err.Error()})
			return
		}
		emailData := helpers.EmailData{
			Name:    string(query.User.UserName),
			Email:   string(query.User.Email),
			Subject: "Account Deletion Confirmation",
		}

		success, errorString := helpers.SendEmail(
			[]string{emailData.Email},
			"deletedAccount.html", emailData,
		)

		if !success {
			// Undo the deletion if email fails
			var undoMutation struct {
				InsertUser struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"insert_users_one(object: {id: $id, email: $email, user_name: $userName})"`
			}
			undoVars := map[string]interface{}{
				"id":       graphql.Int(query.User.ID),
				"email":    graphql.String(query.User.Email),
				"userName": graphql.String(query.User.UserName),
			}
			if undoErr := client.Mutate(context.Background(), &undoMutation, undoVars); undoErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to rollback deletion", "details": undoErr.Error()})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email, deletion rolled back", "details": errorString})
			return
		}

		c.JSON(http.StatusOK, response.DeleteUserWithEmailResponse{
			Status:  "success",
			Message: "User deleted and email sent successfully",
		})
	}
}

func UpdateProfilePicture() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var req requests.UpdateProfileImage

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("Error from JSON bind: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Profile picture is required"})
			return
		}

		proPicUrl, ok := imageUrl.(string)
		if !ok || proPicUrl == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid profile picture URL"})
			return
		}

		// Define the GraphQL mutation
		var mutation struct {
			UpdateProfilePicture struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {profile: $Profile})"`
		}

		// Mutation variables
		mutationVars := map[string]interface{}{
			"userId":  graphql.Int(req.Input.UserId),
			"Profile": graphql.String(proPicUrl),
		}

		err := client.Mutate(context.Background(), &mutation, mutationVars)
		if err != nil {
			log.Println("Failed to update profile picture:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update profile picture", "details": err.Error()})
			return
		}
		res := response.UpdateProfileResponce{
			Message: "Profile picture updated successfully",
		}
		c.JSON(http.StatusOK, res)
	}
}

// GetAllUsers retrieves all users.
// @Summary Get all users
// @Description Fetches a list of all users from the database.
// @Tags Get all users
// @Accept json
// @Produce json
// @Success 200 {array} response.AllUserResponse "List of users"
// @Failure 500 {object} map[string]string "Failed to query user data"
// @Router /users/all-users [get]

// func GetAllUsers() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		client := libs.SetupGraphqlClient()

// 		var query struct {
// 			Users []struct {
// 				ID       graphql.Int    `graphql:"id"`
// 				UserName graphql.String `graphql:"username"`
// 				Email    graphql.String `graphql:"email"`
// 			} `graphql:"users"`
// 		}

// 		if err := client.Query(context.Background(), &query, nil); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user data", "details": err.Error()})
// 			return
// 		}
// 		// Convert query.Users to AllUserResponse format
// 		var response []response.AllUserResponse
// 		for _, user := range query.Users {
// 			response = append(response, AllUserResponse{
// 				ID:       int(user.ID),
// 				UserName: string(user.UserName),
// 				Email:    string(user.Email),
// 			})
// 		}

// 		c.JSON(http.StatusOK, response)
// 	}
// }

// GetUserById retrieves a user by their ID.
// @Summary Get user by ID
// @Description Fetches a user from the database using their unique ID.
// @Tags Get user by ID
// @Accept json
// @Produce json
// @Param request body requests.GetUserByIdInput true "User ID input"
// @Success 200 {object} response.SingleUserResponse "User details"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Failed to query user data"
// @Router /users/user [post]
func GetUserById() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var req requests.GetUserByIdInput
		if err := c.ShouldBindBodyWithJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		var query struct {
			User struct {
				ID       graphql.Int    `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
			} `graphql:"users_by_pk(id: $id)"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(req.Input.UserID),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user data", "details": err.Error()})
			return
		}

		response := response.SingleUserResponse{
			ID:       int(query.User.ID),
			UserName: string(query.User.UserName),
			Email:    string(query.User.Email),
		}
		c.JSON(http.StatusOK, response)
	}
}
