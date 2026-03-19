package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	DBUser      string
	DBPassword  string
	DBName      string
	DBHost      string
	DBPort      string
	RedisHost   string
	RedisPort   string
	JWTSecret   string
	AWSRegion   string
	AWSEndpoint string
}

func LoadConfig() *Config {
	// Try to load .env, but don't fail if it doesn't exist (e.g. in Docker)
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:     getEnv("API_PORT", "8080"),
		DBUser:      getEnv("DB_USER", "fc_user"),
		DBPassword:  getEnv("DB_PASSWORD", "fc_password"),
		DBName:      getEnv("DB_NAME", "fitchallenge"),
		DBHost:      getEnv("DB_HOST", "postgres"), // Default for docker
		DBPort:      getEnv("DB_PORT", "5432"),
		RedisHost:   getEnv("REDIS_HOST", "redis"), // Default for docker
		RedisPort:   getEnv("REDIS_PORT", "6379"),
		JWTSecret:   getEnv("JWT_SECRET", "super_secret"),
		AWSRegion:   getEnv("AWS_DEFAULT_REGION", "us-east-1"),
		AWSEndpoint: getEnv("AWS_ENDPOINT_URL", ""),
	}

	log.Println("✅ Configuration loaded successfully")
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
