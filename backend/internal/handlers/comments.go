package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
)

func CommentsHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method == "GET" {
		placeID := r.URL.Query().Get("place_id")
		rows, _ := db.DB.Query("SELECT id, place_id, content, rating, created_at FROM comments WHERE place_id = $1 ORDER BY created_at DESC", placeID)
		defer rows.Close()
		comments := []models.Comment{}
		for rows.Next() {
			var c models.Comment
			rows.Scan(&c.ID, &c.PlaceID, &c.Content, &c.Rating, &c.CreatedAt)
			comments = append(comments, c)
		}
		json.NewEncoder(w).Encode(comments)
	} else if r.Method == "POST" {
		authHeader := r.Header.Get("Authorization")
		var userID int
		if authHeader != "" {
			tokenStr := auth.ExtractToken(authHeader)
			claims, err := auth.ValidateToken(tokenStr)
			if err == nil {
				db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
			}
		}
		var c models.Comment
		json.NewDecoder(r.Body).Decode(&c)
		if userID > 0 {
			db.DB.QueryRow("INSERT INTO comments (place_id, content, rating, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating, userID).Scan(&c.ID, &c.CreatedAt)
			// Award Points (+10 XP)
			db.DB.Exec("UPDATE users SET points = points + 10 WHERE id = $1", userID)
		} else {
			db.DB.QueryRow("INSERT INTO comments (place_id, content, rating) VALUES ($1, $2, $3) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating).Scan(&c.ID, &c.CreatedAt)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}
