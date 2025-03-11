package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
)

var DB *pgx.Conn

func InitDB() {
	connStr := os.Getenv("SUPABASE_DB_URL") // Gunakan connection string dari .env
	if connStr == "" {
		log.Fatal("SUPABASE_DB_URL is not set in environment variables")
	}

	var err error
	DB, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to Supabase successfully")
}
