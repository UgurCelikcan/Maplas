package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func PlacesHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method == "GET" {
		authHeader := r.Header.Get("Authorization")
		var userID int
		if authHeader != "" {
			tokenStr := auth.ExtractToken(authHeader)
			claims, err := auth.ValidateToken(tokenStr)
			if err == nil {
				db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
			}
		}

		latStr := r.URL.Query().Get("lat")
		lngStr := r.URL.Query().Get("lng")
		radiusStr := r.URL.Query().Get("radius")

		query := `
			SELECT p.id, p.name, p.description, p.lat, p.lng, p.category, p.city, COALESCE(p.image_url, '') as image_url, p.status,
			EXISTS(SELECT 1 FROM favorites f WHERE f.place_id = p.id AND f.user_id = $1) as is_favorite
			FROM places p WHERE p.status = 'approved'`

		args := []interface{}{userID}
		if latStr != "" && lngStr != "" && radiusStr != "" {
			query = `
				SELECT id, name, description, lat, lng, category, city, image_url, status, is_favorite
				FROM (
					SELECT p.*, (6371 * acos(cos(radians($2)) * cos(radians(lat)) * cos(radians(lng) - radians($3)) + sin(radians($2)) * sin(radians(lat)))) AS distance,
					EXISTS(SELECT 1 FROM favorites f WHERE f.place_id = p.id AND f.user_id = $1) as is_favorite
					FROM places p WHERE status = 'approved'
				) AS p WHERE distance < $4 ORDER BY distance ASC`
			args = append(args, latStr, lngStr, radiusStr)
		} else {
			query += " ORDER BY id DESC"
		}

		rows, err := db.DB.Query(query, args...)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		places := []models.Place{}
		for rows.Next() {
			var p models.Place
			var nameJSON, descJSON []byte
			if err := rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status, &p.IsFavorite); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			json.Unmarshal(nameJSON, &p.Name)
			json.Unmarshal(descJSON, &p.Description)
			places = append(places, p)
		}
		json.NewEncoder(w).Encode(places)
	} else if r.Method == "POST" {
		authHeader := r.Header.Get("Authorization")
		var creatorID int
		if authHeader != "" {
			tokenStr := auth.ExtractToken(authHeader)
			claims, err := auth.ValidateToken(tokenStr)
			if err == nil {
				db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&creatorID)
			}
		}
		var pr models.PlaceRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}

		// Normalize City Name (Title Case with Turkish support)
		pr.City = cases.Title(language.Turkish).String(pr.City)

		// --- AI AUTO-VERIFICATION ---
		aiApproved, aiReason := utils.ValidatePlaceWithAI(pr.Name, pr.Description, pr.Category, pr.City)
		status := "pending"
		if aiApproved {
			status = "approved"
			log.Printf("AI Auto-Approved: %s (%s)", pr.Name, aiReason)
		} else {
			status = "rejected"
			log.Printf("AI Auto-Rejected: %s (%s)", pr.Name, aiReason)
		}
		// ----------------------------

		nameMap := db.TranslateContent(pr.Name)
		descMap := db.TranslateContent(pr.Description)
		nameJSON, _ := json.Marshal(nameMap)
		descJSON, _ := json.Marshal(descMap)
		var id int
		var err error
		if creatorID > 0 {
			err = db.DB.QueryRow("INSERT INTO places (name, description, lat, lng, category, city, image_url, status, creator_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id", string(nameJSON), string(descJSON), pr.Lat, pr.Lng, pr.Category, pr.City, pr.ImageURL, status, creatorID).Scan(&id)
			// Award Points (+50 XP) if AI approved
			if err == nil && status == "approved" {
				db.DB.Exec("UPDATE users SET points = points + 50 WHERE id = $1", creatorID)
			}
		} else {
			err = db.DB.QueryRow("INSERT INTO places (name, description, lat, lng, category, city, image_url, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", string(nameJSON), string(descJSON), pr.Lat, pr.Lng, pr.Category, pr.City, pr.ImageURL, status).Scan(&id)
		}
		if err != nil {
			log.Printf("Error inserting place: %v", err)
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		p := models.Place{ID: id, Name: nameMap, Description: descMap, Lat: pr.Lat, Lng: pr.Lng, Category: pr.Category, City: pr.City, ImageURL: pr.ImageURL, Status: status}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	} else if r.Method == "PUT" {
		var pr models.PlaceRequest
		json.NewDecoder(r.Body).Decode(&pr)
		json.NewEncoder(w).Encode(map[string]string{"status": "Update not fully implemented in multi-language mode yet"})
	}
}
