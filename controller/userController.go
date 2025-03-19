package controller

import (
	"context"
	"encoding/json"

	// "encoding/json"
	"log"
	"net/http"

	// "github.com/gorilla/mux"
	"absensi/database"
	"absensi/models"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
    query := "SELECT id, name, email, role, created_at FROM users"
    rows, err := database.DB.Query(context.Background(), query)
    if err != nil {
        log.Println("Error fetching users:", err)
        http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        var user models.User
        var id pgtype.UUID // Gunakan pgtype.UUID untuk membaca UUID dari database

        err := rows.Scan(&id, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
        if err != nil {
            log.Println("Error scanning user:", err)
            continue
        }

        user.ID = id.String() // Konversi UUID ke string

        users = append(users, user)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	
	var data struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := "UPDATE users SET role = $1 WHERE id = $2"
    _, err := database.DB.Exec(context.Background(), query, data.Role, userID)
    if err != nil {
        log.Println("Error updating user role:", err)
        http.Error(w, "Failed to update role", http.StatusInternalServerError)
        return
    }

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User role updated"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	query := "DELETE FROM users WHERE id = $1"
	_, err := database.DB.Exec(context.Background(), query, userID)
	if err != nil {
		log.Println("Error deleting user:", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted"})
}