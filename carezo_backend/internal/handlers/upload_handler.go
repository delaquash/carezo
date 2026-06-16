package handlers

import (
	"fmt"
	"net/http"

	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	cloudinaryService *services.CloudinaryService
}

func NewUploadHandler(cloudinaryService *services.CloudinaryService) *UploadHandler {
	return &UploadHandler{
		cloudinaryService: cloudinaryService,
	}
}

// UploadSingleImage handles POST /api/upload
//
// Request: multipart/form-data
//   - "image":  the file (required)
//   - "folder": Cloudinary folder (optional, default: "carezo/general")
//     Send "carezo/users" for profile pics, "carezo/reviews" for review photos
//
// Response:
//
//	{ "url": "https://res.cloudinary.com/...", "public_id": "carezo/users/abc123" }
//
// Mobile flow:
//  1. User picks image from gallery
//  2. Call POST /api/upload with the file
//  3. Store the returned url and public_id in component state
//  4. Include them in the next API call (update-profile, create-review etc.)
func (h *UploadHandler) UploadSingleImage(c *gin.Context) {
	// Parse the multipart form with a max memory of 10MB
	fileHeader, err := c.FormFile("image")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "image field is required - send as multipart/form-data with field name 'image'")
		return
	}

	// reject images larger than 10mb
	const maxSize = 10 << 20
	if fileHeader.Size > maxSize {
		response.Error(c, http.StatusBadRequest, "file size must not exceed 10MB limit")
		return
	}

	// Validate MIME type to prevent non-images such as pdfs etc disguised as images.
	// also this serves as server validation
	contentType := fileHeader.Header.Get("Content-Type")
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
		"image/webp": true,
	}
	// if the content type is not in the allowed list, reject the upload
	if !allowed[contentType] {
		response.Error(c, http.StatusBadRequest, "invalid file type - only JPEG, PNG, JPG and WEBP are allowed")
		return
	}

	// folder tells cloudinary where to organise the image into my dashboard
	folder := c.DefaultPostForm("folder", "carezo/general")

	// Open the file to get an io.Reader for streaming to Cloudinary.
	// fileHeader is just metadata — Open() gives us the actual bytes

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to open uploaded file")
		return
	}
	defer file.Close()

	// Upload to cloudinary and get the result

	result, err := h.cloudinaryService.UploadImage(file, folder)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to upload image: "+err.Error())
		return
	}

	// Return URL and public_id to mobile.
	// Mobile stores both:
	//   url       → used as the image source in the app
	//   public_id → sent to the API so it can delete the old image on update
	response.Success(c, http.StatusOK, "image uploaded successfully", gin.H{
		"url":       result.URL,
		"public_id": result.PublicID,
		"format":    result.Format,
	})
}

// UploadMultipleImages handles POST /api/upload/multiple
// Used for car images (up to 5) and review images (up to 3).
//
func (h *UploadHandler) UploadMultipleImages(c *gin.Context) {
	// parse the entire multipart form to access all uploaded files
	form, err := c.MultipartForm()

	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid multipart form data")
		return
	}

	// form>file["images"] gives us a slice of file headers for the "images" field
	files := form.File["images"]

	if len(files) == 0 {
		response.Error(c, http.StatusBadRequest, "at least one image is required")
		return
	}

	// Get max allowed files from query param, default 5 (for cars).
	// Review uploads send max=3 to enforce the 3-image limit.

	maxFiles := 5

	if c.Query("max") == "3" {
		maxFiles = 3			
	}

	if len(files) > maxFiles {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("maximum %d images allowed", maxFiles))
		return
	}

	// validate each file before uploading any
	const maxSize = 10 << 20
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
		"image/webp": true,
	}
	// validate each file's size and type before uploading any to avoid partial uploads
	
	
	for i, fh := range files {
		if fh.Size > maxSize {
			response.Error(c, http.StatusBadRequest,
				fmt.Sprintf("file %d exceeds 10MB limit", i+1))
			return
		}
		if !allowed[fh.Header.Get("Content-Type")] {
			response.Error(c, http.StatusBadRequest,
				fmt.Sprintf("file %d: only JPEG, PNG and WebP are allowed", i+1))
			return
		}
	}

	folder := c.DefaultPostForm("folder", "carezo/cars")

	// upload validated files to cloudinary
	results, err := h.cloudinaryService.UploadMultipleImages(files, folder)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to upload images: "+err.Error())
		return
	}

	// Build parallel url and public_id arrays.
	// urls[0] was uploaded with public_ids[0].
	// Keeping them parallel means the caller can delete image at index N
	// by passing public_ids[N] to the delete endpoint.

	urls := make([]string, len(results))
	publickIDs := make([]string, len(results))

	for i, r := range results {
		urls[i] = r.URL
		publickIDs[i] = r.PublicID
	}

	response.Success(c, http.StatusOK, "images uploaded successfully", gin.H {
		"counr": len(results),
		"urls":  urls,
		"public_ids": publickIDs,
	})
}
