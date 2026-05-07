package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"backend/internal/models"
	"backend/internal/utils"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	dbHost := utils.GetEnv("DB_HOST", "localhost")
	dbUser := utils.GetEnv("DB_USER", "user")
	dbPassword := utils.GetEnv("DB_PASSWORD", "password")
	dbName := utils.GetEnv("DB_NAME", "places_db")
	dbPort := utils.GetEnv("DB_PORT", "5432")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	for i := 0; i < 10; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err == nil {
			err = DB.Ping()
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

	createTables()
	migrateToJSONB()
	seedData()
}

func createTables() {
	createTablesQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		email TEXT DEFAULT '',
		bio TEXT DEFAULT '',
		avatar_url TEXT DEFAULT '',
		points INT DEFAULT 0
	);

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
		creator_id INT REFERENCES users(id) ON DELETE SET NULL,
		price DOUBLE PRECISION DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		place_id INT REFERENCES places(id) ON DELETE CASCADE,
		content TEXT,
		rating INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		user_id INT REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS favorites (
		user_id INT REFERENCES users(id) ON DELETE CASCADE,
		place_id INT REFERENCES places(id) ON DELETE CASCADE,
		PRIMARY KEY (user_id, place_id)
	);
	`

	_, err := DB.Exec(createTablesQuery)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func migrateToJSONB() {
	var nameType string
	err := DB.QueryRow("SELECT data_type FROM information_schema.columns WHERE table_name = 'places' AND column_name = 'name'").Scan(&nameType)
	if err == nil && nameType != "jsonb" {
		log.Println("Migrating places columns to JSONB...")
		_, err = DB.Exec(`
			ALTER TABLE places ALTER COLUMN name TYPE JSONB USING jsonb_build_object('tr', name);
			ALTER TABLE places ALTER COLUMN description TYPE JSONB USING jsonb_build_object('tr', description);
		`)
		if err != nil {
			log.Printf("Migration error: %v", err)
		}
	}
}

func seedData() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM places").Scan(&count)
	if count == 0 {
		log.Println("Seeding initial places from places.json...")
		file, err := os.ReadFile("places.json")
		if err == nil {
			var seedPlaces []models.PlaceRequest
			if err := json.Unmarshal(file, &seedPlaces); err == nil {
				for _, sp := range seedPlaces {
					nameMap := TranslateContent(sp.Name)
					descMap := TranslateContent(sp.Description)
					nameJSON, _ := json.Marshal(nameMap)
					descJSON, _ := json.Marshal(descMap)
					_, err = DB.Exec("INSERT INTO places (name, description, lat, lng, category, city, status) VALUES ($1, $2, $3, $4, $5, $6, $7)",
						string(nameJSON), string(descJSON), sp.Lat, sp.Lng, sp.Category, sp.City, "approved")
					if err != nil {
						log.Printf("Error seeding place %s: %v", sp.Name, err)
					}
				}
				log.Printf("Successfully seeded %d places", len(seedPlaces))
			}
		} else {
			log.Printf("Could not find places.json for seeding: %v", err)
		}
	}
}

func TranslateContent(text string) map[string]string {
	result := make(map[string]string)
	result["tr"] = text
	targetLangs := []string{"en", "de", "fr", "ru", "ar"}
	for _, lang := range targetLangs {
		result[lang] = text
	}
	return result
}
