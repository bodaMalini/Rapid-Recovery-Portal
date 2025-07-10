package routes

import (
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterRoutes initializes all the routes
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/signup", SignUp).Methods("POST")
	r.HandleFunc("/login/user", Login).Methods("POST")
	r.HandleFunc("/login/admin", AdminLogin).Methods("POST")
	r.HandleFunc("/lost_items", AddLostItem).Methods("POST")
	r.HandleFunc("/found_items", AddFoundItem).Methods("POST")
	r.HandleFunc("/lost_feed", GetLostFeed).Methods("GET")
	r.HandleFunc("/found_feed", GetFoundFeed).Methods("GET")
	r.HandleFunc("/my_listings", MyListings).Methods("GET") // Added MyListings route
	r.HandleFunc("/admin/lost/return/{id}", LostMoveToReturned).Methods("POST")
	r.HandleFunc("/admin/found/return/{id}", FoundMoveToReturned).Methods("POST")
	r.HandleFunc("/returned_feed", ReturnedFeed).Methods("GET")
	r.HandleFunc("/upload", UploadImage).Methods("POST")
}

// HomeHandler provides a simple hello world message to check the server status
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World - API is running"))
}
