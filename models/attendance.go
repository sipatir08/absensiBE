package models

import "time"

type Attendance struct {
	ID        	string    	`json:"id"`
	UserID    	string    	`json:"user_id"`
	CheckIn 	time.Time 	`json:"check_in"`
	CheckOut 	time.Time 	`json:"check_out"`
	Latitude 	float64 	`json:"latitude"`
	Longitude 	float64 	`json:"longitude"`
	Status   	string    	`json:"status"`
}