package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"backend/internal/auth"
	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/utils"
)

func main() {
	if err := auth.Init(); err != nil {
		log.Fatalf("Startup aborted: %v", err)
	}
	db.InitDB()
	os.MkdirAll("uploads", os.ModePerm)
	
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", noDirListing(fs)))
	
	// Rate limits per IP (requests per minute, burst)
	general := middleware.NewRateLimiter(120, 60) // every API call
	authLimit := middleware.NewRateLimiter(10, 5) // login / register (password guessing)
	uploadLimit := middleware.NewRateLimiter(10, 5)
	writeLimit := middleware.NewRateLimiter(20, 10) // adding places, comments, favorites

	http.HandleFunc("/upload", general.Limit(uploadLimit.Limit(handlers.UploadHandler)))
	http.HandleFunc("/register", general.Limit(authLimit.Limit(handlers.RegisterHandler)))
	http.HandleFunc("/login", general.Limit(authLimit.Limit(handlers.LoginHandler)))
	http.HandleFunc("/places", general.Limit(writeLimit.LimitWrites(handlers.PlacesHandler)))
	http.HandleFunc("/comments", general.Limit(writeLimit.LimitWrites(handlers.CommentsHandler)))
	http.HandleFunc("/admin", general.Limit(handlers.AdminHandler))
	// Writes on /user include account deletion (password check), so they get the strict limit.
	http.HandleFunc("/user", general.Limit(authLimit.LimitWrites(handlers.UserHandler)))
	http.HandleFunc("/favorites", general.Limit(writeLimit.LimitWrites(handlers.FavoritesHandler)))
	http.HandleFunc("/leaderboard", general.Limit(handlers.LeaderboardHandler))

	port := utils.GetEnv("SERVER_PORT", "8080")
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// noDirListing hides the list of uploaded files; single files are still served.
func noDirListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// File names are random and never reused, so browsers can cache them for a year.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
