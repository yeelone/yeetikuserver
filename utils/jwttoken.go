package utils

import (
	"errors"
	"fmt"
	"time"

	c "../config"
	"github.com/dgrijalva/jwt-go"
)

type MyCustomClaims struct {
	ID uint64 `json:"id"`
	jwt.StandardClaims
}

func SetJWTToken(id uint64) string {
	// Expires the token and cookie in 24 hours
	expireToken := time.Now().Add(time.Hour * 24).Unix()
	claims := MyCustomClaims{
		id,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    c.Config.Domain,
		},
	}
	// Create the token using your claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Signs the token with a secret.
	signedToken, _ := token.SignedString([]byte(c.Config.SecretKey))

	return signedToken
}

func TokenParse(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.Config.SecretKey), nil
	})

	return token, err
}

func ValidateJWTToken(tokenString string) error {
	token, _ := TokenParse(tokenString)

	if token == nil {
		return nil
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return nil
	} else {
		return errors.New("error token")
	}
}

//获取用户ID
func ParseUserProperty(tokenString string) string {
	if len(tokenString) > 0 {
		token, _ := TokenParse(tokenString)
		if token == nil {
			return ""
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return fmt.Sprintf("%v", claims["id"])
		} else {
			return ""
		}
	}

	return ""
}
