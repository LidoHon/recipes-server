package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/LidoHon/recipes-server/controllers"
	"github.com/LidoHon/recipes-server/helpers"
	"github.com/LidoHon/recipes-server/middlewares"

	// "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	// google open auth
	incomingRoutes.GET("/auth/google", func(c *gin.Context) {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	incomingRoutes.GET("/auth/google/callback", func(c *gin.Context) {
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			responseError := map[string]string{
				"message": "Error logging in with google",
				"details": err.Error(),
			}
			jsonData, _ := json.Marshal(responseError)
			escapedData := url.QueryEscape(string(jsonData))
			clientLoginUrl := os.Getenv("CLIENT_LOGIN_URL")
			redirectUrl := clientLoginUrl + "?error=" + escapedData
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
			return
		}

		token, refreshToken, userId, userRole, err := helpers.HandleAuth(user.Email, user.Name, user.AvatarURL, user.UserID, user.Provider)
		if err != nil {
			log.Println("Something went wrong:", err.Error())
			responseError := map[string]string{
				"message": "Error logging in with google",
				"detail":  err.Error(),
			}
			jsonData, _ := json.Marshal(responseError)
			escapedData := url.QueryEscape(string(jsonData))
			clientLoginUrl := os.Getenv("CLIENT_LOGIN_URL")
			fmt.Println("client loginurl",clientLoginUrl)
			redirectUrl := clientLoginUrl + "?error=" + escapedData
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
			return
		}

		responseData := map[string]string{
			"message":      "user registerd sucessfully",
			"token":        token,
			"refreshToken": refreshToken,
			"id":           strconv.Itoa(userId),
			"role":         userRole,
		}
		jsonData, _ := json.Marshal(responseData)
		escapedData := url.QueryEscape(string(jsonData))
		redirectUrl := os.Getenv("CLIENT_HOMEPAGE_URL") + "?data=" + escapedData
		log.Println("Redirecting to:", redirectUrl)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	})

	// github open auth

	incomingRoutes.GET("/auth/github", func(c *gin.Context) {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	incomingRoutes.GET("/auth/github/callback", func(c *gin.Context) {
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			responseError := map[string]string{
				"message": "error logging inn with github",
				"details": err.Error(),
			}
			jsonData, _ := json.Marshal(responseError)
			escapedData := url.QueryEscape(string(jsonData))
			redirectUrl := os.Getenv("CLIENT_LOGIN_URL") + "?error=" + escapedData
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
			return
		}

		token, refreshToken, userId, userRole, err := helpers.HandleAuth(user.Email, user.Name, user.AvatarURL, user.UserID, user.Provider)
		if err != nil {
			log.Println("Something went wrong:", err.Error())
			responseError := map[string]string{
				"message": "error logging with github",
				"details": err.Error(),
			}
			jsonData, _ := json.Marshal(responseError)
			escapedData := url.QueryEscape(string(jsonData))
			redirectUrl := os.Getenv("CLIENT_LOGIN_URL") + "?error=" + escapedData
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
			return
		}

		responseData := map[string]string{
			"message":      "user registerd sucessfully",
			"token":        token,
			"refreshToken": refreshToken,
			"id":           strconv.Itoa(userId),
			"role":         userRole,
		}
		if len(responseData) == 0 {
			log.Println("Response data is empty")
		}
		log.Println("response data:", responseData)

		jsonData, _ := json.Marshal(responseData)
		escapedData := url.QueryEscape(string(jsonData))
		log.Println("escapedDataaaaaa:", escapedData)
		redirectUrl := os.Getenv("CLIENT_HOMEPAGE_URL") + "?data=" + escapedData
		log.Println("Redirecting to:", redirectUrl)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	})

	userRoutes := incomingRoutes.Group("/api/users")
	{
		userRoutes.POST("/register", middlewares.ImageUpload(), controllers.RegisterUser())
		userRoutes.POST("/verify-email", controllers.VerifyEmail())
		userRoutes.POST("/login", controllers.Login())
		userRoutes.POST("/reset-password", controllers.ResetPassword())
		userRoutes.POST("/update-password", controllers.UpdatePassword())
		userRoutes.POST("/delete", controllers.DeleteUser())
		userRoutes.PUT("/update-profile", middlewares.ImageUpload(), controllers.UpdateProfile())
		userRoutes.DELETE("/delete", controllers.DeleteUserById())
		// userRoutes.GET("/all-users", controllers.GetAllUsers())
		userRoutes.POST("/user", controllers.GetUserById())

	}
}
