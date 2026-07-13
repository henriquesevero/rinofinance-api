// Package config centralizes reading environment variables so the rest of
// the codebase never calls os.Getenv directly. This is the one file that
// needs to change for Railway-specific env var names.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds every environment-driven setting the API needs to boot.
type Config struct {
	// Port is the TCP port the HTTP server listens on. Railway injects
	// this dynamically via the PORT env var.
	Port string

	// MongoURI is the MongoDB connection string (a mongodb+srv:// Atlas URI
	// in production).
	MongoURI string

	// MongoDatabase is the name of the database to use within the MongoDB
	// deployment (Atlas connection strings don't always carry one).
	MongoDatabase string

	// JWTSecret signs and validates authentication tokens.
	JWTSecret string

	// JWTTTL is how long an issued token stays valid.
	JWTTTL time.Duration

	// AllowedOrigins is the list of origins allowed by CORS (the Vercel
	// frontend URL(s) in production).
	AllowedOrigins []string
}

// Load reads and validates configuration from environment variables,
// failing fast on startup if anything required is missing rather than
// surfacing a confusing error later.
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("variável de ambiente MONGO_URI é obrigatória")
	}

	mongoDatabase := os.Getenv("MONGO_DATABASE")
	if mongoDatabase == "" {
		mongoDatabase = "rinofinance"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("variável de ambiente JWT_SECRET é obrigatória")
	}

	ttl := 24 * time.Hour
	if raw := os.Getenv("JWT_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("JWT_TTL inválido (%q): %w", raw, err)
		}
		ttl = parsed
	}

	origins := []string{"http://localhost:5173"}
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		origins = splitAndTrim(raw)
	}

	return &Config{
		Port:           port,
		MongoURI:       mongoURI,
		MongoDatabase:  mongoDatabase,
		JWTSecret:      jwtSecret,
		JWTTTL:         ttl,
		AllowedOrigins: origins,
	}, nil
}

func splitAndTrim(raw string) []string {
	var result []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
