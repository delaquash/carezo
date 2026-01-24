package services

import (
	"context"
	"fmt"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	"github.com/delaquash/carezo/internal/utils"
)

type OTPService struct {
	cfg *configs.Config
}

func NewOTPService(cfg *configs.Config) *OTPService {
	return &OTPService{cfg: cfg}
}

// create OTP and stores it in Redis with expiration

func (s *OTPService) GenerateAndStoreOTP(email string) (string, error) {
	// Generate 6-digit OTP
	otp, err := utils.GeneratedOTP(s.cfg.OTPLength)
	if err != nil {
		return "", err
	}

	// Store in Redis with expiration
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", email) // Changed from phone to email
	expiration := time.Duration(s.cfg.OTPExpirationMinutes) * time.Minute

	err = database.RedisClient.Set(ctx, key, otp, expiration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}

	return otp, nil
}

//  checks if provided OTP matches stored OTP
func (s *OTPService) VerifyOTP(email, providedOTP string) (bool, error) {
	ctx := context.Background()
	// Changed from phone to email
	key := fmt.Sprintf("otp:%s", email) 

	// Get OTP from Redis
	storedOTP, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("OTP not found or expired")
	}

	// Compare OTPs
	if storedOTP != providedOTP {
		return false, nil
	}

	// Delete OTP after successful verification (can only be used once)
	database.RedisClient.Del(ctx, key)

	return true, nil
}