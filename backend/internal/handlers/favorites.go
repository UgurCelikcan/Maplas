package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
)

func FavoritesHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("Favorites: Missing authorization header")
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := auth.ExtractToken(authHeader)
	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		log.Printf("Favorites: Invalid token: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	var userID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
	if err != nil || userID == 0 {
		log.Printf("Favorites: User not found for username %s", claims.Username)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		rows, err := db.DB.Query(`
			SELECT p.id, p.name, p.description, p.lat, p.lng, p.category, p.city, COALESCE(p.image_url, '') as image_url, p.status 
			FROM places p 
			JOIN favorites f ON p.id = f.place_id 
			WHERE f.user_id = $1`, userID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		places := []models.Place{}
		for rows.Next() {
			var p models.Place
			var nameJSON, descJSON []byte
			if err := rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status); err != nil {
				log.Printf("Favorites scan error: %v", err)
				continue
			}
			json.Unmarshal(nameJSON, &p.Name)
			json.Unmarshal(descJSON, &p.Description)
			places = append(places, p)
		}
		json.NewEncoder(w).Encode(places)
	} else if r.Method == "POST" {
		var req struct {
			PlaceID int `json:"place_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Favorites POST: Invalid body: %v", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		_, err = db.DB.Exec("INSERT INTO favorites (user_id, place_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, req.PlaceID)
		if err != nil {
			log.Printf("Favorites POST: DB Error: %v", err)
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	} else if r.Method == "DELETE" {
		placeIDStr := r.URL.Query().Get("place_id")
		placeID, err := strconv.Atoi(placeIDStr)
		if err != nil {
			log.Printf("Favorites DELETE: Invalid place_id: %v", err)
			http.Error(w, "Invalid place ID", http.StatusBadRequest)
			return
		}

		_, err = db.DB.Exec("DELETE FROM favorites WHERE user_id = $1 AND place_id = $2", userID, placeID)
		if err != nil {
			log.Printf("Favorites DELETE: DB Error: %v", err)
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
