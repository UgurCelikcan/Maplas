package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}
	tokenStr := auth.ExtractToken(authHeader)
	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	var userID int
	err = db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	action := r.URL.Query().Get("action")
	if r.Method == "GET" {
		if action == "places" {
			rows, _ := db.DB.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, ''), status FROM places WHERE creator_id = $1 ORDER BY id DESC", userID)
			defer rows.Close()
			var places []models.Place
			for rows.Next() {
				var p models.Place
				var nameJSON, descJSON []byte
				rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status)
				json.Unmarshal(nameJSON, &p.Name)
				json.Unmarshal(descJSON, &p.Description)
				places = append(places, p)
			}
			json.NewEncoder(w).Encode(places)
			return
		}
		if action == "comments" {
			rows, _ := db.DB.Query("SELECT c.id, c.content, c.rating, c.created_at, p.id, p.name FROM comments c JOIN places p ON c.place_id = p.id WHERE c.user_id = $1 ORDER BY c.created_at DESC", userID)
			defer rows.Close()
			var results []map[string]interface{}
			for rows.Next() {
				var id, rating, placeID int
				var content, placeName string
				var createdAt time.Time
				rows.Scan(&id, &content, &rating, &createdAt, &placeID, &placeName)
				results = append(results, map[string]interface{}{"id": id, "content": content, "rating": rating, "created_at": createdAt, "place_id": placeID, "place_name": placeName})
			}
			json.NewEncoder(w).Encode(results)
			return
		}
		var u models.User
		err := db.DB.QueryRow("SELECT id, username, role, COALESCE(email, ''), COALESCE(bio, ''), COALESCE(avatar_url, ''), points FROM users WHERE id=$1", userID).Scan(&u.ID, &u.Username, &u.Role, &u.Email, &u.Bio, &u.AvatarURL, &u.Points)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(u)
	} else if r.Method == "PUT" {
		var u models.User
		json.NewDecoder(r.Body).Decode(&u)
		db.DB.Exec("UPDATE users SET email=$1, bio=$2, avatar_url=$3 WHERE id=$4", u.Email, u.Bio, u.AvatarURL, userID)
		u.Username = claims.Username
		u.Role = claims.Role
		json.NewEncoder(w).Encode(u)
	}
}
