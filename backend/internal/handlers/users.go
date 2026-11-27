package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/models"
	"golang.org/x/crypto/bcrypt"
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
	claims, err := auth.ValidateToken(auth.ExtractToken(authHeader))
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	var userID int
	if err := db.DB.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	action := r.URL.Query().Get("action")

	switch r.Method {
	case "GET":
		switch action {
		case "places":
			places, err := userPlaces(userID)
			if err != nil {
				serverError(w, "user places", err)
				return
			}
			json.NewEncoder(w).Encode(places)
		case "comments":
			comments, err := userComments(userID)
			if err != nil {
				serverError(w, "user comments", err)
				return
			}
			json.NewEncoder(w).Encode(comments)
		case "export":
			exportUserData(w, userID)
		default:
			u, err := userProfile(userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(u)
		}

	case "PUT":
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if _, err := db.DB.Exec("UPDATE users SET email=$1, bio=$2, avatar_url=$3 WHERE id=$4", u.Email, u.Bio, u.AvatarURL, userID); err != nil {
			serverError(w, "update profile", err)
			return
		}
		u.Username = claims.Username
		u.Role = claims.Role
		u.Password = ""
		json.NewEncoder(w).Encode(u)

	case "DELETE":
		deleteAccount(w, r, userID)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func serverError(w http.ResponseWriter, what string, err error) {
	log.Printf("%s error: %v", what, err)
	http.Error(w, "Database error", http.StatusInternalServerError)
}

func userProfile(userID int) (models.User, error) {
	var u models.User
	err := db.DB.QueryRow("SELECT id, username, role, COALESCE(email, ''), COALESCE(bio, ''), COALESCE(avatar_url, ''), points FROM users WHERE id=$1", userID).
		Scan(&u.ID, &u.Username, &u.Role, &u.Email, &u.Bio, &u.AvatarURL, &u.Points)
	return u, err
}

func userPlaces(userID int) ([]models.Place, error) {
	rows, err := db.DB.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, ''), status FROM places WHERE creator_id = $1 ORDER BY id DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	places := []models.Place{}
	for rows.Next() {
		var p models.Place
		var nameJSON, descJSON []byte
		if err := rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status); err != nil {
			log.Printf("user places scan error: %v", err)
			continue
		}
		json.Unmarshal(nameJSON, &p.Name)
		json.Unmarshal(descJSON, &p.Description)
		places = append(places, p)
	}
	return places, rows.Err()
}

func userComments(userID int) ([]map[string]interface{}, error) {
	rows, err := db.DB.Query("SELECT c.id, COALESCE(c.content, ''), COALESCE(c.rating, 0), c.created_at, p.id, COALESCE(p.name->>'tr', '') FROM comments c JOIN places p ON c.place_id = p.id WHERE c.user_id = $1 ORDER BY c.created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []map[string]interface{}{}
	for rows.Next() {
		var id, rating, placeID int
		var content, placeName string
		var createdAt time.Time
		if err := rows.Scan(&id, &content, &rating, &createdAt, &placeID, &placeName); err != nil {
			log.Printf("user comments scan error: %v", err)
			continue
		}
		results = append(results, map[string]interface{}{"id": id, "content": content, "rating": rating, "created_at": createdAt, "place_id": placeID, "place_name": placeName})
	}
	return results, rows.Err()
}

// exportUserData sends everything stored about the user as a JSON file (KVKK md. 11).
func exportUserData(w http.ResponseWriter, userID int) {
	profile, err := userProfile(userID)
	if err != nil {
		serverError(w, "export profile", err)
		return
	}
	places, err := userPlaces(userID)
	if err != nil {
		serverError(w, "export places", err)
		return
	}
	comments, err := userComments(userID)
	if err != nil {
		serverError(w, "export comments", err)
		return
	}
	favorites := []int{}
	rows, err := db.DB.Query("SELECT place_id FROM favorites WHERE user_id = $1 ORDER BY place_id", userID)
	if err != nil {
		serverError(w, "export favorites", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if rows.Scan(&id) == nil {
			favorites = append(favorites, id)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="maplas-verilerim.json"`)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(map[string]interface{}{
		"exported_at":        time.Now().UTC(),
		"profile":            profile,
		"places":             places,
		"comments":           comments,
		"favorite_place_ids": favorites,
		"note":               "Şifreniz yalnızca geri döndürülemez şekilde (bcrypt) saklanır ve bu dosyada yer almaz.",
	})
}

// deleteAccount permanently deletes the user after checking their password.
// Comments and favorites are deleted; places they added stay on the map without an owner.
func deleteAccount(w http.ResponseWriter, r *http.Request, userID int) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Hesabı silmek için şifrenizi girin.")
		return
	}
	var hash string
	if err := db.DB.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&hash); err != nil {
		serverError(w, "delete account lookup", err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeJSONError(w, http.StatusForbidden, "Şifre hatalı.")
		return
	}
	// comments and favorites are removed by ON DELETE CASCADE; places.creator_id becomes NULL.
	if _, err := db.DB.Exec("DELETE FROM users WHERE id = $1", userID); err != nil {
		serverError(w, "delete account", err)
		return
	}
	log.Printf("Account %d deleted at user request", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hesabınız silindi."})
}
