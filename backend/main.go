package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// --- Structs ---

type Place struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Category    string  `json:"category"`
	City        string  `json:"city"`
	ImageURL    string  `json:"imageUrl"`
	Status      string  `json:"status"` // 'pending' or 'approved'
}

type Comment struct {
	ID        int       `json:"id"`
	PlaceID   int       `json:"place_id"`
	Content   string    `json:"content"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // omitempty to hide in responses
	Role     string `json:"role"`
}

type Credentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	SecretCode string `json:"secret_code,omitempty"` // For becoming admin
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// --- Globals ---

var db *sql.DB
var jwtKey = []byte("my_super_secret_key_2026") // In production, use env var
const AdminSecretCode = "HIDDEN_KING_2026"      // Code to become admin

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

	createTables := `
	CREATE TABLE IF NOT EXISTS places (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		category TEXT,
		city TEXT,
		image_url TEXT,
		status TEXT DEFAULT 'pending'
	);

	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		place_id INT REFERENCES places(id) ON DELETE CASCADE,
		content TEXT,
		rating INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'user'
	);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	
	// Ensure status column exists (migration)
	db.Exec("ALTER TABLE places ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'pending'")
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// --- Auth Handlers ---

func registerHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	role := "user"
	if creds.SecretCode == AdminSecretCode {
		role = "admin"
	}

	var userID int
	err = db.QueryRow("INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id",
		creds.Username, string(hashedPassword), role).Scan(&userID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created", "role": role})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var storedPassword, role string
	err := db.QueryRow("SELECT password, role FROM users WHERE username=$1", creds.Username).Scan(&storedPassword, &role)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"role":  role,
		"username": creds.Username,
	})
}

// --- Middleware ---

func validateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == "OPTIONS" {
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := validateToken(tokenStr)
		if err != nil || claims.Role != "admin" {
			http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// --- Place Handlers ---

func placesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method == "GET" {
		rows, err := db.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, '') as image_url, status FROM places WHERE status = 'approved' ORDER BY id DESC")
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var places []Place
		for rows.Next() {
			var p Place
			rows.Scan(&p.ID, &p.Name, &p.Description, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status)
			places = append(places, p)
		}
		json.NewEncoder(w).Encode(places)

	} else if r.Method == "POST" {
		// Public can post, but it goes to pending
		var p Place
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}

		status := "pending"
		err := db.QueryRow(
			"INSERT INTO places (name, description, lat, lng, category, city, image_url, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
			p.Name, p.Description, p.Lat, p.Lng, p.Category, p.City, p.ImageURL, status,
		).Scan(&p.ID)

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		p.Status = status
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)

	} else if r.Method == "DELETE" {
		// Only admins can delete
		adminOnly(func(w http.ResponseWriter, r *http.Request) {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "Missing id", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("DELETE FROM places WHERE id = $1", id)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})(w, r)
	} else if r.Method == "PUT" {
		// Basic update (simplified permissions for now, could act like POST)
		var p Place
		json.NewDecoder(r.Body).Decode(&p)
		db.Exec("UPDATE places SET name=$1, description=$2, category=$3, city=$4, image_url=$5 WHERE id=$6",
			p.Name, p.Description, p.Category, p.City, p.ImageURL, p.ID)
		json.NewEncoder(w).Encode(p)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	adminOnly(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")
		
		if r.Method == "GET" && action == "pending" {
			rows, _ := db.Query("SELECT id, name, description, lat, lng, category, city, COALESCE(image_url, '') as image_url, status FROM places WHERE status = 'pending' ORDER BY id DESC")
			defer rows.Close()
			var places []Place
			for rows.Next() {
				var p Place
				rows.Scan(&p.ID, &p.Name, &p.Description, &p.Lat, &p.Lng, &p.Category, &p.City, &p.ImageURL, &p.Status)
				places = append(places, p)
			}
			json.NewEncoder(w).Encode(places)
			return
		}

		if r.Method == "POST" && (action == "approve" || action == "reject") {
			var req struct { ID int `json:"id"` }
			json.NewDecoder(r.Body).Decode(&req)
			
			if action == "approve" {
				db.Exec("UPDATE places SET status = 'approved' WHERE id = $1", req.ID)
			} else {
				db.Exec("DELETE FROM places WHERE id = $1", req.ID)
			}
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
		var c Comment
		json.NewDecoder(r.Body).Decode(&c)
		db.QueryRow("INSERT INTO comments (place_id, content, rating) VALUES ($1, $2, $3) RETURNING id, created_at", c.PlaceID, c.Content, c.Rating).Scan(&c.ID, &c.CreatedAt)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func main() {
	initDB()
	
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	
	http.HandleFunc("/api/places", placesHandler)
	http.HandleFunc("/api/comments", commentsHandler)
	http.HandleFunc("/api/admin", adminHandler)

	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)
}