package main

import (
	"log"
	"os"

	"mcp-server/internal/database"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "file:cloud-pricing.db"
	}
	db, err := database.Connect(url)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migration complete")
}
