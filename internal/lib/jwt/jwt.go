package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

const (
	accessTokenDuration  = time.Minute * 15
	refreshTokenDuration = time.Hour * 128
)

type AuthDetails struct {
	AuthUuid string
	UserId   uint64
}

func CreateAccessToken(authD AuthDetails) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"auth_uuid":  authD.AuthUuid,
		"user_id":    authD.UserId,
		"exp":        time.Now().Add(accessTokenDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

func CreateRefreshToken(authD AuthDetails) (string, error) {
	claims := jwt.MapClaims{
		"auth_uuid": authD.AuthUuid,
		"exp":       time.Now().Add(refreshTokenDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

func CreateTokenPair(authD AuthDetails) (map[string]string, error) {
	accessToken, err := CreateAccessToken(authD)
	if err != nil {
		return nil, err
	}

	refreshToken, err := CreateRefreshToken(authD)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func VerifyAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
}
