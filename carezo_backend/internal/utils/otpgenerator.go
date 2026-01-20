package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

)

// func that generateds a 6-digit OTP
func GeneratedOTP(length int) (string, error) {
	// alculate the maximum value
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	// Generate random number
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("Failed to generate OTP: %")
	}

	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}