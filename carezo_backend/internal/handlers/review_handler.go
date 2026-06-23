package handlers

import (
	"fmt"
	"net/http"

	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService     *services.ReviewService
	cloudinaryService *services.CloudinaryService
}

func NewReviewHandler(cloudinaryService *services.CloudinaryService) *ReviewHandler {
	return &ReviewHandler{
		reviewService:     services.NewReviewService(),
		cloudinaryService: cloudinaryService,
	}
}

type EditReviewImagesRequest struct {
	// NewImages are Cloudinary URLs to ADD to the review.
	// omitempty — a request might ONLY remove images, no additions.
	NewImages []string `json:"new_images,omitempty"`

	// NewImagePublicIDs parallel to NewImages — validated equal length in handler.
	NewImagePublicIDs []string `json:"new_image_public_ids,omitempty"`

	// RemoveImagePublicIDs are public_ids of images to DELETE from the review.
	// omitempty — a request might ONLY add images, no removals.
	RemoveImagePublicIDs []string `json:"remove_image_public_ids,omitempty"`
}

func (h *ReviewHandler) EditReviewImage(c *gin.Context) {
	reviewID := c.Param("id")

	// get requesting userID from from JWT middleware
	userID, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get the user's role from the JWT middleware to check for admin privileges.
	// WHY check role here: the AuthMiddleware already decoded the JWT and
	// set both user_id and user_role in the gin Context. We reuse that
	// instead of querying the DB again to check if this user is an admin.

	userRole, _ := c.Get("user_role")
	isAdmin := userRole == "admin"

	var req EditReviewImagesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Validate that at least ONE operation is being requested.
	// WHY: an empty request (no new images, no removals) is a no-op and
	// likely indicates a client bug — better to reject it clearly.

	if len(req.NewImages) == 0 && len(req.RemoveImagePublicIDs) == 0 {
		response.Error(c, http.StatusBadRequest, "provide new images to add or remove_image_publics_ids to remove")
		return
	}

	// Validate parallel arrays — same length check as everywhere else
	// images/public_ids appear in this codebase.
	if len(req.NewImages) != len(req.NewImagePublicIDs) {
		response.Error(c, http.StatusBadRequest,
			"new_images and new_image_public_ids must have the same number of items")
		return
	}

	// Call the service. It handles:
	//   - fetching the review
	//   - checking ownership (unless isAdmin)
	//   - merging new images with existing ones minus removed ones
	//   - enforcing the 3-image max
	//   - updating the DB
	// It returns the updated review AND the list of public_ids that were
	// actually removed (so we know what to delete from Cloudinary here).

	updatedReview, removePublicIDs, err := h.reviewService.EditReviewImages(
		reviewID,
		userID.(string),
		isAdmin,
		req.NewImages,
		req.NewImagePublicIDs,
		req.RemoveImagePublicIDs,
	)

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// delete the remove images from Cloudinary after the DB update succeeded
	// WHY here and not in the service: the service layer should not know
	// about Cloudinary at all — that's a handler/infrastructure concern.
	// This keeps the service testable without mocking Cloudinary.
	if len(removePublicIDs) > 0 {
		go func(ids []string) {
			for _, id := range ids {
				if err := h.cloudinaryService.DeleteImage(id); err != nil {
					fmt.Printf("[Cloudinary] failed to delete review image %s: %v\n", id, err)
				}
			}
		}(removePublicIDs)
	}
	response.Success(c, http.StatusOK, "review images updated successfully", updatedReview)
}

// GetReview handles GET /api/reviews/:id
// Simple read endpoint — no Cloudinary interaction, included here for completeness
// since it lives alongside EditReviewImages in the same handler.

func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	// get review id from URL
	reviewID := c.Param("id")

	if reviewID == "" {
		response.Error(c, http.StatusBadRequest, "review ID is required")
		return
	}

	review, err := h.reviewService.GetReviewByID(reviewID)

	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(
		c,
		http.StatusOK,
		"review retrieved successfully",
		review,
	)
}
