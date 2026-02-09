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

	var req models.CompleteProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	user, err := h.userService.CompleteProfile(userID.(string), &req)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile Completed successfully", user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// bind json into a map to handle partial update
	var updates map[string]interface{}

	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Request: "+err.Error())
		return
	}

	user, err  := h.userService.UpdateProfile(userID.(string), updates)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile Updated Successfully", user)
}

// PUT /api/user/change-password

// type ChangePasswordRequest struct {
// 	CurrentPassword string `json:"current_password" binding:"required"`
// 	NewPassword     string `json:"new_password" binding:"required,min=8"`
// }

// func (h *UserHandler) ChangePassword(c *gin.Context) {
// 	userID, exists := c.Get("user_id")

// 	if !exists {
// 		response.Error(c, http.StatusUnauthorized, "Unauthorized")
// 		return
// 	}

// 	var req ChangePasswordRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
// 		return
// 	}

// 	err := h.userService.ChangePassword(userID.(string), req.CurrentPassword, req.NewPassword)
// 	if err != nil {
// 		response.Error(c, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	response.Success(c, http.StatusOK, "Password changed successfully", nil)
// }

// DELETE   /api/user/delete-account
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.userService.DeactivateAccount(userID.(string))

	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Account deactivated successfully", nil)

}



// Admin-only endpoints

// GET  /api/admin/get-all-users
//    ?status=active&role=user&page=1&limit=10

type ListUsersRequest struct {
	Status  string `form:"status"`
	Role    string `form:"role"`
	Page    int    `form:"page,default=1"`
	Limit   int    `form:"limit,default=10"`
}

func (h *UserHandler) ListAllUsers(c *gin.Context) {
	var req ListUsersRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query params: "+err.Error())
		return
	}

	// default query params
	if req.Page < 1 { req.Page = 1 }
	if req.Limit < 1 { req.Limit = 10 }
	if req.Limit > 20 { req.Limit = 20 }

	users, total, err := h.userService.GetAllUsers(req.Status, req.Role, req.Page, req.Limit)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// calculate total pages

	totalPages := total / req.Limit

	if total % req.Limit != 0 {
		totalPages++
	}

	response.Success(c, http.StatusOK, "Users retrieved successfully", gin.H {
		"users": users,
		"meta": gin.H {
			"total": total,
			"page": req.Page,
			"limit": req.Limit,
			"total_pages": totalPages,
		},
	})
}