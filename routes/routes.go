package routes

import (
	"absensi/controller"
	"absensi/middleware" // Pastikan middleware diimpor
	"github.com/gorilla/mux"
	"firebase.google.com/go/auth"
)

func SetupRoutes(client *auth.Client) *mux.Router {
	r := mux.NewRouter()

	// Routes untuk login dan register
	r.HandleFunc("/register", controller.Register).Methods("POST")
	r.HandleFunc("/login", controller.Login).Methods("POST")
	r.HandleFunc("/getUsers", controller.GetUsers).Methods("GET")
	r.HandleFunc("/updateUserRole/{id}", controller.UpdateUserRole).Methods("PUT")
	r.HandleFunc("/delete/{id}", controller.DeleteUser).Methods("DELETE")

	// Subrouter untuk endpoint yang memerlukan autentikasi JWT
	protected := r.PathPrefix("/api/protected").Subrouter()
	protected.Use(middleware.AuthMiddleware) // Middleware untuk autentikasi JWT

	// Routes untuk check-in dan check-out yang hanya bisa diakses jika autentikasi berhasil
	protected.HandleFunc("/check-in", controller.CheckIn).Methods("POST")
	protected.HandleFunc("/check-out", controller.CheckOut).Methods("POST")
	protected.HandleFunc("/attendance/monthly", controller.GetMonthlyAttendance).Methods("POST")
	protected.HandleFunc("/attendance/All-User", controller.GetAllUsersMonthlyAttendance).Methods("GET")
	protected.HandleFunc("/attendance/logs", controller.GetAttendanceLogs).Methods("GET")


	return r
}
