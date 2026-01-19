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
		if value := os.Getenv(key); value != "" {
			if intVal, err := strconv.Atoi(value); err == nil {
				return intVal
			}
		}

		return defaultValue	
	}

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "carezo_user"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "carezo_db"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Application
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "carezo"),

		// JWT
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpirationHours:     getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		RefreshTokenExpiration: time.Duration(getEnvAsInt("REFRESH_TOKEN_EXPIRATION_DAYS", 30)) * 24 * time.Hour,

		// OTP
		OTPExpirationMinutes: getEnvAsInt("OTP_EXPIRATION_MINUTES", 10),
		OTPLength:            getEnvAsInt("OTP_LENGTH", 6),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@carezo.com"),
		FromName:     getEnv("FROM_NAME", "carezo"),

		// Twilio
		TwilioAccountSID:  getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:   getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber: getEnv("TWILIO_PHONE_NUMBER", ""),

		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),

		// Paystack
		PaystackSecretKey: getEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackPublicKey: getEnv("PAYSTACK_PUBLIC_KEY", ""),

		// Rate Limiting
		RateLimitRequests:      getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSeconds: getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60),

		// File Upload
		MaxUploadSizeMB:   getEnvAsInt("MAX_UPLOAD_SIZE_MB", 5),
		UploadPath:        getEnv("UPLOAD_PATH", "./uploads"),
		AllowedImageTypes: getEnv("ALLOWED_IMAGE_TYPES", "jpg,jpeg,png,webp"),
	}
}