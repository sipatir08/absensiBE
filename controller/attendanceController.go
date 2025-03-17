package controller

import (
	"context"
	"encoding/json"
	"log"
	"math"
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
	att.UserID = userID.(string)

	// Lokasi kantor
	officeLat := -6.876771186974438
	officeLon := 107.57603549878579
	maxDistance := 100.0 // dalam meter

	// Hitung jarak user ke lokasi kantor
	distance := HaversineDistance(att.Latitude, att.Longitude, officeLat, officeLon)

	// Cek apakah dalam radius
	if distance > maxDistance {
		http.Error(w, "You are too far from the allowed location", http.StatusForbidden)
		return
	}

	// Set check-in time
	att.CheckIn = time.Now()

	// Masukkan data ke attendance dan dapatkan ID
	var attendanceID string
	err = database.DB.QueryRow(context.Background(), `
		INSERT INTO attendance (user_id, check_in, latitude, longitude, status) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		att.UserID, att.CheckIn, att.Latitude, att.Longitude, att.Status,
	).Scan(&attendanceID)

	if err != nil {
		log.Println("Database error while inserting attendance:", err)
		http.Error(w, "Failed to record attendance", http.StatusInternalServerError)
		return
	}

	// Tambahkan log lokasi
	_, err = database.DB.Exec(context.Background(), `
		INSERT INTO attendance_logs (attendance_id, latitude, longitude) 
		VALUES ($1, $2, $3)`, attendanceID, att.Latitude, att.Longitude)

	if err != nil {
		log.Println("Error inserting log:", err)
	}

	log.Println("Check-in success")
	json.NewEncoder(w).Encode(map[string]string{"message": "Check-in success"})
}


func CheckOut(w http.ResponseWriter, r *http.Request) {
	var att models.Attendance

	// Ambil UserID dari context
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "User ID is missing", http.StatusBadRequest)
		return
	}
	att.UserID = userID.(string)

	// Ambil lokasi dari request body
	err := json.NewDecoder(r.Body).Decode(&att)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Lokasi kantor
	officeLat := -6.876771186974438
	officeLon := 107.57603549878579
	maxDistance := 100.0

	// Hitung jarak lokasi user
	distance := HaversineDistance(att.Latitude, att.Longitude, officeLat, officeLon)

	// Jika terlalu jauh, tolak check-out
	if distance > maxDistance {
		http.Error(w, "You are too far from the allowed location", http.StatusForbidden)
		return
	}

	// Ambil ID attendance yang masih aktif
	var attendanceID string
	err = database.DB.QueryRow(context.Background(), `
		SELECT id FROM attendance 
		WHERE user_id = $1 AND check_out IS NULL`, att.UserID).Scan(&attendanceID)

	if err != nil {
		log.Println("Error fetching active attendance:", err)
		http.Error(w, "No active check-in found", http.StatusBadRequest)
		return
	}

	// Set waktu check-out
	att.CheckOut = time.Now()

	// Update check-out di attendance
	_, err = database.DB.Exec(context.Background(), `
		UPDATE attendance 
		SET check_out = $1 
		WHERE id = $2`, att.CheckOut, attendanceID)

	if err != nil {
		log.Println("Database error while updating attendance:", err)
		http.Error(w, "Failed to record check-out", http.StatusInternalServerError)
		return
	}

	// Tambahkan log lokasi
	_, err = database.DB.Exec(context.Background(), `
		INSERT INTO attendance_logs (attendance_id, latitude, longitude) 
		VALUES ($1, $2, $3)`, attendanceID, att.Latitude, att.Longitude)

	if err != nil {
		log.Println("Error inserting log:", err)
	}

	log.Println("Check-out success")
	json.NewEncoder(w).Encode(map[string]string{"message": "Check-out success"})
}


// HaversineDistance menghitung jarak antara dua titik koordinat dalam meter
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Radius bumi dalam meter

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c // Jarak dalam meter
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

func GetAllUsersMonthlyAttendance(w http.ResponseWriter, r *http.Request) {
	// Ambil paramneter bulan dan tahun dari query string
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

	// Query untuk mengambil data kehadiran semua user dalam bulan & tahun tertentu
    query := `
        SELECT id, user_id, check_in, check_out, latitude, longitude, status
        FROM attendance
        WHERE EXTRACT(MONTH FROM check_in) = $1 AND EXTRACT(YEAR FROM check_in) = $2
        ORDER BY check_in ASC`

		rows, err := database.DB.Query(context.Background(), query, month, year)
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

func GetAttendanceLogs(w http.ResponseWriter, r *http.Request) {
	// Ambil user_id dari context
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "User ID is missing", http.StatusUnauthorized)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
		return
	}

	// Query untuk mengambil data log berdasarkan user_id
	query := `
        SELECT al.id::TEXT, al.attendance_id, al.latitude, al.longitude, al.created_at
        FROM attendance_logs al
        JOIN attendance a ON al.attendance_id = a.id
        WHERE a.user_id = $1
        ORDER BY al.created_at DESC
    `

	rows, err := database.DB.Query(context.Background(), query, userIDStr)
	if err != nil {
		log.Println("Failed to fetch logs:", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []models.AttendanceLog

	for rows.Next() {
		var logEntry models.AttendanceLog

		// Karena `al.id::TEXT`, kita pakai string langsung
		err := rows.Scan(&logEntry.ID, &logEntry.AttendanceID, &logEntry.Latitude, &logEntry.Longitude, &logEntry.CreatedAt)
		if err != nil {
			log.Println("Error scanning log data:", err)
			http.Error(w, "Error scanning log data", http.StatusInternalServerError)
			return
		}

		logs = append(logs, logEntry)
	}

	// Cek jika ada error saat iterasi rows
	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		http.Error(w, "Error processing logs", http.StatusInternalServerError)
		return
	}

	// Set response header dan encode hasil ke JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}