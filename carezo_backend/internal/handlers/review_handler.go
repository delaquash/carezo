package handlers

import (
	"fmt"
	"net/http"

	// "github.com/delaquash/carezo/internal/response"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService     *services.ReviewService
	cloudinaryService services.CloudinaryServiceInterface
}

func NewReviewHandler(cloudinaryService services.CloudinaryServiceInterface) *ReviewHandler {
	return &ReviewHandler{
		reviewService:     services.NewReviewService(),
		cloudinaryService: cloudinaryService,
	}
}

// POST /api/reviews — multipart, not JSON, since images ride along.
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := c.GetString("user_id")

	req := models.CreateReviewRequest{
		BookingID: c.PostForm("booking_id"),
	}
	fmt.Sscanf(c.PostForm("rating"), "%d", &req.Rating)
	fmt.Sscanf(c.PostForm("punctuality_rating"), "%d", &req.PunctualityRating)
	fmt.Sscanf(c.PostForm("professionalism_rating"), "%d", &req.ProfessionalismRating)
	fmt.Sscanf(c.PostForm("vehicle_condition_rating"), "%d", &req.VehicleConditionRating)
	title := c.PostForm("title")
	comment := c.PostForm("comment")
	req.Title = &title
	req.Comment = &comment

	form, _ := c.MultipartForm()
	var imageURLs, imagePublicIDs []string

	if form != nil {
		files := form.File["images"]
		if len(files) > 3 {
			response.Error(c, http.StatusBadRequest, "maximum of 3 images allowed")
			return
		}
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				response.Error(c, http.StatusBadRequest, "failed to read image: "+err.Error())
				return
			}
			result, err := h.cloudinaryService.UploadImage(file, "reviews")
			file.Close()
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "failed to upload image: "+err.Error())
				return
			}
			imageURLs = append(imageURLs, result.URL)
			imagePublicIDs = append(imagePublicIDs, result.PublicID)
		}
	}

	review, err := h.reviewService.CreateReview(userID, &req, imageURLs, imagePublicIDs)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Review created successfully", review)
}

// PUT /api/reviews/:id — full edit, JSON, no images (use EditReviewImage for that)
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	reviewID := c.Param("id")
	userID := c.GetString("user_id")
	role := c.GetString("user_role")

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	review, err := h.reviewService.UpdateReview(reviewID, userID, role == "admin", &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Review updated successfully", review)
}

func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	reviewID := c.Param("id")
	review, err := h.reviewService.GetReviewByID(reviewID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Review retrieved successfully", review)
}

// review_handler.go
func (h *ReviewHandler) GetCarReviews(c *gin.Context) {
	carID := c.Param("id")
	reviews, err := h.reviewService.GetCarReviews(carID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Reviews retrieved successfully", gin.H{
		"car_id":  carID,
		"reviews": reviews,
		"total":   len(reviews),
	})
}

// EditReviewImage — assumed to already exist correctly on this handler
// per your earlier route logs; not rewritten here since I haven't seen
// it and it wasn't flagged as broken.
