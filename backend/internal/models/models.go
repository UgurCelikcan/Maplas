package models

import "time"

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
	Name        string  `json:"name"`        // Frontend sends string
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
