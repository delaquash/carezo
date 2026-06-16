package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// admin seeder
	AdminEmail     string
	AdminPassword  string
	AdminFirstName string
	AdminLastName  string

	// Redis settings
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// Application settings
	AppEnv  string
	AppPort string
	AppName string

	// JWT settings
	JWTSecret              string
	JWTExpirationHours     int
	RefreshTokenExpiration time.Duration

	// OTP settings (email-only now)
	OTPExpirationMinutes int
	OTPLength            int

	// Email settings (Gmail SMTP)
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string

	// Google OAuth settings
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Paystack settings
	PaystackSecretKey     string
	PaystackPublicKey     string
	PaystackWebhookSecret string

	// Rate limiting settings
	RateLimitRequests      int
	RateLimitWindowSeconds int

	// Cloudinary settings
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string

	// File upload settings
	MaxUploadSizeMB   int
	UploadPath        string
	AllowedImageTypes string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	getEnv := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

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
		DBUser:     getEnv("DB_USER", "driveease_user"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "driveease_db"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Application
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "DriveEase"),

		// JWT
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpirationHours:     getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		RefreshTokenExpiration: time.Duration(getEnvAsInt("REFRESH_TOKEN_EXPIRATION_DAYS", 30)) * 24 * time.Hour,

		// OTP (email-only)
		OTPExpirationMinutes: getEnvAsInt("OTP_EXPIRATION_MINUTES", 10),
		OTPLength:            getEnvAsInt("OTP_LENGTH", 6),

		// Email (Gmail SMTP)
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@driveease.com"),
		FromName:     getEnv("FROM_NAME", "DriveEase"),

		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),

		// Paystack
		PaystackSecretKey:     getEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackPublicKey:     getEnv("PAYSTACK_PUBLIC_KEY", ""),
		PaystackWebhookSecret: getEnv("PAYSTACK_WEBHOOK_SECRET", ""),

		// Rate Limiting
		RateLimitRequests:      getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSeconds: getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60),

		// File Upload
		MaxUploadSizeMB:   getEnvAsInt("MAX_UPLOAD_SIZE_MB", 5),
		UploadPath:        getEnv("UPLOAD_PATH", "./uploads"),
		AllowedImageTypes: getEnv("ALLOWED_IMAGE_TYPES", "jpg,jpeg,png,webp"),

		// Cloudinary
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),

		// Admin seeder
		// Admin seeder
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@carezo.com"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "Admin123!"),
		AdminFirstName: getEnv("ADMIN_FIRST_NAME", "Carezo"),
		AdminLastName:  getEnv("ADMIN_LAST_NAME", "Admin"),
	}
}
