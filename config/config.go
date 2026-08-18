package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort         string
	DatabaseURL        string
	JWKSUrl            string
	AllowedOrigins     string
	KafkaBrokers       string
	KafkaConsumerGroup string
	KafkaRequestTopic  string
	KafkaInspectTopic  string
	KafkaReviewTopic   string
	RedisAddr          string
	RedisPassword      string
	DedupTTL           time.Duration
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPassword       string
	SMTPFrom           string
	SMTPEnabled        bool
	ClientPortalURL    string
	OfficePortalURL    string
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8083"),
		DatabaseURL:        dbURL,
		JWKSUrl:            getEnv("JWKS_URL", "http://localhost:8180/realms/appraisal/protocol/openid-connect/certs"),
		AllowedOrigins:     getEnv("ALLOWED_ORIGINS", "*"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9094"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "notification-service"),
		KafkaRequestTopic:  getEnv("KAFKA_REQUEST_TOPIC", "request.events"),
		KafkaInspectTopic:  getEnv("KAFKA_INSPECT_TOPIC", "inspect.events"),
		KafkaReviewTopic:   getEnv("KAFKA_REVIEW_TOPIC", "review.events"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6381"),
		RedisPassword:      getEnv("REDIS_PASSWORD", "appraisal"),
		DedupTTL:           getDurationEnv("DEDUP_TTL", 24*time.Hour),
		SMTPHost:           getEnv("SMTP_HOST", "localhost"),
		SMTPPort:           getIntEnv("SMTP_PORT", 1025),
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:           getEnv("SMTP_FROM", "noreply@appraisal-crm.ru"),
		SMTPEnabled:        getBoolEnv("SMTP_ENABLED", true),
		ClientPortalURL:    getEnv("CLIENT_PORTAL_URL", "http://localhost:5173"),
		OfficePortalURL:    getEnv("OFFICE_PORTAL_URL", "http://localhost:5174"),
	}
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("invalid %s: %v", key, err)
		}
		return d
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("invalid %s: %v", key, err)
		}
		return i
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Fatalf("invalid %s: %v", key, err)
		}
		return b
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
