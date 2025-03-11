package main

import (
	"absensi/database"
	"absensi/routes"
	"log"
	"net/http"
	"os"

	"firebase.google.com/go/auth"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Inisialisasi Supabase
	database.InitDB()

	// Setup router
	router := routes.SetupRoutes(&auth.Client{})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
