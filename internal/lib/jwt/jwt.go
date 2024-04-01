package jwt

import (
	"errors"
	"github.com/17HIERARCH70/SocialManager/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

// NewToken creates new JWT token for given user.
func NewToken(user models.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	JWTSecretKey := os.Getenv("JWT_SECRET_KEY")
	if JWTSecretKey == "" {
		return "", errors.New("JWT secret key is not set")
	}
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
