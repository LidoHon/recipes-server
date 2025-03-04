package helpers

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Panic(err)
	}
	return string(hashedPassword)
}

func VerifyPassword(userPassword string, hashedpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(userPassword))

	if err != nil {
		return false, "incorrect credentials"

	}
	return true, ""
}
