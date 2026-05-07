package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
)

func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query("SELECT id, username, COALESCE(avatar_url, ''), points FROM users ORDER BY points DESC LIMIT 10")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.AvatarURL, &u.Points)
		users = append(users, u)
	}
	json.NewEncoder(w).Encode(users)
}
