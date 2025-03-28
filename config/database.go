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

	dsn := os.Getenv("DB_URL_PROD")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erreur de connexion à la base de données :", err)
	}

	DB = db
	fmt.Println("✅ Connecté à PostgreSQL")

	// Auto-migration des modèles
	db.AutoMigrate(&models.User{})
}
