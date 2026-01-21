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

func NewOTPServices (cfg *configs.Config) *OTPService {
	return &OTPService{cfg: cfg}
}


func(s *OTPService) GenerateAndStoreOTP(phoneNumber string)(string, error) {
	// generate 6digit OTP
	otp, err := utils.GeneratedOTP(s.cfg.OTPLength)
	if err != nil {
		return "", err
	}

	// store in redis with expiration
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", phoneNumber)
	expiration := time.Duration(s.cfg.OTPExpirationMinutes) * time.Minute


	err = database.RedisClient.Set(ctx, key, otp, expiration).Err()
	if err != nil {
		return "", fmt.Errorf("Failed to store OTP: %w", err)
	}
	return otp, nil
}


// this function checks if provided otp matches stored otp
func (s *OTPService) VerifyOTP(phoneNumber, providedOTP string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", phoneNumber)

	// Get OTP from Redis
	storedOTP, err := database.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("OTP not found or expired")
	}

	// compare OTPS

	if storedOTP != providedOTP {
		return false, nil
	}

	// Delete OTP after successful verification 
	database.RedisClient.Del(ctx, key)

	return true, nil
}