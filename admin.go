package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"project_backend/models"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"project_backend/database"
)

// HandleMoveToReturned moves an item from lost_items to returned_items

// LostMoveToReturned moves a lost item to the returned items table
// LostMoveToReturned moves a lost item to the returned items table
func LostMoveToReturned(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get lost item ID from URL path
	vars := mux.Vars(r)
	lostItemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("Invalid lost item ID:", err)
		http.Error(w, `{"error": "Invalid lost item ID"}`, http.StatusBadRequest)
		return
	}

	// Parse found_item_id from request body
	var requestBody struct {
		FoundItemID int `json:"found_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Println("Failed to decode request body:", err)
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		http.Error(w, `{"error": "Failed to start transaction"}`, http.StatusInternalServerError)
		return
	}

	// Move lost item to returned_items table
	query := `
        INSERT INTO returned_items (item_name, description, place, category, datetime, user_id, image_url) 
        SELECT item_name, description, place, category, datetime, user_id, COALESCE(image_url, '') 
        FROM lost_items WHERE id = ?`
	_, err = tx.Exec(query, lostItemID)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert into returned_items:", err)
		http.Error(w, `{"error": "Failed to move lost item"}`, http.StatusInternalServerError)
		return
	}

	// Delete from lost_items
	_, err = tx.Exec("DELETE FROM lost_items WHERE id = ?", lostItemID)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to delete lost item:", err)
		http.Error(w, `{"error": "Failed to delete lost item"}`, http.StatusInternalServerError)
		return
	}

	// Delete from found_items
	_, err = tx.Exec("DELETE FROM found_items WHERE id = ?", requestBody.FoundItemID)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to delete found item:", err)
		http.Error(w, `{"error": "Failed to delete found item"}`, http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		http.Error(w, `{"error": "Failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	// Respond success
	response := map[string]string{
		"message": fmt.Sprintf("Lost item %d and found item %d moved to returned", lostItemID, requestBody.FoundItemID),
	}
	json.NewEncoder(w).Encode(response)
}

func FoundMoveToReturned(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"]) // Convert id to int
	if err != nil {
		log.Print("Invalid item ID: ", err)
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		log.Print("Failed to start transaction: ", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Insert into returned_items including `image_url`
	_, err = tx.Exec(`
        INSERT INTO returned_items (item_name, description, place, category, datetime, image_url, user_id) 
        SELECT item_name, description, place, category, datetime, COALESCE(image_url, ''), user_id 
        FROM found_items WHERE id = ?`, itemID)
	if err != nil {
		tx.Rollback()
		log.Print("Failed to move item: ", err)
		http.Error(w, "Failed to move item", http.StatusInternalServerError)
		return
	}

	// Delete from found_items after successful insert
	_, err = tx.Exec("DELETE FROM found_items WHERE id = ?", itemID)
	if err != nil {
		tx.Rollback()
		log.Print("Failed to delete item: ", err)
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Print("Failed to commit transaction: ", err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	log.Print("Item moved to returned items: ", itemID)
	fmt.Fprintf(w, "Item %d moved to returned items", itemID)
}

// ReturnedFeed - Fetches all returned items

func ReturnedFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query to get returned items
	rows, err := database.DB.Query("SELECT item_name, description, datetime, place, category, image_url, user_id FROM returned_items")
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Error fetching feed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.ItemResponse

	for rows.Next() {
		var item models.Item
		var dateTimeStr string

		if err := rows.Scan(&item.ItemName, &item.Description, &dateTimeStr, &item.Place, &item.Category, &item.ImageURL, &item.UserID); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Error processing feed data", http.StatusInternalServerError)
			return
		}

		// Handle datetime parsing
		formattedDateTime := "Unknown"
		if dateTimeStr != "" {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", dateTimeStr); err == nil {
				formattedDateTime = parsedTime.Format("2006-01-02 15:04:05")
			} else {
				log.Println("Datetime parsing error:", err)
			}
		}

		// Ensure image URL is properly formatted
		if item.ImageURL == "" {
			item.ImageURL = "http://localhost:8080/uploads/default_image.png"
		} else if !strings.HasPrefix(item.ImageURL, "http") {
			item.ImageURL = "http://localhost:8080" + item.ImageURL
		}

		// Append processed item to response slice
		items = append(items, models.ItemResponse{
			ItemName:    item.ItemName,
			Description: item.Description,
			DateTime:    formattedDateTime,
			Place:       item.Place,
			Category:    item.Category,
			ImageURL:    item.ImageURL,
			UserID:      item.UserID,
		})
	}

	if len(items) == 0 {
		log.Println("No returned items found.")
	}

	// Send JSON response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Error sending response", http.StatusInternalServerError)
	}
}
