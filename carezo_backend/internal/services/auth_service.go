package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	"github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	cfg          *configs.Config
	otpService   *OTPService
	emailService *EmailService
}

func NewAuthService(cfg *configs.Config) *AuthService {
	return &AuthService{
		cfg:          cfg,
		otpService:   NewOTPService(cfg),
		emailService: NewEmailService(cfg),
	}
}

// Register creates a new user account and sends OTP
// Step 1 of registration process
func (s *AuthService) Register(req *model.RegisterRequest) error {
	// 1. Check if user already exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone_number = $1 OR email = $2)`
	err := database.DB.Get(&exists, query, req.PhoneNumber, req.Email)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if exists {
		return errors.New("user with this phone number or email already exists")
	}

	// 2. Hash password (NEVER store plain text passwords!)
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// 3. Create user in database (incomplete profile, pending verification)
	userID := uuid.New().String()
	query = `
		INSERT INTO users (id, phone_number, email, password_hash, oauth_provider, status, role)
		VALUES ($1, $2, $3, $4, 'local', 'active', 'user')
	`
	_, err = database.DB.Exec(query, userID, req.PhoneNumber, req.Email, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 4. Generate and send OTP to phone
	otp, err := s.otpService.GenerateAndStoreOTP(req.PhoneNumber)
	if err != nil {
		return err
	}

	// 5. Send OTP via SMS (Twilio) - In development, just log it
	fmt.Printf("📱 OTP for %s: %s\n", req.PhoneNumber, otp)
	// In production: use Twilio API to send SMS

	// 6. Also send OTP via email as backup
	err = s.emailService.SendOTPEmail(req.Email, otp)
	if err != nil {
		// Log error but don't fail the registration
		fmt.Printf("Warning: Failed to send OTP email: %v\n", err)
	}

	return nil
}

// VerifyOTP verifies the OTP code sent to user's phone
// Step 2 of registration process
func (s *AuthService) VerifyOTP(req *model.VerifyOTPRequest) error {
	// 1. Verify OTP from Redis
	valid, err := s.otpService.VerifyOTP(req.PhoneNumber, req.OTP)
	if err != nil || !valid {
		return errors.New("invalid or expired OTP")
	}

	// 2. Mark phone as verified in database
	query := `UPDATE users SET phone_verified = true WHERE phone_number = $1`
	_, err = database.DB.Exec(query, req.PhoneNumber)
	if err != nil {
		return fmt.Errorf("failed to verify phone: %w", err)
	}

	return nil
}

// CompleteProfile completes user registration after OTP verification
// Step 3 (final) of registration process
func (s *AuthService) CompleteProfile(phoneNumber string, req *model.CompleteProfileRequest) error {
	// Update user profile with complete information
	query := `
		UPDATE users 
		SET first_name = $1, last_name = $2, age = $3, profession = $4, location = $5
		WHERE phone_number = $6 AND phone_verified = true
	`
	result, err := database.DB.Exec(
		query,
		req.FirstName,
		req.LastName,
		req.Age,
		req.Profession,
		req.Location,
		phoneNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Check if any rows were affected
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("user not found or phone not verified")
	}

	return nil
}

// Login authenticates user with email/phone and password
func (s *AuthService) Login(req *model.LoginRequest) (*model.AuthResponse, error) {
	// 1. Find user by email or phone number
	var user model.User
	
	// Check if identifier is email or phone (simple check: contains @)
	var query string
	if strings.Contains(req.Identifier, "@") {
		query = `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	} else {
		query = `SELECT * FROM users WHERE phone_number = $1 AND deleted_at IS NULL`
	}

	err := database.DB.Get(&user, query, req.Identifier)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2. Verify password
	err = utils.VerifyPassword(user.PasswordHash, req.Password)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Check if account is active
	if user.Status != "active" {
		return nil, errors.New("account is suspended or inactive")
	}

	// 4. Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.cfg)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, s.cfg)
	if err != nil {
		return nil, err
	}

	// 5. Update last login time
	query = `UPDATE users SET last_login_at = $1 WHERE id = $2`
	database.DB.Exec(query, time.Now(), user.ID)

	// 6. Return auth response
	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &user,
	}, nil
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(req *model.ForgotPasswordRequest) error {
	// 1. Find user by email
	var user model.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, req.Email)
	if err != nil {
		// Don't reveal if email exists or not (security best practice)
		// Always return success to prevent email enumeration
		return nil
	}

	// 2. Generate reset token (random secure string)
	resetToken, err := generateSecureToken(32)
	if err != nil {
		return err
	}

	// 3. Store reset token in database (expires in 1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	query = `UPDATE users SET reset_token = $1, reset_token_expires_at = $2 WHERE id = $3`
	_, err = database.DB.Exec(query, resetToken, expiresAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// 4. Send reset email
	err = s.emailService.SendPasswordResetEmail(user.Email, resetToken)
	if err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

// ResetPassword resets user password using reset token
func (s *AuthService) ResetPassword(req *model.ResetPasswordRequest) error {
	// 1. Find user with valid reset token
	var user model.User
	query := `
		SELECT * FROM users 
		WHERE reset_token = $1 
		AND reset_token_expires_at > $2 
		AND deleted_at IS NULL
	`
	err := database.DB.Get(&user, query, req.Token, time.Now())
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid or expired reset token")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// 2. Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// 3. Update password and clear reset token
	query = `
		UPDATE users 
		SET password_hash = $1, reset_token = NULL, reset_token_expires_at = NULL 
		WHERE id = $2
	`
	_, err = database.DB.Exec(query, hashedPassword, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
