package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	APIKey         string
	BaseURL        string
	PublicURL      string
	WebhookURL     string
	WebhookSecret  string
	DefaultInterestRate float64
	DBPath         string
}

func Load() *Config {
	godotenv.Load()

	port := getEnv("MOCKPAY_PORT", "8080")
	baseURL := getEnv("MOCKPAY_BASE_URL", "http://localhost:"+port)
	publicURL := getEnv("MOCKPAY_PUBLIC_URL", baseURL)
	return &Config{
		Port:           port,
		APIKey:         getEnv("MOCKPAY_API_KEY", "mock_key"),
		BaseURL:        baseURL,
		PublicURL:      publicURL,
		WebhookURL:     getEnv("MOCKPAY_WEBHOOK_URL", ""),
		WebhookSecret:  getEnv("MOCKPAY_WEBHOOK_SECRET", ""),
		DefaultInterestRate: getEnvFloat("MOCKPAY_INTEREST_RATE", 0),
		DBPath:         getEnv("MOCKPAY_DB_PATH", "mockpay.db"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return strings.ToLower(v) == "true" || v == "1"
}
