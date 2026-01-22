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
	cfg			*configs.Config
}

func NewAuthHandler(cfg *configs.Config) *AuthHandler {
	return &AuthHandler{
		authService:  services.NewAuthService(cfg),
		cfg:		  cfg,
	}
}

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	// Bind JSON request to struct
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
	}


// Call service to register user

err := h.authService.Register(&req)
	if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	// Return resp
response.Success(
		c, 
		http.StatusCreated, 
		"Registration successfull.Please verify OTP sent to you",
		nil,
	)
}


// Verify otp that was sent
// POST /api/auth/verify-otp
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req model.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c, 
			http.StatusBadRequest,
			"Invalid request data: "+err.Error())
			return
	}

	err := h.authService.VerifyOTP(&req)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}
	response.Success(
		c,
		http.StatusOK,
		"OTP verified successfully.Please complete your profile",
		nil,
	)
}

// POST /api/auth/complete-profile
func (h *AuthHandler) CompleteProfile(c *gin.Context) {
	// Get phone nu,ber from query parameter (this is from otp verification)
	phoneNumber := c.Query("phone")

	if phoneNumber == "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"Phone Number is required",
		)
		return
	}

	var req model.CompleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"Invalid request data: " +err.Error(),
		)
		return
	}
	err := h.authService.CompleteProfile(phoneNumber, &req)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}
	response.Success(
		c,
		http.StatusOK,
		"Profile Completed Successfuly. You can login now",
		nil,
	)
}


// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"Invalid request data: "+err.Error(),
		)
		return
	}
	authResponse, err := h.authService.Login(&req)
	if err != nil {
		response.Error(
			c,
			http.StatusUnauthorized,
			err.Error(),
		)
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"Login Successful",
		authResponse,
	)
}

// POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context){
	var req model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err!= nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"Invalid request data: "+err.Error(),
		)
		return
	}
	err := h.authService.ForgotPassword(&req)
	if err != nil {
		// error is logged
		c.Error(err)
	}


	response.Success(c, http.StatusOK, "If your email exists, you will receive a password reset link.", nil)
}


// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"Invalid Request data: "+err.Error())
			return
	}
	err := h.authService.ResetPassword(&req)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest, err.Error())
		return
	}
	response.Success(
		c,
		http.StatusOK,
		"Password reset successful. You can login with your new password", nil,
	)

}