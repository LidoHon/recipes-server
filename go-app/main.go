package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/LidoHon/recipes-server/docs"
	"github.com/LidoHon/recipes-server/routes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var store *sessions.CookieStore

// @title User API for  recipes project
// @version 1.0
// @description This is the API documentation for user api.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email liduhon3@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:5000
// @BasePath /api
func main() {
	err := godotenv.Load(`../.env`)
	if err != nil {
		fmt.Println("error loading environment variables", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Load session secret from environment variable
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET is not set in environment variables")
		sessionSecret = "secret"
	}

	// Assign the session store to `gothic` using the session secret
	store = sessions.NewCookieStore([]byte(sessionSecret))
	gothic.Store = store

	router := gin.New()
	router.Use(gin.Logger())

	// Configure Google OAuth
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:"+port+"/auth/google/callback",
			"email", "profile",
		),
		github.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			"http://localhost:"+port+"/auth/github/callback",
		),
	)

	routes.RegisterRoutes(router)

	// Serve Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Printf("Server running on port %s", port)

	router.Run(":" + port)
}
