package models

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// Admin represents the admin model
type Admin struct {
	RegisterNo string `json:"register_no"` // Admin's email
	Password   string `json:"password"`    // Admin's password
}

type User struct {
	RegisterNo string `json:"register_no"` // The register number used for login
	Password   string `json:"password"`    // Password used for login
	FullName   string `json:"full_name"`   // Full name of the user (stored in DB)
	Email      string `json:"email"`
	Username   string `json:"username"` // User's email
}
type Claims struct {
	RegisterNo string `json:"register_no"`
	Username   string `json:"username"`
	jwt.RegisteredClaims
}
type Item struct {
	ID          int       `json:"id"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	DateTime    time.Time `json:"datetime"` // Combined Date & Time
	Place       string    `json:"place"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	UserID      string    `json:"user_id"`
}
type ItemResponse struct {
	ID          int    `json:"id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	DateTime    string `json:"datetime"`
	Place       string `json:"place"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
	UserID      string `json:"user_id"`
}
