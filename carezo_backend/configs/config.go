package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost 		string
	DBPort 		string
	DBUser 		string
	DBPassword 	string
	DBName		string

	// Redis Setting
	RedisHost 	  string
	RedisPort 	  string
	RedisPassword string

	AppEnv 	string
	AppPort	string
	AppName	string

	JWTSecret              string
	JWTExpirationHours     int
	RefreshTokenExpiration time.Duration

	OTPExpirationMinutes int
	OTPLength            int

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string

	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioPhoneNumber string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	PaystackSecretKey string
	PaystackPublicKey string


	RateLimitRequests       int
	RateLimitWindowSeconds  int


	MaxUploadSizeMB    int
	UploadPath         string
	AllowedImageTypes  string
}


func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variable")
	}
	// helper to get env variable with original value and return default if not exist
	getEnv := func (key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	// convert string env to int
	getEnvAsInt := func(key string, defaultValue int) int {
		if value := os.Getenv(key); value !+ "" {
			if intVal, err := strconv.Atoi(value); err == nil {
				return intVal
			}
		}

		return defaultValue	
}