package config

import (
	"fmt"
	"log"
	"os"
	"github.com/Azertdev/FiberTest/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Charger les variables d’environnement
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erreur lors du chargement du fichier .env")
	}

	dsn := os.Getenv("DB_URL_NEON")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erreur de connexion à la base de données :", err)
	}

	DB = db
	fmt.Println("✅ Connecté à PostgreSQL")
// Créer le type ENUM 'user_role' dans PostgreSQL
	// Utilisation de la commande SQL pour créer le type ENUM directement dans la base de données
	db.Exec(`CREATE TYPE user_role AS ENUM ('user', 'admin')`)
	db.Exec(`CREATE TYPE user_platform AS ENUM ('instagram', 'twitter', 'youtube')`)
	db.Exec(`CREATE TYPE subscription_plan AS ENUM ('free', 'pro', 'business')`)
	db.Exec(`CREATE TYPE subscription_Status AS ENUM ('active', 'cancelled')`)
	db.Exec(`CREATE TYPE Notification_Type AS ENUM ('analysis', 'payment', 'alert')`)

// Supprimer la table 'users' si elle existe déjà
// if err := db.Migrator().DropTable(&models.User{}); err != nil {
// 	log.Fatal("Erreur lors de la suppression de la table 'users':", err)
// }
// fmt.Println("🗑️ Table 'users' supprimée avec succès")

// Auto-migrer les modèles, ce qui recréera la table 'users' avec le nouveau schéma
if err := db.AutoMigrate(&models.User{}, &models.Subscription{}, &models.Comment{}, &models.Insight{}, &models.Notification{}); err != nil {
	log.Fatal("Erreur lors de la migration des modèles :", err)
}
fmt.Println("✅ Tables recréées avec succès")
}
