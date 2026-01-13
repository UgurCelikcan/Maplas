package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// --- Structs ---

type Place struct {
	ID          int               `json:"id"`
	Name        map[string]string `json:"name"`        // JSONB
	Description map[string]string `json:"description"` // JSONB
	Lat         float64           `json:"lat"`
	Lng         float64           `json:"lng"`
	Category    string            `json:"category"`
	City        string            `json:"city"`
	ImageURL    string            `json:"imageUrl"`
	Status      string            `json:"status"` // 'pending' or 'approved'
	IsFavorite  bool              `json:"is_favorite"`
}

type PlaceRequest struct {
	Name        string  `json:"name"` // Frontend sends string
	Description string  `json:"description"` // Frontend sends string
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Category    string  `json:"category"`
	City        string  `json:"city"`
	ImageURL    string  `json:"imageUrl"`
}

type Comment struct {
	ID        int       `json:"id"`
	PlaceID   int       `json:"place_id"`
	Content   string    `json:"content"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
	Points    int    `json:"points"`
}

type Credentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	SecretCode string `json:"secret_code,omitempty"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// --- Globals ---

var db *sql.DB
var jwtKey = []byte(getEnv("JWT_SECRET", "my_super_secret_key_2026")) // Fallback for dev only
var AdminSecretCode = getEnv("ADMIN_SECRET", "Maplas-2026") // Fallback for dev only

// --- Helpers ---

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func initDB() {
	var err error
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "places_db")
	dbPort := getEnv("DB_PORT", "5432")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to DB, retrying... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// Create tables with JSONB support
	createTables := `
	CREATE TABLE IF NOT EXISTS places (
		id SERIAL PRIMARY KEY,
		name JSONB NOT NULL,
		description JSONB,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		category TEXT,
		city TEXT,
		image_url TEXT,
		status TEXT DEFAULT 'pending',
		creator_id INT REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		place_id INT REFERENCES places(id) ON DELETE CASCADE,
		content TEXT,
		rating INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		user_id INT REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'user'
	);

	CREATE TABLE IF NOT EXISTS favorites (
		user_id INT REFERENCES users(id) ON DELETE CASCADE,
		place_id INT REFERENCES places(id) ON DELETE CASCADE,
		PRIMARY KEY (user_id, place_id)
	);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	
	// Migration: Convert TEXT to JSONB if needed
	var nameType string
	err = db.QueryRow("SELECT data_type FROM information_schema.columns WHERE table_name = 'places' AND column_name = 'name'").Scan(&nameType)
	if err == nil && nameType != "jsonb" {
		log.Println("Migrating places columns to JSONB...")
		_, err = db.Exec(`
			ALTER TABLE places ALTER COLUMN name TYPE JSONB USING jsonb_build_object('tr', name);
			ALTER TABLE places ALTER COLUMN description TYPE JSONB USING jsonb_build_object('tr', description);
		`)
		if err != nil {
			log.Printf("Migration error: %v", err)
		}
	}

	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'pending'")
	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS image_url TEXT")
	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS city TEXT")
	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS category TEXT")
	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS creator_id INT REFERENCES users(id) ON DELETE SET NULL")
	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS price DOUBLE PRECISION DEFAULT 0")
	db.Exec("ALTER TABLE comments ADD COLUMN IF NOT EXISTS user_id INT REFERENCES users(id) ON DELETE CASCADE")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT DEFAULT ''")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT ''")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT DEFAULT ''")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS points INT DEFAULT 0")
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" { return }
	if r.Method != "POST" { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil { http.Error(w, "Invalid request", http.StatusBadRequest); return }
	
	// Password Strength Check
	if len(creds.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters long", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil { http.Error(w, "Server error", http.StatusInternalServerError); return }
	role := "user"
	if creds.SecretCode == AdminSecretCode { role = "admin" }
	var userID int
	err = db.QueryRow("INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id", creds.Username, string(hashedPassword), role).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") { http.Error(w, "Username already taken", http.StatusConflict); return }
		http.Error(w, "Database error", http.StatusInternalServerError); return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created", "role": role})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" { return }
	if r.Method != "POST" { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil { http.Error(w, "Invalid request", http.StatusBadRequest); return }
	var storedPassword, role string
	err := db.QueryRow("SELECT password, role FROM users WHERE username=$1", creds.Username).Scan(&storedPassword, &role)
	if err != nil { http.Error(w, "Invalid credentials", http.StatusUnauthorized); return }
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)); err != nil { http.Error(w, "Invalid credentials", http.StatusUnauthorized); return }
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{Username: creds.Username, Role: role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expirationTime)}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil { http.Error(w, "Server error", http.StatusInternalServerError); return }
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "role": role, "username": creds.Username})
}

func validateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil { return nil, err }
	if !token.Valid { return nil, fmt.Errorf("invalid token") }
	return claims, nil
}

func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == "OPTIONS" { return }
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" { http.Error(w, "Missing authorization header", http.StatusUnauthorized); return }
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := validateToken(tokenStr)
		if err != nil || claims.Role != "admin" { http.Error(w, "Forbidden: Admins only", http.StatusForbidden); return }
		next(w, r)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" { return }
	if r.Method != "POST" { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("image")
	if err != nil { http.Error(w, "Error retrieving file", http.StatusBadRequest); return }
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" { http.Error(w, "Invalid file type", http.StatusBadRequest); return }
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join("uploads", filename)
	dst, err := os.Create(filePath)
	if err != nil { http.Error(w, "Error saving file", http.StatusInternalServerError); return }
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil { http.Error(w, "Error saving file", http.StatusInternalServerError); return }
	fileURL := fmt.Sprintf("/uploads/%s", filename)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"url": fileURL})
}

func translateContent(text string) map[string]string {
	result := make(map[string]string)
	result["tr"] = text
	targetLangs := []string{"en", "de", "fr", "ru", "ar"}
	for _, lang := range targetLangs {
		// Temporary disable translation to prevent timeouts/errors from external service
		// translated, err := gtranslate.TranslateWithParams(text, gtranslate.TranslationParams{From: "tr", To: lang})
		// if err == nil { result[lang] = translated } else { result[lang] = text }
		result[lang] = text
	}
	return result
}

func placesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" { return }
	if r.Method == "GET" {
		authHeader := r.Header.Get("Authorization")
		var userID int
		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validateToken(tokenStr)
			if err == nil {
				db.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
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
		} else { query += " ORDER BY id DESC" }

		rows, err := db.Query(query, args...)
		if err != nil { http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError); return }
		defer rows.Close()
		var places []Place
		for rows.Next() {
			var p Place
			var nameJSON, descJSON []byte
			rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status, &p.IsFavorite)
			json.Unmarshal(nameJSON, &p.Name)
			json.Unmarshal(descJSON, &p.Description)
			places = append(places, p)
		}
		json.NewEncoder(w).Encode(places)
	} else if r.Method == "POST" {
		authHeader := r.Header.Get("Authorization")
		var creatorID int
		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validateToken(tokenStr)
			if err == nil { db.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&creatorID) }
		}
		var pr PlaceRequest
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil { http.Error(w, "Invalid body", http.StatusBadRequest); return }
		
		// Normalize City Name (Title Case with Turkish support)
		pr.City = cases.Title(language.Turkish).String(pr.City)

		nameMap := translateContent(pr.Name)
		descMap := translateContent(pr.Description)
		nameJSON, _ := json.Marshal(nameMap)
		descJSON, _ := json.Marshal(descMap)
		status := "pending"
		var id int
		var err error
		if creatorID > 0 {
			err = db.QueryRow("INSERT INTO places (name, description, lat, lng, category, city, image_url, status, creator_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id", string(nameJSON), string(descJSON), pr.Lat, pr.Lng, pr.Category, pr.City, pr.ImageURL, status, creatorID).Scan(&id)
			// Award Points (+50 XP)
			if err == nil {
				db.Exec("UPDATE users SET points = points + 50 WHERE id = $1", creatorID)
			}
		} else {
			err = db.QueryRow("INSERT INTO places (name, description, lat, lng, category, city, image_url, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", string(nameJSON), string(descJSON), pr.Lat, pr.Lng, pr.Category, pr.City, pr.ImageURL, status).Scan(&id)
		}
		if err != nil {
			log.Printf("Error inserting place: %v", err)
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		p := Place{ID: id, Name: nameMap, Description: descMap, Lat: pr.Lat, Lng: pr.Lng, Category: pr.Category, City: pr.City, ImageURL: pr.ImageURL, Status: status}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	} else if r.Method == "PUT" {
		var pr PlaceRequest
		json.NewDecoder(r.Body).Decode(&pr)
		nameMap := translateContent(pr.Name)
		descMap := translateContent(pr.Description)
		nameJSON, _ := json.Marshal(nameMap)
		descJSON, _ := json.Marshal(descMap)
		_ = nameJSON
		_ = descJSON
		json.NewEncoder(w).Encode(map[string]string{"status": "Update not fully implemented in multi-language mode yet"})
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	adminOnly(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		if r.Method == "GET" && action == "stats" {
			stats := make(map[string]interface{})
			var totalPlaces, pendingPlaces, totalUsers, totalComments int
			db.QueryRow("SELECT COUNT(*) FROM places").Scan(&totalPlaces)
			db.QueryRow("SELECT COUNT(*) FROM places WHERE status = 'pending'").Scan(&pendingPlaces)
			db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
			db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&totalComments)
			stats["total_places"] = totalPlaces
			stats["pending_places"] = pendingPlaces
			stats["total_users"] = totalUsers
			stats["total_comments"] = totalComments
			rows, _ := db.Query("SELECT category, COUNT(*) FROM places GROUP BY category")
			categories := make(map[string]int)
			for rows.Next() {
				var cat string
				var count int
				rows.Scan(&cat, &count)
				categories[cat] = count
			}
			rows.Close()
			stats["categories"] = categories
			json.NewEncoder(w).Encode(stats)
			return
		}
		if r.Method == "GET" && action == "users" {
			rows, _ := db.Query("SELECT id, username, role FROM users ORDER BY id ASC")
			defer rows.Close()
			var users []User
			for rows.Next() {
				var u User
				rows.Scan(&u.ID, &u.Username, &u.Role)
				users = append(users, u)
			}
			json.NewEncoder(w).Encode(users)
			return
		}
		if r.Method == "GET" && action == "pending" {
			rows, _ := db.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, '') as image_url, status FROM places WHERE status = 'pending' ORDER BY id DESC")
			defer rows.Close()
			var places []Place
			for rows.Next() {
				var p Place
				var nameJSON, descJSON []byte
				rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status)
				json.Unmarshal(nameJSON, &p.Name)
				json.Unmarshal(descJSON, &p.Description)
				places = append(places, p)
			}
			json.NewEncoder(w).Encode(places)
			return
		}
		if r.Method == "POST" && (action == "approve" || action == "reject") {
			var req struct { ID int `json:"id"` }
			json.NewDecoder(r.Body).Decode(&req)
			if action == "approve" { db.Exec("UPDATE places SET status = 'approved' WHERE id = $1", req.ID) } else { db.Exec("DELETE FROM places WHERE id = $1", req.ID) }
			w.WriteHeader(http.StatusOK)
		}
	})(w, r)
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" { return }
	if r.Method == "GET" {
		placeID := r.URL.Query().Get("place_id")
		rows, _ := db.Query("SELECT id, place_id, content, rating, created_at FROM comments WHERE place_id = $1 ORDER BY created_at DESC", placeID)
		defer rows.Close()
		comments := []Comment{}
		for rows.Next() {
			var c Comment
			rows.Scan(&c.ID, &c.PlaceID, &c.Content, &c.Rating, &c.CreatedAt)
			comments = append(comments, c)
		}
		json.NewEncoder(w).Encode(comments)
	} else if r.Method == "POST" {
		authHeader := r.Header.Get("Authorization")
		var userID int
		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validateToken(tokenStr)
			if err == nil { db.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID) }
		}
		var c Comment
		json.NewDecoder(r.Body).Decode(&c)
		if userID > 0 {
			db.QueryRow("INSERT INTO comments (place_id, content, rating, user_id) VALUES ($1, $2, $3, $4) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating, userID).Scan(&c.ID, &c.CreatedAt)
			// Award Points (+10 XP)
			db.Exec("UPDATE users SET points = points + 10 WHERE id = $1", userID)
		} else {
			db.QueryRow("INSERT INTO comments (place_id, content, rating) VALUES ($1, $2, $3) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating).Scan(&c.ID, &c.CreatedAt)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" { return }
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" { http.Error(w, "Missing authorization header", http.StatusUnauthorized); return }
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := validateToken(tokenStr)
	if err != nil { http.Error(w, "Invalid token", http.StatusUnauthorized); return }
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
	if err != nil { http.Error(w, "User not found", http.StatusNotFound); return }
	action := r.URL.Query().Get("action")
	if r.Method == "GET" {
		if action == "places" {
			rows, _ := db.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, ''), status FROM places WHERE creator_id = $1 ORDER BY id DESC", userID)
			defer rows.Close()
			var places []Place
			for rows.Next() {
				var p Place
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
			rows, _ := db.Query("SELECT c.id, c.content, c.rating, c.created_at, p.id, p.name FROM comments c JOIN places p ON c.place_id = p.id WHERE c.user_id = $1 ORDER BY c.created_at DESC", userID)
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
		var u User
		err := db.QueryRow("SELECT id, username, role, COALESCE(email, ''), COALESCE(bio, ''), COALESCE(avatar_url, ''), points FROM users WHERE id=$1", userID).Scan(&u.ID, &u.Username, &u.Role, &u.Email, &u.Bio, &u.AvatarURL, &u.Points)
		if err != nil { http.Error(w, "User not found", http.StatusNotFound); return }
		json.NewEncoder(w).Encode(u)
	} else if r.Method == "PUT" {
		var u User
		json.NewDecoder(r.Body).Decode(&u)
		db.Exec("UPDATE users SET email=$1, bio=$2, avatar_url=$3 WHERE id=$4", u.Email, u.Bio, u.AvatarURL, userID)
		u.Username = claims.Username
		u.Role = claims.Role
		json.NewEncoder(w).Encode(u)
	}
}

func favoritesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
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
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := validateToken(tokenStr)
	if err != nil { 
		log.Printf("Favorites: Invalid token: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return 
	}
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", claims.Username).Scan(&userID)
	if err != nil || userID == 0 { 
		log.Printf("Favorites: User not found for username %s", claims.Username)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return 
	}

	if r.Method == "GET" {
		// ... GET logic unchanged
		rows, err := db.Query(`
			SELECT p.id, p.name, p.description, p.lat, p.lng, p.category, p.city, COALESCE(p.image_url, ''), p.status 
			FROM places p 
			JOIN favorites f ON p.id = f.place_id 
			WHERE f.user_id = $1`, userID)
		if err != nil { http.Error(w, "Database error", http.StatusInternalServerError); return }
		defer rows.Close()
		var places []Place
		for rows.Next() {
			var p Place
			var nameJSON, descJSON []byte
			rows.Scan(&p.ID, &nameJSON, &descJSON, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status)
			json.Unmarshal(nameJSON, &p.Name)
			json.Unmarshal(descJSON, &p.Description)
			places = append(places, p)
		}
		json.NewEncoder(w).Encode(places)
	} else if r.Method == "POST" {
		var req struct { PlaceID int `json:"place_id"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { 
			log.Printf("Favorites POST: Invalid body: %v", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return 
		}
		_, err = db.Exec("INSERT INTO favorites (user_id, place_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, req.PlaceID)
		if err != nil { 
			log.Printf("Favorites POST: DB Error: %v", err)
			http.Error(w, "Database error: " + err.Error(), http.StatusInternalServerError)
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
		
		_, err = db.Exec("DELETE FROM favorites WHERE user_id = $1 AND place_id = $2", userID, placeID)
		if err != nil { 
			log.Printf("Favorites DELETE: DB Error: %v", err)
			http.Error(w, "Database error: " + err.Error(), http.StatusInternalServerError)
			return 
		}
		w.WriteHeader(http.StatusOK)
	}
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method != "GET" { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed); return }

	rows, err := db.Query("SELECT id, username, COALESCE(avatar_url, ''), points FROM users ORDER BY points DESC LIMIT 10")
	if err != nil { http.Error(w, "Database error", http.StatusInternalServerError); return }
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Username, &u.AvatarURL, &u.Points)
		users = append(users, u)
	}
	json.NewEncoder(w).Encode(users)
}

func main() {
	initDB()
	os.MkdirAll("uploads", os.ModePerm)
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))
	http.HandleFunc("/api/upload", uploadHandler)
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/places", placesHandler)
	http.HandleFunc("/api/comments", commentsHandler)
	http.HandleFunc("/api/admin", adminHandler)
	http.HandleFunc("/api/user", userHandler)
	http.HandleFunc("/api/favorites", favoritesHandler)
	http.HandleFunc("/api/leaderboard", leaderboardHandler)
	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)
}