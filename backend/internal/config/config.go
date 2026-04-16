package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Env               string
		Port              string
		ReadHeaderTimeout time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
	}
	DB struct {
		User     string
		Password string
		Name     string
		Host     string
		Port     string
		MaxConns int32
		MinConns int32
	}
	Redis struct {
		Host string
		Port string
	}
	JWT struct {
		Secret string
	}
	AWS struct {
		Region   string
		Endpoint string
	}
	S3PublicURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{}

	// App configuration
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Port = getEnv("API_PORT", "8080")
	cfg.App.ReadHeaderTimeout = getEnvDuration("API_READ_HEADER_TIMEOUT", 5*time.Second)
	cfg.App.ReadTimeout = getEnvDuration("API_READ_TIMEOUT", 10*time.Second)
	cfg.App.WriteTimeout = getEnvDuration("API_WRITE_TIMEOUT", 30*time.Second)
	cfg.App.IdleTimeout = getEnvDuration("API_IDLE_TIMEOUT", 60*time.Second)

	// DB configuration
	cfg.DB.User = getEnv("DB_USER", "fc_user")
	cfg.DB.Password = getEnv("DB_PASSWORD", "fc_password")
	cfg.DB.Name = getEnv("DB_NAME", "fitchallenge")
	cfg.DB.Host = getEnv("DB_HOST", "postgres")
	cfg.DB.Port = getEnv("DB_PORT", "5432")
	cfg.DB.MaxConns = getEnvInt("DB_MAX_CONNS", 10)
	cfg.DB.MinConns = getEnvInt("DB_MIN_CONNS", 2)

	// Redis configuration
	cfg.Redis.Host = getEnv("REDIS_HOST", "redis")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")

	// JWT configuration
	cfg.JWT.Secret = getEnv("JWT_SECRET", "super_secret")

	// AWS configuration
	cfg.AWS.Region = getEnv("AWS_DEFAULT_REGION", "us-east-1")
	cfg.AWS.Endpoint = getEnv("AWS_ENDPOINT_URL", "")

	// Public configuration
	cfg.S3PublicURL = getEnv("S3_PUBLIC_URL", "http://localhost:4566")

	log.Println("✅ Configuration loaded successfully")
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int32) int32 {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int32(i)
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := time.ParseDuration(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func IsProductionEnv(env string) bool {
	return strings.EqualFold(strings.TrimSpace(env), "production")
}
