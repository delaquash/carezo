package services

import (
	"context"
	"fmt"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/utils"
)

type OTPService struct {
	cfg *configs.Config
}

func NewOTPServices (cfg *configs.Configs) *OTPService {
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
	expiration := time.Duration(s.cfg.OTPExpirationMinutes) * time.Miu\


	err = database.RedisClient.Set(ctx)
}
