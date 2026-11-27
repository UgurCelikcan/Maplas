package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/moderation"
)

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func CommentsHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method == "GET" {
		placeID, err := strconv.Atoi(r.URL.Query().Get("place_id"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid place_id")
			return
		}
		rows, err := db.DB.Query("SELECT id, place_id, content, rating, created_at FROM comments WHERE place_id = $1 ORDER BY created_at DESC", placeID)
		if err != nil {
			log.Printf("Comments query error: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()
		comments := []models.Comment{}
		for rows.Next() {
			var c models.Comment
			if err := rows.Scan(&c.ID, &c.PlaceID, &c.Content, &c.Rating, &c.CreatedAt); err != nil {
				log.Printf("Comments scan error: %v", err)
				continue
			}
			comments = append(comments, c)
		}
		json.NewEncoder(w).Encode(comments)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var userID int
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if claims, err := auth.ValidateToken(auth.ExtractToken(authHeader)); err == nil {
			db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
		}
	}

	var c models.Comment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	c.Content = strings.TrimSpace(c.Content)

	if check := moderation.CheckComment(c.Content, c.Rating); check.Decision == moderation.Reject {
		writeJSONError(w, http.StatusUnprocessableEntity, strings.Join(check.Reasons, " "))
		return
	}

	var placeExists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM places WHERE id = $1 AND status = 'approved')", c.PlaceID).Scan(&placeExists)
	if !placeExists {
		writeJSONError(w, http.StatusNotFound, "Yer bulunamadı.")
		return
	}

	// Points only for the first comment a user writes on a place (no farming).
	firstComment := false
	if userID > 0 {
		var already bool
		db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE user_id = $1 AND place_id = $2)", userID, c.PlaceID).Scan(&already)
		firstComment = !already
	}

	var err error
	if userID > 0 {
		err = db.DB.QueryRow("INSERT INTO comments (place_id, content, rating, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating, userID).Scan(&c.ID, &c.CreatedAt)
	} else {
		err = db.DB.QueryRow("INSERT INTO comments (place_id, content, rating) VALUES ($1, $2, $3) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating).Scan(&c.ID, &c.CreatedAt)
	}
	if err != nil {
		log.Printf("Comment insert error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if firstComment {
		db.DB.Exec("UPDATE users SET points = points + 10 WHERE id = $1", userID) // +10 XP
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}
