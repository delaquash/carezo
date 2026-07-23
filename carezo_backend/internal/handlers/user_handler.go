package handlers

import (
	"fmt"
	"net/http"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService       *services.UserService
	cloudinaryService services.CloudinaryServiceInterface // needed to delete OLD profile pictures
}

func NewUserHandler(cloudinaryService services.CloudinaryServiceInterface) *UserHandler {
	return &UserHandler{
		userService:       services.NewUserService(database.DB),
		cloudinaryService: cloudinaryService,
	}
}

// GET /api/user/me

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id") // set by AuthMiddleware after decoding the JWT
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
	// straight pass-through — no Cloudinary cleanup needed on first-time setup

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

	// a map lets the client send ONLY the fields they want to change —
	// e.g. just {"first_name": "New Name"} without touching anything else
	var updates map[string]interface{}

	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Request: "+err.Error())
		return
	}
	// pull out old_profile_image_public_id BEFORE it reaches the DB layer —
	// it's not a real column, it only exists to tell US what to delete from Cloudinary
	oldPublicID, _ := updates["old_profile_image_public_id"].(string)  //type assert
	delete(updates, "old_profile_image_publi_id")  // this is to remove it so it wont reach sql builder
	// write whatever is left in the map to postgresql
	user, err := h.userService.UpdateProfile(userID.(string), updates)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	// check if the new request actually include a new profile_image_url
	_, newProfileImageProvided := updates["profile_image_url"]
	
	// only delete the old photo if a new one was set and we have it in public_ids
	if newProfileImageProvided && oldPublicID != "" {
		go func(id string) { //background  -- dont make user wait for cloudinary
			if err := h.cloudinaryService.DeleteImage(id); err != nil {
				fmt.Printf("[Cloudinary] failed to delete old profile image  %s: %v\n", id, err)
			}

		}(oldPublicID)
	}
	response.Success(c, http.StatusOK, "Profile Updated Successfully", user)
}

// DELETE   /api/user/delete-account
// DELETE /api/user/delete-user — deactivates (not permanently deletes) the account
// photo is deliberately KEPT in case the user reactivates later
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
	Status string `form:"status"`
	Role   string `form:"role"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}

func (h *UserHandler) ListAllUsers(c *gin.Context) {
	var req ListUsersRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid query params: "+err.Error())
		return
	}

	// default query params
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Limit > 20 {
		req.Limit = 20
	}

	users, total, err := h.userService.GetAllUsers(req.Status, req.Role, req.Page, req.Limit)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// calculate total pages

	totalPages := total / req.Limit

	if total%req.Limit != 0 {
		totalPages++  // round up for a partial last page
	}

	response.Success(c, http.StatusOK, "Users retrieved successfully", gin.H{
		"users": users,
		"meta": gin.H{
			"total":       total,
			"page":        req.Page,
			"limit":       req.Limit,
			"total_pages": totalPages,
		},
	})
}

// GET  /api/admin/users/:id --- get user by id

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUserByID(userID)

	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", user)
}

// PUS  /api/admin/users/:id/status
// Thois is to update user status (active, inactive, suspended) by admin

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"` // "active" | "suspended" etc, required field
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	err := h.userService.UpdateUserStatus(userID, req.Status)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User status updated successfully", nil)
}
