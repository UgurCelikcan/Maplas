package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AdminOnly(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		if r.Method == "GET" && action == "stats" {
			stats := make(map[string]interface{})
			var totalPlaces, pendingPlaces, totalUsers, totalComments int
			db.DB.QueryRow("SELECT COUNT(*) FROM places").Scan(&totalPlaces)
			db.DB.QueryRow("SELECT COUNT(*) FROM places WHERE status = 'pending'").Scan(&pendingPlaces)
			db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
			db.DB.QueryRow("SELECT COUNT(*) FROM comments").Scan(&totalComments)
			stats["total_places"] = totalPlaces
			stats["pending_places"] = pendingPlaces
			stats["total_users"] = totalUsers
			stats["total_comments"] = totalComments
			rows, err := db.DB.Query("SELECT category, COUNT(*) FROM places GROUP BY category")
			categories := make(map[string]int)
			if err == nil {
				for rows.Next() {
					var cat string
					var count int
					rows.Scan(&cat, &count)
					categories[cat] = count
				}
				rows.Close()
			}
			stats["categories"] = categories
			json.NewEncoder(w).Encode(stats)
			return
		}
		if r.Method == "GET" && action == "users" {
			rows, err := db.DB.Query("SELECT id, username, role FROM users ORDER BY id ASC")
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			users := []models.User{}
			for rows.Next() {
				var u models.User
				if err := rows.Scan(&u.ID, &u.Username, &u.Role); err != nil {
					continue
				}
				users = append(users, u)
			}
			json.NewEncoder(w).Encode(users)
			return
		}
		if r.Method == "GET" && action == "pending" {
			rows, err := db.DB.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, '') as image_url, status FROM places WHERE status = 'pending' ORDER BY id DESC")
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
					log.Printf("Admin pending scan error: %v", err)
					continue
				}
				json.Unmarshal(nameJSON, &p.Name)
				json.Unmarshal(descJSON, &p.Description)
				places = append(places, p)
			}
			json.NewEncoder(w).Encode(places)
			return
		}
		if r.Method == "POST" && (action == "approve" || action == "reject") {
			var req struct {
				ID int `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid body", http.StatusBadRequest)
				return
			}
			if action == "approve" {
				_, err := db.DB.Exec("UPDATE places SET status = 'approved' WHERE id = $1", req.ID)
				if err != nil {
					log.Printf("Approve error: %v", err)
					http.Error(w, "DB error", http.StatusInternalServerError)
					return
				}
			} else {
				_, err := db.DB.Exec("DELETE FROM places WHERE id = $1", req.ID)
				if err != nil {
					log.Printf("Reject error: %v", err)
					http.Error(w, "DB error", http.StatusInternalServerError)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
		}
	})(w, r)
}
