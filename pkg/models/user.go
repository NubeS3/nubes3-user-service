package models

import "time"

type User struct {
	Id        string    `json:"id"`
	Firstname string    `json:"firstname" binding:"required"`
	Lastname  string    `json:"lastname" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Pass      string    `json:"password" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	Dob       time.Time `json:"dob" binding:"required"`
	Company   string    `json:"company" binding:"required"`
	Gender    bool      `json:"gender" binding:"required"`
	IsActive  bool      `json:"is_active,omitempty"`
	IsBanned  bool      `json:"is_banned,omitempty"`
	// DB Info
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type UserRes struct {
	Firstname string    `json:"firstname" binding:"required"`
	Lastname  string    `json:"lastname" binding:"required"`
	Username  string    `json:"username" binding:"required"`
	Pass      string    `json:"password" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	Dob       time.Time `json:"dob" binding:"required"`
	Company   string    `json:"company" binding:"required"`
	Gender    bool      `json:"gender" binding:"required"`
}
