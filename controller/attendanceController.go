package controller

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"absensi/database"
	"absensi/models"
	"absensi/utils"

	"github.com/jackc/pgx/v5"
)

func CheckIn(w http.ResponseWriter, r *http.Request) {
	// Ambil user_id dari context dengan aman
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parsing request body untuk mendapatkan latitude & longitude
	var requestData struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Cek apakah ada attendance record untuk user
	var attendanceID string
	err := database.DB.QueryRow(
		r.Context(),
		`SELECT id FROM attendance WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&attendanceID)

	// Jika tidak ada, buat attendance baru
	if err == pgx.ErrNoRows {
		query := `INSERT INTO attendance (user_id, created_at) VALUES ($1, NOW()) RETURNING id`
		err = database.DB.QueryRow(r.Context(), query, userID).Scan(&attendanceID)
		if err != nil {
			log.Println("Error creating new attendance record:", err)
			http.Error(w, "Failed to create attendance record", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Println("Error fetching attendance ID:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Simpan data check-in di attendance_logs
	query := `INSERT INTO attendance_logs (attendance_id, latitude, longitude, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`
	_, err = database.DB.Exec(r.Context(), query, attendanceID, requestData.Latitude, requestData.Longitude)
	if err != nil {
		log.Println("Error inserting check-in:", err)
		http.Error(w, "Failed to check-in", http.StatusInternalServerError)
		return
	}

	// Ambil email user untuk notifikasi
	var email string
	err = database.DB.QueryRow(r.Context(), "SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		log.Println("Error fetching user email:", err)
		http.Error(w, "Failed to retrieve user email", http.StatusInternalServerError)
		return
	}

	// Kirim notifikasi email
	err = utils.SendEmailNotification(email, "Check-in Berhasil", "Anda berhasil check-in hari ini!")
	if err != nil {
		log.Println("Error sending email:", err)
		http.Error(w, "Failed to send email notification", http.StatusInternalServerError)
		return
	}

	// Respon sukses
	json.NewEncoder(w).Encode(map[string]string{"message": "Check-in berhasil dan email telah dikirim!"})
}


func CheckOut(w http.ResponseWriter, r *http.Request) {
	// Ambil user_id dari context dengan aman
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parsing request body untuk mendapatkan latitude & longitude
	var requestData struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Ambil attendance_id berdasarkan user_id
	var attendanceID string
	err := database.DB.QueryRow(r.Context(), `SELECT id FROM attendance WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID).Scan(&attendanceID)
	if err != nil {
		log.Println("Error fetching attendance ID:", err)
		http.Error(w, "Attendance record not found", http.StatusNotFound)
		return
	}

	// Simpan data check-out di database
	query := `INSERT INTO attendance_logs (attendance_id, latitude, longitude, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`
	_, err = database.DB.Exec(r.Context(), query, attendanceID, requestData.Latitude, requestData.Longitude)
	if err != nil {
		log.Println("Error inserting check-out:", err)
		http.Error(w, "Failed to check-out", http.StatusInternalServerError)
		return
	}

	// Ambil email user dari database
	var email string
	err = database.DB.QueryRow(r.Context(), "SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		log.Println("Error fetching user email:", err)
		http.Error(w, "Failed to retrieve user email", http.StatusInternalServerError)
		return
	}

	// Kirim notifikasi email
	err = utils.SendEmailNotification(email, "Check-out Berhasil", "Anda berhasil check-out hari ini!")
	if err != nil {
		log.Println("Error sending email:", err)
		http.Error(w, "Failed to send email notification", http.StatusInternalServerError)
		return
	}

	// Beri response sukses
	json.NewEncoder(w).Encode(map[string]string{"message": "Check-out berhasil dan email telah dikirim!"})
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
	// Ambil user_id dari context dengan aman
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "User ID is missing", http.StatusUnauthorized)
		return
	}

	// Query untuk mengambil data check-in berdasarkan user_id
	query := `
        SELECT al.id::TEXT, al.attendance_id, al.latitude, al.longitude, al.created_at
        FROM attendance_logs al
        JOIN attendance a ON al.attendance_id = a.id
        WHERE a.user_id = $1
        ORDER BY al.created_at DESC
    `

	rows, err := database.DB.Query(r.Context(), query, userID)
	if err != nil {
		log.Println("Failed to fetch logs:", err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []models.AttendanceLog

	for rows.Next() {
		var logEntry models.AttendanceLog
		err := rows.Scan(&logEntry.ID, &logEntry.AttendanceID, &logEntry.Latitude, &logEntry.Longitude, &logEntry.CreatedAt)
		if err != nil {
			log.Println("Error scanning log data:", err)
			http.Error(w, "Error scanning log data", http.StatusInternalServerError)
			return
		}
		logs = append(logs, logEntry)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		http.Error(w, "Error processing logs", http.StatusInternalServerError)
		return
	}

	// Kirim response JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
	}
}
