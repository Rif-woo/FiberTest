package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("JWT_SECRET")

// VerifyJWT valide un token JWT et retourne ses claims
func VerifyJWT(tokenString string) (*jwt.Token, error) {
	// Parse et validation du token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Vérification de l'algorithme de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("mauvaise méthode de signature")
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
