package models

import (
	"time"

	uuid "github.com/jackc/pgx/pgtype/ext/gofrs-uuid"
)

type AttendanceLog struct {
    ID           uuid.UUID `json:"id"`
    AttendanceID uuid.UUID `json:"attendance_id"`
    Latitude     float64   `json:"latitude"`
    Longitude    float64   `json:"longitude"`
    CreatedAt    time.Time `json:"created_at"`
}