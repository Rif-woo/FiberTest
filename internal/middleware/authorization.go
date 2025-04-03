package middleware

import (
	"github.com/Azertdev/FiberTest/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

// Middleware pour vérifier le JWT dans l'en-tête Authorization
func JWTMiddleware(c *fiber.Ctx) error {
	// Récupérer le token depuis l'en-tête Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token manquant"})
	}

	tokenString := authHeader[7:]

	// Vérification du token
	token, err := utils.VerifyJWT(tokenString)
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token invalide"})
	}

	// Extraction des claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Attacher les informations du token au contexte
		c.Locals("username", claims["username"])
		c.Locals("userId", claims["id"]) // Si tu stockes un ID dans le token
		return c.Next()
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Accès non autorisé"})
}
