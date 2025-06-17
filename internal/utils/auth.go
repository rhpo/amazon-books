package utils

import (
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(TOKEN_EXPIRY).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("APP_SECRET")))
}

func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("APP_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string), nil
	}

	return "", jwt.ErrSignatureInvalid
}

func HashPassword(password string, salt string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		Report("Failed to hash password: "+err.Error(), true)
		return ""
	}
	return string(hashedPassword)
}

func CheckPasswordHash(password, hash string, salt string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt)) == nil
}
