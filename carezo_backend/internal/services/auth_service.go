package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	"github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/utils"
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

// Register creates a new user account and sends OTP via email
// SIMPLIFIED: Email-only, no SMS
func(s * AuthService) Register(req *models.RegisterRequest) error {
	// 1. Check if user already exists
	var exists bool
	query:= `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	err := database.DB.Get(&exists, query, req.Email)

	if err != nil {
		return fmt.Errorf("Database error: %w", err)
	}

	if exists{
		return errors.New("email is already registered")
	}

	// hashed password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Split full name
	firstName, lastName := utils.SplitFullName(req.FullName)

	firstName = utils.CapitalizeName(firstName)
	lastName = utils.CapitalizeName(lastName)

	// Create user in database
	userID := uuid.New()
	query = `
	    INSERT INTO users(
			id,
			email,
			password_hash,
			first_name,
			last_name,
			oauth_provider,
			status,
			role,
			email_verified,
		)
		VALUES ($1, $2, $3, $4, $5, 'local', 'active', 'user', false)	
		`
		_, err = database.DB.Exec(
			query,
			userID,
			req.Email,
			hashedPassword,
			firstName,
			lastName,
		)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// generate and store OTP
		otp, err := s.otpService.GenerateAndStoreOTP(req.Email)
		if err != nil {
			return fmt.Errorf("failed to generate OTP: %w", err)
		}

		// send OTP email
		err = s.emailService.SendOTPEmail(req.Email, otp)
		if err != nil {
			fmt.Printf("Warning: Failed to send OTP email: %v\n", err)
			return fmt.Errorf("failed to send verification email: %w", err)
		}

		fmt.Printf("User registered with email %s. OTP sent: %s\n", req.Email, otp)
		return nil
}

// VerifyOTP verifies the OTP code and logs user in
func (s *AuthService) VerifyOTP(req *models.VerifyOTPRequest) (*models.AuthResponse, error) {
	// 1. Verify OTP from Redis
	valid, err := s.otpService.VerifyOTP(req.Email, req.OTP)
	if err != nil || !valid {
		return nil, errors.New("invalid or expired OTP")
	}

	// 2. Mark email as verified and get user
	var user models.User
	query := `
		UPDATE users 
		SET email_verified = true 
		WHERE email = $1 AND deleted_at IS NULL
		RETURNING *
	`
	err = database.DB.Get(&user, query, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// 3. Generate JWT tokens (auto-login)
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.cfg)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, s.cfg)
	if err != nil {
		return nil, err
	}

	// 4. Update last login time
	query = `UPDATE users SET last_login_at = $1 WHERE id = $2`
	database.DB.Exec(query, time.Now(), user.ID)

	// 5. Return auth response with token
	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &user,
	}, nil
}
// resend otp to user's email 
func (s *AuthService) ResendOTP(req *models.ResendOTPRequest) (*models.AuthResponse, error){
	// 1. Check if user exists and is not verified
	var user models.User
	query := `SELECT * FROM user WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, req.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Dont resend if user is verified
	if user.EmailVerified {
		return nil, errors.New("email is already verified")
	}

	// Generate new OTp
	otp, err := s.otpService.GenerateAndStoreOTP(req.Email)
	if err != nil {
		return nil, err
	}

	// Send OTP via email
	err = s.emailService.SendOTPEmail(req.Email, otp)
	if err != nil {
		fmt.Printf("Warning: Failed to send OTP email: %v\n", err)
		return nil, fmt.Errorf("failed to resend verification email: %w", err)
	}
	
	fmt.Printf("OTP resent to %s: %s\n", req.Email, otp)
	return nil, err
}

// Login authenticates user with email and password
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// 1. Find user by email
	var user models.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2. Check if email is verified
	if !user.EmailVerified {
		return nil, errors.New("please verify your email before logging in")
	}

	// 3. Verify password
	err = utils.VerifyPassword(user.PasswordHash, req.Password)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 4. Check if account is active
	if user.Status != "active" {
		return nil, errors.New("account is suspended or inactive")
	}

	// 5. Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, s.cfg)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, s.cfg)
	if err != nil {
		return nil, err
	}

	// 6. Update last login time
	query = `UPDATE users SET last_login_at = $1 WHERE id = $2`
	database.DB.Exec(query, time.Now(), user.ID)

	// 7. Return auth response
	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         &user,
	}, nil
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(req *models.ForgotPasswordRequest) error {
	// 1. Find user by email
	var user models.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, req.Email)
	if err != nil {
		// Don't reveal if email exists (security best practice)
		return nil
	}

	// 2. Generate reset token
	resetToken, err :=utils.GenerateSecureToken(32)
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
func (s *AuthService) ResetPassword(req *models.ResetPasswordRequest) error {
	// 1. Find user with valid reset token
	var user models.User
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
