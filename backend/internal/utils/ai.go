package utils

import (
	"bytes"
	"time"
)

type AIValidationResult struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

// ValidatePlaceWithAI simulates or calls an AI service to validate a new place.
// In a production environment, you would call Gemini API or a similar LLM.
func ValidatePlaceWithAI(name, description, category, city string) (bool, string) {
	// For this implementation, we simulate an AI decision.
	// You can replace this with a real call to Gemini API.
	
	// Example prompt that would be sent to an LLM:
	/*
	"Analyze the following place submission for a travel app:
	Name: {{name}}
	Description: {{description}}
	Category: {{category}}
	City: {{city}}
	
	Is this a real or plausible place? Is the description appropriate and safe?
	Respond in JSON format: {\"approved\": boolean, \"reason\": \"string\"}"
	*/

	// Simulation Logic:
	// 1. Basic length check
	if len(name) < 3 || len(description) < 10 {
		return false, "Name or description is too short."
	}

	// 2. Mock "Safe Search"
	badWords := []string{"küfür1", "kötükelime2"} // Örnek filtreler
	for _, word := range badWords {
		if containsIgnoreCase(name, word) || containsIgnoreCase(description, word) {
			return false, "Content contains inappropriate language."
		}
	}

	// 3. Mock AI check (In reality, this would be an API call)
	// We simulate a 2-second AI processing time
	time.Sleep(1 * time.Second)

	return true, "AI Verified: Place looks authentic and safe."
}

func containsIgnoreCase(s, substr string) bool {
	return bytes.Contains(bytes.ToLower([]byte(s)), bytes.ToLower([]byte(substr)))
}

// CallGeminiAPI is a placeholder for actual integration
func CallGeminiAPI(apiKey, prompt string) (*AIValidationResult, error) {
	// This is where you'd implement the actual HTTP call to Google's Generative AI API
	// API URL: https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent
	return &AIValidationResult{Approved: true, Reason: "Mocked AI success"}, nil
}
