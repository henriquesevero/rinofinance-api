package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port string

	MongoURI string

	MongoDatabase string

	JWTSecret string

	JWTTTL time.Duration

	AllowedOrigins []string

	RegistrationCode string

	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDEmail      string
}

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

	ttl := 365 * 24 * time.Hour
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

	registrationCode := os.Getenv("REGISTRATION_CODE")
	if registrationCode == "" {
		registrationCode = "171598"
	}

	vapidEmail := os.Getenv("VAPID_EMAIL")
	if vapidEmail == "" {
		vapidEmail = "contato@henriquesevero.com"
	}

	return &Config{
		Port:             port,
		MongoURI:         mongoURI,
		MongoDatabase:    mongoDatabase,
		JWTSecret:        jwtSecret,
		JWTTTL:           ttl,
		AllowedOrigins:   origins,
		RegistrationCode: registrationCode,
		VAPIDPublicKey:   os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:  os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDEmail:       vapidEmail,
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
