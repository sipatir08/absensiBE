package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"absensi/database"
	"absensi/models"
)

func CheckIn(w http.ResponseWriter, r *http.Request) {
	var att models.Attendance

	// Ambil data dari request body
	err := json.NewDecoder(r.Body).Decode(&att)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Ambil UserID dari context
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "User ID is missing", http.StatusBadRequest)
		return
	}
	att.UserID = userID.(string) // Pastikan tipe data sesuai dengan yang diinginkan

	// Set check-in time
	att.CheckIn = time.Now()

	// Masukkan data attendance ke database
	_, err = database.DB.Exec(context.Background(), `
		INSERT INTO attendance (user_id, check_in, latitude, longitude, status) 
		VALUES ($1, $2, $3, $4, $5)`,
		att.UserID, att.CheckIn, att.Latitude, att.Longitude, att.Status)

	if err != nil {
		log.Println("Database error while inserting attendance:", err)
		http.Error(w, "Failed to record attendance", http.StatusInternalServerError)
		return
	}

	log.Println("Check in success")

	// Response sukses
	json.NewEncoder(w).Encode(map[string]string{"message": "Check in success"})
}

func CheckOut(w http.ResponseWriter, r *http.Request) {
	var att models.Attendance

	// Ambil UserID dari context
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "User ID is missing", http.StatusBadRequest)
		return
	}
	att.UserID = userID.(string) // Pastikan tipe data sesuai

	// Set waktu check-out
	att.CheckOut = time.Now()

	// Update database untuk check-out
	result, err := database.DB.Exec(context.Background(), `
		UPDATE attendance 
		SET check_out = $1 
		WHERE user_id = $2 AND check_out IS NULL`, att.CheckOut, att.UserID)

	if err != nil {
		log.Println("Database error while updating attendance:", err)
		http.Error(w, "Failed to record check-out", http.StatusInternalServerError)
		return
	}

	// Cek apakah ada baris yang ter-update
	rowsAffected := result.RowsAffected() // ✅ Hanya tangkap 1 nilai

	if rowsAffected == 0 {
		http.Error(w, "No active check-in found", http.StatusBadRequest)
		return
	}

	log.Println("Check out success")

	// Response sukses
	json.NewEncoder(w).Encode(map[string]string{"message": "Check out success"})
}

func GetMonthlyAttendance(w http.ResponseWriter, r *http.Request) {
    // Ambil user_id dari context (pastikan middleware sudah berjalan)
    userID := r.Context().Value("user_id")
    if userID == nil {
        http.Error(w, "User ID is missing", http.StatusUnauthorized)
        return
    }
    userIDStr, ok := userID.(string)
    if !ok {
        http.Error(w, "Invalid user ID", http.StatusInternalServerError)
        return
    }

    // Ambil parameter bulan dan tahun dari query string
    monthStr := r.URL.Query().Get("month")
    yearStr := r.URL.Query().Get("year")

    if monthStr == "" || yearStr == "" {
        http.Error(w, "Month and year are required", http.StatusBadRequest)
        return
    }

    // Konversi month & year ke integer
    month, err := strconv.Atoi(monthStr)
    if err != nil || month < 1 || month > 12 {
        http.Error(w, "Invalid month", http.StatusBadRequest)
        return
    }
    
    year, err := strconv.Atoi(yearStr)
    if err != nil || year < 2000 || year > 2100 {
        http.Error(w, "Invalid year", http.StatusBadRequest)
        return
    }

    // Ambil data dari database
    query := `
        SELECT id, user_id, check_in, check_out, latitude, longitude, status
    	FROM attendance
    	WHERE user_id = $1 AND EXTRACT(MONTH FROM check_in) = $2 AND EXTRACT(YEAR FROM check_in) = $3
    	ORDER BY check_in ASC`

    rows, err := database.DB.Query(context.Background(), query, userIDStr, month, year)
    if err != nil {
        http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var attendances []models.Attendance
    for rows.Next() {
        var att models.Attendance
        err := rows.Scan(&att.ID, &att.UserID, &att.CheckIn, &att.CheckOut, &att.Latitude, &att.Longitude, &att.Status)
        if err != nil {
            http.Error(w, "Error scanning data", http.StatusInternalServerError)
            return
        }
        attendances = append(attendances, att)
    }

    // Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(attendances)
}