package services

import (
	// "os" // Ne plus lire os.Getenv ici

	// Importer les interfaces des adapters/utils que vous avez définies (probablement dans interfaces.go)
	// Si elles ne sont pas dans le même package 'services', ajustez le chemin.
	// "github.com/Azertdev/FiberTest/internal/services" // Si interfaces.go est séparé

	"github.com/Azertdev/FiberTest/internal/repositories"
	"log" // Pour la validation des dépendances
)

// AllServices regroupe tous les services de l'application.
// J'ai supprimé GroqService car il est maintenant injecté dans CommentService via GroqAdapter.
// Si d'autres services ont besoin de l'adapter Groq, vous pouvez l'ajouter ici (ex: GroqAdapter GroqAdapter).
type AllServices struct {
	UserService    UserService
	CommentService CommentService
	// Pas de GroqService ici, car il est géré via l'adapter injecté.
	// Ajoutez d'autres services ici si nécessaire (ex: InsightService, etc.)
}

// NewAllServices est le constructeur pour la structure AllServices.
// Il prend maintenant TOUTES les dépendances nécessaires pour TOUS les services qu'il initialise.
func NewAllServices(
	// Dépendances pour UserService
	userRepo repositories.UserRepository,

	// Dépendances pour CommentService
	commentRepo    repositories.CommentRepository, // Peut être nil si FindAll/FindByID non utilisés par CommentService
	insightRepo    repositories.InsightRepository,  // <- Ajouté
	youtubeAdapter YouTubeAdapter,                  // <- Ajouté (Interface)
	groqAdapter    GroqAdapter,                     // <- Ajouté (Interface)
	transcriptUtil TranscriptUtil,                  // <- Ajouté (Interface)

	// Ajoutez les dépendances pour d'autres services ici
) *AllServices {

	// --- Validation des dépendances critiques ---
	// C'est une bonne pratique de vérifier que les dépendances essentielles ne sont pas nil.
	if userRepo == nil {
		log.Fatal("ERREUR FATALE: UserRepository manquant lors de la création de AllServices")
	}
	if insightRepo == nil {
		log.Fatal("ERREUR FATALE: InsightRepository manquant lors de la création de AllServices")
	}
	if youtubeAdapter == nil {
		log.Fatal("ERREUR FATALE: YouTubeAdapter manquant lors de la création de AllServices")
	}
	if groqAdapter == nil {
		log.Fatal("ERREUR FATALE: GroqAdapter manquant lors de la création de AllServices")
	}
	if transcriptUtil == nil {
		log.Fatal("ERREUR FATALE: TranscriptUtil manquant lors de la création de AllServices")
	}
	// Note: commentRepo peut être nil si CommentService ne l'utilise plus pour FindAll/FindByID.

	// --- Initialisation des services ---
	userService := NewUserService(userRepo)

	// Initialisation de CommentService avec TOUTES ses dépendances
	commentService := NewCommentService(
		commentRepo, // Passez le repo Commentaire (ou nil)
		insightRepo,
		youtubeAdapter,
		groqAdapter,
		transcriptUtil,
	)

	// Initialisation d'autres services ici...

	// --- Retourner la structure AllServices remplie ---
	return &AllServices{
		UserService:    userService,
		CommentService: commentService,
		// Pas de GroqService directement ici.
		// Initialiser d'autres champs de service ici...
	}
}