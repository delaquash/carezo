package handlers

import (
	"net/http"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
	// "golang.org/x/tools/go/cfg"
)


type AuthHandler struct {
	authService *services.AuthService
	cfg         *configs.Config
}

func NewAuthHandler(cfg *configs.Config) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(cfg),
		cfg:         cfg,
	}
}

// POST /api/auth/register
// Body: {"email": "user@example.com", "password": "SecurePass123!"}
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	err := h.authService.Register(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Registration successful. Please check your email for verification code.", nil)
}


// POST /api/auth/verify-otp
// Body: {"email": "user@example.com", "otp": "123456"}
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	err := h.authService.VerifyOTP(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Email verified successfully. Please complete your profile.", nil)
}


// POST /api/auth/complete-profile?email=user@example.com
// Body: {"first_name": "John", "last_name": "Doe", "age": 28, ...}
func (h *AuthHandler) CompleteProfile(c *gin.Context) {
	// Get email from query parameter
	email := c.Query("email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "Email is required")
		return
	}

	var req models.CompleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	err := h.authService.CompleteProfile(email, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile completed successfully. You can now login.", nil)
}

// POST /api/auth/login
// Body: {"email": "user@example.com", "password": "SecurePass123!"}
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", authResponse)
}


// POST /api/auth/forgot-password
// Body: {"email": "user@example.com"}
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	err := h.authService.ForgotPassword(&req)
	if err != nil {
		c.Error(err)
	}

	response.Success(c, http.StatusOK, "If your email exists, you will receive a password reset link.", nil)
}


// POST /api/auth/reset-password
// Body: {"token": "abc123...", "new_password": "NewPass456!"}
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	err := h.authService.ResetPassword(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Password reset successful. You can now login with your new password.", nil)
}