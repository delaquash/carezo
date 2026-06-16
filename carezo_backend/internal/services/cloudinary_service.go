package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/delaquash/carezo/configs"
)

type CloudinaryService struct {
	// cld is the authenticated Cloudinary client
	// It is created once and reused for every upload — no repeated auth
	cld * cloudinary.Cloudinary
}


type UploadResult struct {
	URL 		string // The URL of the uploaded file in Cloudinary stored in postgres db for retrieval
	PublicID 	string // The unique identifier for the uploaded file in Cloudinary used for deletion or further management
	Format 		string // e.g., "jpg", "png"
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
