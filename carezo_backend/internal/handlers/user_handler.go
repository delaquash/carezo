package handlers

import (
	"net/http"

	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)


type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

// GET /api/user/me 

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User profile retrieved successfully", user)
}

// PUT /api/user/complete-profile
func (h *UserHandler) CompleteUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var reg models.CompleteProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	user, err := h.userService.CompleteProfile()(userID, (string), &req)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile Completed successfully", user)
}