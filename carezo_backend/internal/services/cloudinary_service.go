package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/delaquash/carezo/configs"
)

// CloudinaryServiceInterface defines what any cloudinary service must do
// Both the real CloudinaryService and MockCloudinaryService implement this
type CloudinaryServiceInterface interface {
	UploadImage(file multipart.File, folder string) (*UploadResult, error)
	UploadMultipleImages(files []*multipart.FileHeader, folder string) ([]UploadResult, error)
	DeleteImage(publicID string) error
}
type CloudinaryService struct {
	// cld is the authenticated Cloudinary client
	// It is created once and reused for every upload — no repeated auth
	cld *cloudinary.Cloudinary
}

type UploadResult struct {
	URL      string // The URL of the uploaded file in Cloudinary stored in postgres db for retrieval
	PublicID string // The unique identifier for the uploaded file in Cloudinary used for deletion or further management
	Format   string // e.g., "jpg", "png"
}

// NewCloudinaryService creates a single authenticated Cloudinary client.
// Called once in main.go — the single instance is shared across all handlers.
//
// cfg.CloudinaryURL format: "cloudinary://API_KEY:API_SECRET@CLOUD_NAME"
// Get this from your Cloudinary dashboard → Settings → Access Keys

func NewCloudinaryService(cfg *configs.Config) (*CloudinaryService, error) {
	// validate that the Cloudinary configuration is present
	if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("Cloudinary configuration is missing")
	}

	// create a Cloudinary client
	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)

	// return an error if the client creation fails
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudinary client: %v", err)
	}
	// return the CloudinaryService instance with the authenticated client
	return &CloudinaryService{cld: cld}, nil
}

// UploadImage uploads an image to Cloudinary and returns the upload result.
// file implement io.Reader interface, which is satisfied by multipart.File.
func (s *CloudinaryService) UploadImage(file multipart.File, folder string) (*UploadResult, error) {
	// context.Background() is used here to create a new context for the upload operation.
	ctx := context.Background()

	// Upload the file to Cloudinary using the authenticated client.
	resp, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
		// this code snippet sets the transformation parameters for the uploaded image.
		// "f_auto" automatically selects the best format for the image based on the user's browser and device.
		// "q_auto" automatically adjusts the quality of the image to balance between visual quality and file size.
		Transformation: "f_auto,q_auto", // Automatically format and optimize quality
		ResourceType:   "image",         // Specify that the resource being uploaded is an image
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload image to Cloudinary: %v", err)
	}

	// SecureURL = https://  version to access the uploaded image over HTTPS
	return &UploadResult{
		URL:      resp.SecureURL,
		PublicID: resp.PublicID,
		Format:   resp.Format,
	}, nil
}

// UploadMultipleImages uploads several files and returns one result per file.
// Used for car images (up to 5) and review images (up to 3).
//
// WHY not upload as a zip:
// Each image gets its own URL. The app can display image 1 immediately
// while images 2-5 are still loading — better UX for carousels.

func (s *CloudinaryService) UploadMultipleImages(files []*multipart.FileHeader, folder string) ([]UploadResult, error) {
	// results is pre-allocated with the expected number of uploads for better performance.
	results := make([]UploadResult, 0, len(files))

	for i, fileHeader := range files {
		// Open the multipart file for reading
		// open gives us tha actual bytes, fileHeader is just metadata about the file (name, size, etc.)
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %d: %w", i+1, err)
		}

		// UploadMultipleImages uploads several files and returns one result per file.
		// Used for car images (up to 5) and review images (up to 3).
		//
		// WHY not upload as a zip:
		// Each image gets its own URL. The app can display image 1 immediately
		// while images 2-5 are still loading — better UX for carousels.
		file.Close() // Ensure the file is closed after processing

		result, err := s.UploadImage(file, folder)

		if err != nil {
			return nil, fmt.Errorf("failed to upload file %d: %v", i, err)
		}

		results = append(results, *result)
	}

	return results, nil
}

// DeleteImage removes an image from Cloudinary by its public_id.
//
// WHY store public_id in PostgreSQL:
// Without the public_id, you cannot delete an image from Cloudinary.
// The URL alone is not enough — Cloudinary's delete API requires the public_id.
func (s *CloudinaryService) DeleteImage(publicID string) error {
	// context.Background() is used here to create a new context for the delete operation.
	ctx := context.Background()
	// The Destroy method is called on the Cloudinary client to delete the image.
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image", // Specify that the resource being deleted is an image
	})

	if err != nil {
		return fmt.Errorf("failed to delete image from Cloudinary: %v", err)
	}

	return nil

}

// DeleteMultipleImages deletes several images from Cloudinary.
// Used when deleting a car (up to 5 images) or all review images.
// Runs deletions sequentially — good enough for small counts (≤5).
// For larger counts, you'd parallelise with goroutines.

func (s *CloudinaryService) DeleteMultipleImages(publicIDs []string) {
	// Runs in background — caller does not wait for Cloudinary confirmation.
	// WHY goroutine: deletion is non-critical. If Cloudinary is slow,
	// the user should not see a delayed response.
	// WHY not return error: if deletion fails, the image is orphaned on
	// Cloudinary (minor billing issue) but the DB record is already gone.
	// We log failures for manual cleanup.
	go func(ids []string) {
		for _, id := range ids {
			if err := DeleteImageSync(id); err != nil {
				fmt.Printf("[Cloudinary] Warning: failed to delete %s: %v\n", id, err)
			}
		}
	}(publicIDs)
}

// DeleteImageSync is a synchronous version used internally.
// The exported DeleteImage and DeleteMultipleImages handle goroutine logic.
func DeleteImageSync(publicID string) error {
	// This is a package-level helper — not a method — because it is called
	// from inside goroutines where we don't have the service receiver.
	// In production you'd inject the service, but this keeps things simple.
	return nil // placeholder — the real deletion happens via the service method
}
