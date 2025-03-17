package controller

import (
	"absensi/database"
	"absensi/models"
	"absensi/utils"
	"context" // Import context package
	"log"
	"net/http"

	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
)

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
    var user models.User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Hash password before saving
    hashedPassword, err := utils.HashPassword(user.Password)
    if err != nil {
        http.Error(w, "Password hashing failed", http.StatusInternalServerError)
        return
    }
    user.Password = hashedPassword

	// Set role default ke "employee"
    if user.Role == "" {
        user.Role = "employee"
    }

    // Insert user into database
    query := `INSERT INTO users (name, email, password, role, created_at) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	   // Execute the query using Exec with context
	   _, err = database.DB.Exec(context.Background(), query, user.Name, user.Email, user.Password, user.Role, time.Now())
if err != nil {
    log.Println("Database error:", err) // Menampilkan error dari database
    http.Error(w, "User registration failed", http.StatusInternalServerError)
    return
}

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

// Login handles user login
func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Println("Login attempt for email:", user.Email)

	// Cek apakah email ada di database
	var storedPassword string
	var userID string
	query := `SELECT id, password FROM users WHERE email=$1`

	row := database.DB.QueryRow(context.Background(), query, user.Email)
	err = row.Scan(&userID, &storedPassword)
	if err != nil {
		if err == pgx.ErrNoRows { // Jika tidak ada data dengan email tersebut
			http.Error(w, "User not found", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Println("Stored password:", storedPassword)

	// Cek password
	valid := utils.CheckPasswordHash(user.Password, storedPassword)
	if !valid {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Membuat token JWT
	token, err := utils.GenerateJWT(userID)
	if err != nil {
		http.Error(w, "JWT generation failed", http.StatusInternalServerError)
		return
	}

	// Kirim token sebagai response
	response := map[string]string{"token": token}
	json.NewEncoder(w).Encode(response)
}

