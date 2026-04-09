
package models

import (
	"time"
)

type User struct {
	ID string `json:"id" db:"id"`

	// Authentication
	Email        string  `json:"email" db:"email"`
	PasswordHash string  `json:"-" db:"password_hash"`

	// OAuth
	GoogleID      *string `json:"google_id,omitempty" db:"google_id"`
	OAuthProvider *string `json:"oauth_provider,omitempty" db:"oauth_provider"`

	// Profile 
	FirstName       string  `json:"first_name" db:"first_name"`
	LastName        string  `json:"last_name" db:"last_name"`
	PhoneNumber     *string `json:"phone_number,omitempty" db:"phone_number"` // Now optional
	Age             *int    `json:"age,omitempty" db:"age"`
	Profession      *string `json:"profession,omitempty" db:"profession"`
	Location        *string `json:"location,omitempty" db:"location"`
	ProfileImageURL *string `json:"profile_image_url,omitempty" db:"profile_image_url"`

	// Verification 
	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_token"`
	OTPExpiresAt           *time.Time `json:"-" db:"otp_expires_at"`

	// Account status
	Status string `json:"status" db:"status"`
	Role   string `json:"role" db:"role"`

	// Password reset
	ResetToken          *string    `json:"-" db:"reset_token"`
	ResetTokenExpiresAt *time.Time `json:"-" db:"reset_token_expires_at"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// RegisterRequest 
type RegisterRequest struct {
	FullName string `json:"fullName" binding:"required,min=2,max=200"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Country  string `json:"country" binding:"required"`
}

// VerifyOTPRequest 
type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// CompleteProfileRequest 
type CompleteProfileRequest struct {
	FirstName   string  `json:"first_name" binding:"required"`
	LastName    string  `json:"last_name" binding:"required"`
	PhoneNumber *string `json:"phone_number,omitempty"` 
	Age         int     `json:"age" binding:"required,min=18,max=120"`
	Profession  *string `json:"profession,omitempty"`
	Location    *string `json:"location,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
}

// LoginRequest
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
    Email           string `json:"email" binding:"required,email"`
    OTP             string `json:"otp" binding:"required,len=4"`
    NewPassword     string `json:"new_password" binding:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" binding:"required,min=8"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}