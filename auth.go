package routes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"project_backend/database"
	"project_backend/models"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return // Exit early if preflight request
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Println("Received Password:", user.Password)

	// Basic validation
	if user.RegisterNo == "" || user.FullName == "" || user.Email == "" || user.Password == "" || user.Username == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Insert user into database (without hashing)
	_, err = database.DB.Exec("INSERT INTO users (register_no, full_name, email, password, username) VALUES (?, ?, ?, ?, ?)",
		user.RegisterNo, user.FullName, user.Email, user.Password, user.Username)
	if err != nil {
		log.Println("Database insert error:", err)
		http.Error(w, "Error signing up", http.StatusInternalServerError)
		return
	}

	// Success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Signup successful"})
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var admin models.Admin
	err := json.NewDecoder(r.Body).Decode(&admin)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var dbAdmin models.Admin
	err = database.DB.QueryRow("SELECT register_no, password FROM admins WHERE register_no = ?", admin.RegisterNo).
		Scan(&dbAdmin.RegisterNo, &dbAdmin.Password)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Println("Error querying admin:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Directly compare passwords (since it's stored as plain text)
	if dbAdmin.Password != admin.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

// JwtKey should be stored securely, preferably in an environment variable

var JwtKey = []byte("mySuperSecretKey123!@#") // Make sure it's capitalized for export

// Ensure this is consistent across files

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var dbUser models.User
	err = database.DB.QueryRow("SELECT register_no, full_name, email, username, password FROM users WHERE register_no = ?", user.RegisterNo).
		Scan(&dbUser.RegisterNo, &dbUser.FullName, &dbUser.Email, &dbUser.Username, &dbUser.Password)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Compare passwords (Consider using bcrypt for security)
	if dbUser.Password != user.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	claims := &models.Claims{                        // Use Claims from models.go
		RegisterNo: dbUser.RegisterNo,
		Username:   dbUser.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey) // Use JwtKey from auth.go
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
