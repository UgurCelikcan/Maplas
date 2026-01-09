package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Place struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Category    string  `json:"category"`
	City        string  `json:"city"`
}

var db *sql.DB

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

	// Retry connection loop for Docker
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to DB, retrying in 2 seconds... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS places (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		category TEXT,
		city TEXT
	);
	`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// seedDB()
}

func seedDB() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM places").Scan(&count)
	if err != nil {
		log.Printf("Error checking row count: %v", err)
		return
	}

	if count > 0 {
		log.Println("Database already seeded.")
		return
	}

	log.Println("Seeding database from places.json...")
	file, err := os.ReadFile("places.json")
	if err != nil {
		log.Printf("Error reading places.json: %v", err)
		return
	}

	var places []Place
	if err := json.Unmarshal(file, &places); err != nil {
		log.Printf("Error parsing places.json: %v", err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO places (name, description, lat, lng, category, city) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	for _, p := range places {
		_, err := stmt.Exec(p.Name, p.Description, p.Lat, p.Lng, p.Category, p.City)
		if err != nil {
			log.Printf("Error inserting place %s: %v", p.Name, err)
		}
	}
	log.Println("Database seeding completed.")
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func placesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method == "GET" {
		rows, err := db.Query("SELECT id, name, description, lat, lng, category, city FROM places ORDER BY id DESC")
		if err != nil {
			http.Error(w, "Database query error", http.StatusInternalServerError)
			log.Printf("Error querying database: %v", err)
			return
		}
		defer rows.Close()

		var places []Place
		for rows.Next() {
			var p Place
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Lat, &p.Lng, &p.Category, &p.City); err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			places = append(places, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(places)
	} else if r.Method == "POST" {
		var p Place
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := db.QueryRow(
			"INSERT INTO places (name, description, lat, lng, category, city) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
			p.Name, p.Description, p.Lat, p.Lng, p.Category, p.City,
		).Scan(&p.ID)

		if err != nil {
			http.Error(w, "Database insert error", http.StatusInternalServerError)
			log.Printf("Error inserting place: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	} else if r.Method == "PUT" {
		var p Place
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if p.ID == 0 {
			http.Error(w, "Missing place ID", http.StatusBadRequest)
			return
		}

		res, err := db.Exec(
			"UPDATE places SET name=$1, description=$2, lat=$3, lng=$4, category=$5, city=$6 WHERE id=$7",
			p.Name, p.Description, p.Lat, p.Lng, p.Category, p.City, p.ID,
		)

		if err != nil {
			http.Error(w, "Database update error", http.StatusInternalServerError)
			log.Printf("Error updating place: %v", err)
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Place not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	} else if r.Method == "DELETE" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		res, err := db.Exec("DELETE FROM places WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Database delete error", http.StatusInternalServerError)
			log.Printf("Error deleting place: %v", err)
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Place not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	initDB()
	http.HandleFunc("/api/places", placesHandler)

	port := "8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
