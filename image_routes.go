package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"project_backend/database"
	"strings"
)

// UploadImage handles image uploads and saves them locally
func UploadImage(w http.ResponseWriter, r *http.Request) {
	// Check if the request is multipart
	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form data
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check the file type (optional)
	if !strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
		http.Error(w, "Invalid file type. Only image files are allowed", http.StatusBadRequest)
		return
	}

	// Ensure uploads directory exists
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
			return
		}
	}

	// Define the file path where the image will be stored
	filePath := fmt.Sprintf("%s/%s", uploadDir, handler.Filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	// Copy the file to the local folder
	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	// Save the file path and title in the database
	// Assuming that title is passed along with the image upload request
	itemTitle := r.FormValue("title") // Assume the title is passed as form data
	if itemTitle == "" {
		itemTitle = "Lost Item" // Default if not provided
	}

	_, err = database.DB.Exec("INSERT INTO items (title, image_url) VALUES (?, ?)", itemTitle, filePath)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response with the file URL
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "File uploaded successfully",
		"url":     filePath,
	})
}
