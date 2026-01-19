package model

import (
	"time"
)

type User struct {
	ID string `json: "id" db:"id"`

	Email        string `json:"email" db:"email"`
	PhoneNumber  string `json:"phone_number" db:"phone_number"`
	PasswordHash string `json:"-" db:"password_hash"` // Exclude from JSON responses with "-"

	GoogleID      *string `json:"google_id, omitempty" db:"google_id"`
	OauthProvider *string `json:"oauth_provider, omitempty" db:"oauth_provider"`

	FirstName       string  `json:"first_name" db:"first_name"`
	LastName        string  `json:"last_name" db:"last_name"`
	Age             *int    `json:"age, omitempty" db:"age"`
	Profession      *string `json:"profession, omitempty" db:"profession"`
	Location        *string `json:"location, omitempty" db:"location"`
	ProfileImageUrl *string `json:"profile_image_url, omitempty" db:"profile_image_url"`

	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified          bool       `json:"phone_verified" db:"phone_verified"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_tokem"`
	PhoneVerificationToken *string    `json:"-" db:"phone_verification_token"`
	OTPExpiresAt           *time.Time `json:"-" db:"otp_expires_at"`

	Status string `json:"status" db:"status"` // active, inactive, banned
	Roles  string `json:"roles" db:"roles"`   // user, admin

	// reset password
	ResetToken          *string    `json:"-" db:"reset_token"`
	ResetTokenExpiresAt *time.Time `json:"-" db:"reset_token_expires_at"`

	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTP         string `json:"otp" binding:"required,len=6"`
}

// CompleteProfileRequest is sent after OTP verification
type CompleteProfileRequest struct {
	FirstName  string  `json:"first_name" binding:"required"`
	LastName   string  `json:"last_name" binding:"required"`
	Age        int     `json:"age" binding:"required,min=18,max=120"` // Age must be 18-120
	Profession *string `json:"profession,omitempty"`
	Location   *string `json:"location,omitempty"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // Can be email OR phone number
	Password   string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}
