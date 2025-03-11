package models

import "time"

// User model for users table
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"password"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}