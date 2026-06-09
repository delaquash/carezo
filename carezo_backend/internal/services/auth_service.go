package services

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
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

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return  errors.New("password must be at least 8 characters")
	}

	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		return errors.New("password must contain at least one lowercase letter")
	}

	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		return errors.New("password must contain at least one uppercase letter")
	}

	if matched, _ := regexp.MatchString(`[0-9]`, password); !matched {
		return errors.New("password must contain at least one number")
	}

	if matched, _ := regexp.MatchString(`[^A-Za-z0-9]`, password); !matched {
		return errors.New("password must contain at least one special character (!@#$%^&*)")
	}
	return nil
}

// Register creates a new user account and sends OTP via email
// SIMPLIFIED: Email-only, no SMS
func (s *AuthService) Register(req *models.RegisterRequest) error {
	// 1. Check if user already exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL AND email_verified = TRUE)`
	err := database.DB.Get(&exists, query, req.Email)

	if err != nil {
		return fmt.Errorf("Database error: %w", err)
	}

	if !exists {
		// delete any unverified attempt qith this email
		database.DB.Exec(`DELETE FROM users WHERE email = $1 AND email_verified= false`, req.Email)
	}

	if exists {
		return errors.New("email is already registered")
	}

	// 2. Validate password strength\
	if err := validatePasswordStrength(req.Password); err != nil {
		return err 
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
			email_verified
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
func (s *AuthService) ResendOTP(req *models.ResendOTPRequest) error {
	// 1. Check if user exists and is not verified
	var user models.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, req.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Dont resend if user is verified
	if user.EmailVerified {
		return errors.New("email is already verified")
	}

	// Generate new OTp
	otp, err := s.otpService.GenerateAndStoreOTP(req.Email)
	if err != nil {
		return fmt.Errorf("Failed to generate OTP: %w", err)
	}

	// Send OTP via email
	err = s.emailService.SendOTPEmail(req.Email, otp)
	if err != nil {
		fmt.Printf("Warning: Failed to send OTP email: %v\n", err)
		return fmt.Errorf("failed to resend verification email: %w", err)
	}

	fmt.Printf("OTP resent to %s: %s\n", req.Email, otp)
	return nil
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
	var user models.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&user, query, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	otp, err := s.otpService.GenerateAndStoreOTP(req.Email)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	err = s.emailService.SendOTPEmail(req.Email, otp)
	if err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	fmt.Printf("Password reset OTP sent to %s\n", req.Email)
	return nil
}

// ResetPassword resets user password using otp
func (s *AuthService) ResetPassword(req *models.ResetPasswordRequest) error {
	// Validate password
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("Passwords do not match")
	}

	if len(req.NewPassword) < 6 {
		return errors.New("Password must be at least 6 characters")
	}

	// Verify OTP
	valid, err := s.otpService.VerifyOTP(req.Email, req.OTP)
	if err != nil || !valid {
		return errors.New("Invalid or expired OTP")
	}

	// Get user
	var user models.User
	err = database.DB.Get(&user, `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("User not found")
		}
		return fmt.Errorf("Database error: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password ONLY
	_, err = database.DB.Exec(`
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, hashedPassword, user.ID)

	if err != nil {
		return fmt.Errorf("Failed to update password: %w", err)
	}

	return nil
}
