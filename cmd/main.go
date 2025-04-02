package main

import (
	"log"
	"os" // Nécessaire pour lire les variables d'environnement (clés API)

	"github.com/Azertdev/FiberTest/config"
	// Assurez-vous que le chemin vers vos adapters est correct
	"github.com/Azertdev/FiberTest/internal/adapters"
	"github.com/Azertdev/FiberTest/internal/handlers"
	"github.com/Azertdev/FiberTest/internal/repositories"
	"github.com/Azertdev/FiberTest/internal/routes"
	"github.com/Azertdev/FiberTest/internal/services"

	// Assurez-vous que le chemin vers vos utils (pour TranscriptUtil) est correct
	"github.com/Azertdev/FiberTest/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"   // Exemple: ajout de CORS
	"github.com/gofiber/fiber/v2/middleware/logger" // Exemple: ajout de Logger
)

func main() {
	// --- 1. Initialisation de la Configuration et de la Base de Données ---
	config.InitDB() // Ceci initialise la variable globale config.DB
	// Vous pourriez vouloir une fonction qui retourne (db, err) pour une meilleure gestion d'erreur.
	if config.DB == nil {
		log.Fatal("Échec de l'initialisation de la base de données (config.DB est nil)")
	}

	// Charger les clés API depuis les variables d'environnement
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	groqAPIKey := os.Getenv("GROQ_API_KEY")

	// Vérification critique des clés API
	if youtubeAPIKey == "" {
		log.Fatal("ERREUR FATALE: Variable d'environnement YOUTUBE_API_KEY manquante.")
	}
	if groqAPIKey == "" {
		log.Fatal("ERREUR FATALE: Variable d'environnement GROQ_API_KEY manquante.")
	}
	log.Println("Configuration et clés API chargées.")

	// --- 2. Initialisation des Repositories ---
	userRepo := repositories.NewUserRepository(config.DB)
	commentRepo := repositories.NewCommentRepository(config.DB) // Nécessaire si CommentService.FindAll/FindByID est utilisé
	insightRepo := repositories.NewInsightRepository(config.DB) // <-- Initialisation du nouveau repository
	log.Println("Repositories initialisés.")

	// --- 3. Initialisation des Adapters et Utilitaires (avec configuration) ---
	// Assurez-vous que ces constructeurs existent dans les packages correspondants
	// et qu'ils retournent les interfaces définies dans services/interfaces.go
	youtubeAdapter, err := adapters.NewYouTubeAdapter(youtubeAPIKey)
	if err != nil {
		log.Fatalf("Échec de l'initialisation de l'adapter YouTube: %v", err)
	}
	groqAdapter := adapters.NewGroqAdapter(groqAPIKey)
	transcriptUtil := utils.NewTranscriptUtil()
	log.Println("Adapters et Utilitaires initialisés.")

	// --- 4. Initialisation de Tous les Services (Injection des dépendances) ---
	// Appel de NewAllServices avec la nouvelle signature et toutes les dépendances
	allServices := services.NewAllServices(
		userRepo,
		commentRepo,    // Passez commentRepo (ou nil si CommentService ne l'utilise plus)
		insightRepo,    // <-- Injection de insightRepo
		youtubeAdapter, // <-- Injection de youtubeAdapter
		groqAdapter,    // <-- Injection de groqAdapter
		transcriptUtil, // <-- Injection de transcriptUtil
	)
	log.Println("Services initialisés.")

	// --- 5. Initialisation des Handlers (passe les services appropriés) ---
	userHandler := handlers.NewUserHandler(allServices.UserService)
	commentHandler := handlers.NewCommentHandler(allServices.CommentService) // Passe CommentService
	// Créez un InsightHandler si vous avez des routes spécifiques pour les insights
	// insightHandler := handlers.NewInsightHandler(allServices.InsightService) // Exemple si vous créez un InsightService

	log.Println("Handlers initialisés.")

	// --- 6. Configuration de l'Application Fiber (Middlewares, Routes) ---
	app := fiber.New()

	// Ajout de middlewares utiles
	app.Use(cors.New())   // Autoriser les requêtes Cross-Origin (configurez selon vos besoins)
	app.Use(logger.New()) // Logger les requêtes HTTP

	// Création d'un groupe pour les routes API (bonne pratique)
	// api := app.Group("/api")

	// Configuration des routes en passant le groupe API et les handlers
	routes.SetupUserRoutes(app, userHandler)
	// Assurez-vous que la fonction s'appelle SetupCommentsRoutes ou SetupCommentRoutes
	routes.SetupCommentsRoutes(app, commentHandler)
	// Ajoutez les routes pour les insights si nécessaire
	// routes.SetupInsightRoutes(api, insightHandler) // Exemple

	log.Println("Application Fiber et routes configurées.")

	// --- 7. Démarrage du Serveur Fiber ---
	port := ":3001" // Vous pouvez aussi lire ceci depuis une variable d'environnement ou config
	log.Printf("Démarrage du serveur EngageSense sur le port %s", port)
	err = app.Listen(port)
	if err != nil {
		log.Fatalf("Échec du démarrage du serveur Fiber: %v", err)
	}
}
