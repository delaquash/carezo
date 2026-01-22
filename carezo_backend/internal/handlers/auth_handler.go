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