package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// token for email verfication,and password reset
func GenerateToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	resetToken := hex.EncodeToString(token)

	hash, err := bcrypt.GenerateFromPassword([]byte(resetToken), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// jwt token and refresh token

type HasuraClaims struct {
	AllowedRoles []string `json:"x-hasura-allowed-roles"`
	DefaultRole  string   `json:"x-hasura-default-role"`
	UserID       string   `json:"x-hasura-user-id"`
	Role         string   `json:"x-hasura-role"`
}

type SignedDetails struct {
	HasuraClaims `json:"https://hasura.io/jwt/claims"`
	Email        string `json:"X-Hasura-User-Email"`
	UserName     string `json:"X-Hasura-User-Name"`
	UserId       int    `json:"X-Hasura-User-Id"`
	jwt.StandardClaims
}

func GenerateAllTokens(email string, userName string, role string, uid string, id int) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		HasuraClaims: HasuraClaims{
			AllowedRoles: []string{"user", "admin", "systemAdmin"},
			DefaultRole:  role,
			UserID:       uid,
			Role:         role,
		},
		Email:    email,
		UserId:   id,
		UserName: userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	refreshClaims := &SignedDetails{StandardClaims: jwt.StandardClaims{
		ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(68)).Unix(),
	},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	fmt.Println("JWT_SECRET_KEY:", os.Getenv("JWT_SECRET_KEY"))

	if err != nil {
		return "", "", err
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		log.Panic(err)
		return
	}
	return token, refreshToken, nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		
	})

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "The token is invalid"
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "Token has expired"
		return
	}
	return claims, msg
}
