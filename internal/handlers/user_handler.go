package handlers

import (
	"github.com/Azertdev/FiberTest/internal/models"
	"github.com/Azertdev/FiberTest/internal/services"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) UserHandler {
	return UserHandler{userService}
}

// Créer un utilisateur
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	err := h.userService.CreateUser(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Échec de la création"})
	}
	return c.Status(201).JSON(user)
}

// Récupérer tous les utilisateurs
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Impossible de récupérer les utilisateurs"})
	}
	return c.JSON(users)
}

// Récupérer un utilisateur par ID
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID invalide"})
	}
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Utilisateur non trouvé"})
	}
	return c.JSON(user)
}

func (h *UserHandler) LoginHandler(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	userAuth, err := h.userService.AuthenticateUser(user.Username, user.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Échec de la connexion"})
	}

	token, err := h.userService.GenerateJWT(userAuth.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Échec de la creation du token"})
	}
	return c.JSON(fiber.Map{"token": token})
}