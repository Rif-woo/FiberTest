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
	"github.com/gofiber/helmet/v2"

	// Assurez-vous que le chemin vers vos utils (pour TranscriptUtil) est correct
	"github.com/Azertdev/FiberTest/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"   // Exemple: ajout de CORS
	"github.com/gofiber/fiber/v2/middleware/logger" // Exemple: ajout de Logger
)

func main() {
	config.InitDB() // Ceci initialise la variable globale config.DB
	if config.DB == nil {
		log.Fatal("Échec de l'initialisation de la base de données (config.DB est nil)")
	}

	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	groqAPIKey := os.Getenv("GROQ_API_KEY")

	if youtubeAPIKey == "" {
		log.Fatal("ERREUR FATALE: Variable d'environnement YOUTUBE_API_KEY manquante.")
	}
	if groqAPIKey == "" {
		log.Fatal("ERREUR FATALE: Variable d'environnement GROQ_API_KEY manquante.")
	}
	log.Println("Configuration et clés API chargées.")

	allRepositories := repositories.NewAllRepository(config.DB)
	log.Println("Repositories initialisés.")

	youtubeAdapter := adapters.NewYouTubeAdapter(youtubeAPIKey)
	groqAdapter := adapters.NewGroqAdapter(groqAPIKey)
	transcriptUtil := utils.NewTranscriptUtil()
	log.Println("Adapters et Utilitaires initialisés.")

	// --- 4. Initialisation de Tous les Services (Injection des dépendances) ---
	allServices := services.NewAllServices(
		allRepositories,   // <-- Injection de insightRepo
		youtubeAdapter, // <-- Injection de youtubeAdapter
		groqAdapter,    // <-- Injection de groqAdapter
		transcriptUtil, // <-- Injection de transcriptUtil
	)
	log.Println("Services initialisés.")

	// --- 5. Initialisation des Handlers (passe les services appropriés) ---
	allHandlers := handlers.NewAllHandlers(allServices.UserService, allServices.CommentService)
	log.Println("Handlers initialisés.")

	// --- 6. Configuration de l'Application Fiber (Middlewares, Routes) ---
	app := fiber.New()

	// Ajout de middlewares utiles
	app.Use(cors.New())   // Autoriser les requêtes Cross-Origin (configurez selon vos besoins)
	app.Use(logger.New()) // Logger les requêtes HTTP
	app.Use(helmet.New())
	// Création d'un groupe pour les routes API (bonne pratique)
	routes.SetupUserRoutes(app, allHandlers.UserHandler)
	routes.SetupCommentsRoutes(app, allHandlers.CommentHandler)
	log.Println("Application Fiber et routes configurées.")

	// --- 7. Démarrage du Serveur Fiber ---
	port := ":3001" // Vous pouvez aussi lire ceci depuis une variable d'environnement ou config
	log.Printf("Démarrage du serveur EngageSense sur le port %s", port)
	err2 := app.Listen(port)
	if err2 != nil {
		log.Fatalf("Échec du démarrage du serveur Fiber: %v", err2)
	}
}
