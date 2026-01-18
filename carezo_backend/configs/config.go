package configs


import (
	"logs"
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

	 JWT (JSON Web Token) settings for authentication
	JWTSecret              string
	JWTExpirationHours     int
	RefreshTokenExpiration time.Duration

	// OTP (One-Time Password) settings
	OTPExpirationMinutes int
	OTPLength            int

	// Email settings (for sending OTP via email)
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string

	// Twilio settings (for sending OTP via SMS)
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioPhoneNumber string

	// Google OAuth settings
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Paystack settings (for payment processing)
	PaystackSecretKey string
	PaystackPublicKey string

	// Rate limiting settings
	RateLimitRequests       int
	RateLimitWindowSeconds  int

	// File upload settings
	MaxUploadSizeMB    int
	UploadPath         string
	AllowedImageTypes  string
}