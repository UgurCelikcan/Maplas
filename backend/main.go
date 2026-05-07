package main

import (
	"fmt"
	"net/http"
	"os"

	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/utils"
)

func main() {
	db.InitDB()
	os.MkdirAll("uploads", os.ModePerm)
	
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))
	
	http.HandleFunc("/upload", handlers.UploadHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/places", handlers.PlacesHandler)
	http.HandleFunc("/comments", handlers.CommentsHandler)
	http.HandleFunc("/admin", handlers.AdminHandler)
	http.HandleFunc("/user", handlers.UserHandler)
	http.HandleFunc("/favorites", handlers.FavoritesHandler)
	http.HandleFunc("/leaderboard", handlers.LeaderboardHandler)
	
	port := utils.GetEnv("SERVER_PORT", "8080")
	fmt.Printf("Server starting on port %s...\n", port)
	http.ListenAndServe(":"+port, nil)
}
