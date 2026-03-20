package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppPort     string
	DBUser      string
	DBPassword  string
	DBName      string
	DBHost      string
	DBPort      string
	DBMaxConns  int32
	DBMinConns  int32
	RedisHost   string
	RedisPort   string
	JWTSecret   string
	AWSRegion   string
	AWSEndpoint string
	S3PublicURL string // Az URL, amin a kliens eléri az S3 fájlokat (pl. CDN vagy LocalStack)
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("API_PORT", "8080"),
		DBUser:      getEnv("DB_USER", "fc_user"),
		DBPassword:  getEnv("DB_PASSWORD", "fc_password"),
		DBName:      getEnv("DB_NAME", "fitchallenge"),
		DBHost:      getEnv("DB_HOST", "postgres"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBMaxConns:  getEnvInt("DB_MAX_CONNS", 10),
		DBMinConns:  getEnvInt("DB_MIN_CONNS", 2),
		RedisHost:   getEnv("REDIS_HOST", "redis"),
		RedisPort:   getEnv("REDIS_PORT", "6379"),
		JWTSecret:   getEnv("JWT_SECRET", "super_secret"),
		AWSRegion:   getEnv("AWS_DEFAULT_REGION", "us-east-1"),
		AWSEndpoint: getEnv("AWS_ENDPOINT_URL", ""),
		S3PublicURL: getEnv("S3_PUBLIC_URL", "http://localhost:4566"),
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

func getEnvInt(key string, fallback int32) int32 {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int32(i)
		}
	}
	return fallback
}
