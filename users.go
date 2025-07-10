package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"project_backend/database"
	"project_backend/models"
	"strings"
	"time"
)

// AddItem handles adding a new item (Lost/Found)

// AddLostItem handles lost item submissions with image uploads
func AddLostItem(w http.ResponseWriter, r *http.Request) {
	// Limit file size (5MB max)
	r.ParseMultipartForm(5 << 20)

	// Parse form data
	itemName := r.FormValue("item_name")
	description := r.FormValue("description")
	placeLost := r.FormValue("place")
	category := r.FormValue("category")
	userID := r.FormValue("user_id") // This links the item to the user

	// Parse date-time
	dateTimeLost, err := time.Parse("2006-01-02 15:04:05", r.FormValue("datetime"))
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Handle image upload
	file, handler, err := r.FormFile("image_url")
	var imageURL string

	if err == nil {
		defer file.Close()

		// Create folder if not exists
		imageFolder := "uploads"
		os.MkdirAll(imageFolder, os.ModePerm)

		// Generate unique filename
		ext := filepath.Ext(handler.Filename)
		imageName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		imagePath := filepath.Join(imageFolder, imageName)

		// Save file locally
		dst, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)

		// Generate Image URL (relative to backend)
		imageURL = "/uploads/" + imageName
	}

	// Create lost item struct
	item := models.Item{
		ItemName:    itemName,
		Description: description,
		DateTime:    dateTimeLost,
		Place:       placeLost,
		Category:    category,
		ImageURL:    imageURL,
		UserID:      userID,
	}

	// Insert into database
	_, err = database.DB.Exec(
		"INSERT INTO lost_items (item_name, description, datetime, place, category, image_url, user_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		item.ItemName, item.Description, item.DateTime, item.Place, item.Category, item.ImageURL, item.UserID,
	)

	if err != nil {
		http.Error(w, "Error adding item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Item added successfully",
		"data":    item,
	})
}

func AddFoundItem(w http.ResponseWriter, r *http.Request) {
	// Limit file size (5MB max)
	r.ParseMultipartForm(5 << 20)

	// Parse form data
	itemName := r.FormValue("item_name")
	description := r.FormValue("description")
	placeLost := r.FormValue("place")
	category := r.FormValue("category")
	userID := r.FormValue("user_id") // This links the item to the user

	// Parse date-time
	dateTimeLost, err := time.Parse("2006-01-02 15:04:05", r.FormValue("datetime"))
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Handle image upload
	file, handler, err := r.FormFile("image_url")
	var imageURL string

	if err == nil {
		defer file.Close()

		// Create folder if not exists
		imageFolder := "uploads"
		os.MkdirAll(imageFolder, os.ModePerm)

		// Generate unique filename
		ext := filepath.Ext(handler.Filename)
		imageName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		imagePath := filepath.Join(imageFolder, imageName)

		// Save file locally
		dst, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)

		// Generate Image URL (relative to backend)
		imageURL = "/uploads/" + imageName
	}

	// Create lost item struct
	item := models.Item{
		ItemName:    itemName,
		Description: description,
		DateTime:    dateTimeLost,
		Place:       placeLost,
		Category:    category,
		ImageURL:    imageURL,
		UserID:      userID,
	}

	// Insert into database
	_, err = database.DB.Exec(
		"INSERT INTO found_items (item_name, description, datetime, place, category, image_url, user_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		item.ItemName, item.Description, item.DateTime, item.Place, item.Category, item.ImageURL, item.UserID,
	)

	if err != nil {
		http.Error(w, "Error adding item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Item added successfully",
		"data":    item,
	})
}

// GetFeed handles fetching all lost/found items
// Response struct for properly formatted data

func GetLostFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query("SELECT id, item_name, description, datetime, place, category, image_url, user_id FROM lost_items")
	if err != nil {
		http.Error(w, "Error fetching feed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.ItemResponse
	for rows.Next() {
		var item models.Item
		var dateTimeStr string // Scan datetime as a string first

		err := rows.Scan(&item.ID, &item.ItemName, &item.Description, &dateTimeStr, &item.Place, &item.Category, &item.ImageURL, &item.UserID)
		if err != nil {
			http.Error(w, "Error processing feed data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert string to time.Time
		item.DateTime, err = time.Parse("2006-01-02 15:04:05", dateTimeStr)
		if err != nil {
			http.Error(w, "Error parsing datetime: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert time.Time to string format for JSON response
		formattedDateTime := item.DateTime.Format("2006-01-02 15:04:05")

		// Ensure the image URL is correctly formatted
		if item.ImageURL == "" {
			item.ImageURL = "http://localhost:8080/uploads/default_image.png" // Provide a default image
		} else if !strings.HasPrefix(item.ImageURL, "http") {
			item.ImageURL = "http://localhost:8080" + item.ImageURL
		}
		// Append to response slice
		items = append(items, models.ItemResponse{
			ID:          item.ID,
			ItemName:    item.ItemName,
			Description: item.Description,
			DateTime:    formattedDateTime, // Return as string
			Place:       item.Place,
			Category:    item.Category,
			ImageURL:    item.ImageURL,
			UserID:      item.UserID,
		})
	}

	// Send JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

func GetFoundFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := database.DB.Query("SELECT id, item_name, description, datetime, place, category, image_url, user_id FROM found_items")
	if err != nil {
		http.Error(w, "Error fetching feed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.ItemResponse
	for rows.Next() {
		var item models.Item
		var dateTimeStr string // Scan datetime as a string first

		err := rows.Scan(&item.ID, &item.ItemName, &item.Description, &dateTimeStr, &item.Place, &item.Category, &item.ImageURL, &item.UserID)
		if err != nil {
			http.Error(w, "Error processing feed data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert string to time.Time
		item.DateTime, err = time.Parse("2006-01-02 15:04:05", dateTimeStr)
		if err != nil {
			http.Error(w, "Error parsing datetime: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert time.Time to string format for JSON response
		formattedDateTime := item.DateTime.Format("2006-01-02 15:04:05")

		// Ensure the image URL is correctly formatted
		if item.ImageURL == "" {
			item.ImageURL = "http://localhost:8080/uploads/default_image.png" // Provide a default image
		} else if !strings.HasPrefix(item.ImageURL, "http") {
			item.ImageURL = "http://localhost:8080" + item.ImageURL
		}
		// Append to response slice
		items = append(items, models.ItemResponse{
			ID:          item.ID,
			ItemName:    item.ItemName,
			Description: item.Description,
			DateTime:    formattedDateTime, // Return as string
			Place:       item.Place,
			Category:    item.Category,
			ImageURL:    item.ImageURL,
			UserID:      item.UserID,
		})
	}

	// Send JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

func MyListings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from JWT token
	userID, err := ExtractUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var itemsLost []models.ItemResponse
	var itemsFound []models.ItemResponse

	// Fetch lost items
	lostQuery := "SELECT id, item_name, description, datetime, place, category, image_url, user_id FROM lost_items WHERE user_id = ?"
	lostRows, err := database.DB.Query(lostQuery, userID)
	if err != nil {
		http.Error(w, "Error fetching lost items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer lostRows.Close()

	for lostRows.Next() {
		var item models.Item
		var dateTimeStr string

		err := lostRows.Scan(&item.ID, &item.ItemName, &item.Description, &dateTimeStr, &item.Place, &item.Category, &item.ImageURL, &item.UserID)
		if err != nil {
			http.Error(w, "Error processing lost items: "+err.Error(), http.StatusInternalServerError)
			return
		}

		item.DateTime, err = time.Parse("2006-01-02 15:04:05", dateTimeStr)
		if err != nil {
			http.Error(w, "Error parsing datetime: "+err.Error(), http.StatusInternalServerError)
			return
		}

		formattedDateTime := item.DateTime.Format("2006-01-02 15:04:05")

		// Ensure image URL
		imageURL := item.ImageURL
		if imageURL == "" {
			imageURL = "http://localhost:8080/uploads/default_image.png"
		} else if !strings.HasPrefix(imageURL, "http") {
			imageURL = "http://localhost:8080" + imageURL
		}

		itemsLost = append(itemsLost, models.ItemResponse{
			ID:          item.ID,
			ItemName:    item.ItemName,
			Description: item.Description,
			DateTime:    formattedDateTime,
			Place:       item.Place,
			Category:    item.Category,
			ImageURL:    imageURL,
			UserID:      item.UserID,
		})
	}

	// Fetch found items
	foundQuery := "SELECT id, item_name, description, datetime, place, category, image_url, user_id FROM found_items WHERE user_id = ?"
	foundRows, err := database.DB.Query(foundQuery, userID)
	if err != nil {
		http.Error(w, "Error fetching found items: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer foundRows.Close()

	for foundRows.Next() {
		var item models.Item
		var dateTimeStr string

		err := foundRows.Scan(&item.ID, &item.ItemName, &item.Description, &dateTimeStr, &item.Place, &item.Category, &item.ImageURL, &item.UserID)
		if err != nil {
			http.Error(w, "Error processing found items: "+err.Error(), http.StatusInternalServerError)
			return
		}

		item.DateTime, err = time.Parse("2006-01-02 15:04:05", dateTimeStr)
		if err != nil {
			http.Error(w, "Error parsing datetime: "+err.Error(), http.StatusInternalServerError)
			return
		}

		formattedDateTime := item.DateTime.Format("2006-01-02 15:04:05")

		// Ensure image URL
		imageURL := item.ImageURL
		if imageURL == "" {
			imageURL = "http://localhost:8080/uploads/default_image.png"
		} else if !strings.HasPrefix(imageURL, "http") {
			imageURL = "http://localhost:8080" + imageURL
		}

		itemsFound = append(itemsFound, models.ItemResponse{
			ID:          item.ID,
			ItemName:    item.ItemName,
			Description: item.Description,
			DateTime:    formattedDateTime,
			Place:       item.Place,
			Category:    item.Category,
			ImageURL:    imageURL,
			UserID:      item.UserID,
		})
	}

	// Send response as JSON with two lists
	response := map[string]interface{}{
		"itemsLost":  itemsLost,
		"itemsFound": itemsFound,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
