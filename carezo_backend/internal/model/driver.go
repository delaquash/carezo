package models

import (
	"time"
)

type Driver struct {
	ID string `json:"id" db:"id"`

	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
	Age       int    `json:"age" db:"age"`
	Gender    string `json:"gender" db:"gender"`

	State      string `json:"state" db:"state"`
	Religion   string `json:"religion,omitempty" db:"religion"`     // ✅ no space
	Complexion string `json:"complexion,omitempty" db:"complexion"` // ✅ no space
	Height     int    `json:"height" db:"height"`                   // ✅ int not string

	PhoneNumber string  `json:"phone_number" db:"phone_number"`
	Email       *string `json:"email" db:"email"`

	LicenseNumber     string    `json:"license_number" db:"license_number"`
	LicenseExpiryDate time.Time `json:"license_expiry_date" db:"license_expiry_date"`
	YearsOfExperience int       `json:"years_of_experience" db:"years_of_experience"`

	ProfileImageURL      *string `json:"profile_image_url,omitempty" db:"profile_image_url"`
	ProfileImagePublicID *string `json:"profile_image_public_id,omitempty" db:"profile_image_public_id"`

	Bio       *string `json:"bio,omitempty" db:"bio"`
	Languages JSONB   `json:"languages" db:"languages"`

	AverageRating float64 `json:"average_rating" db:"average_rating"`
	TotalReviews  int     `json:"total_reviews" db:"total_reviews"`
	TotalTrips    int     `json:"total_trips" db:"total_trips"`

	IsAvailable bool   `json:"is_available" db:"is_available"`
	Status      string `json:"status" db:"status"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateDriverRequest - for admin
type CreateDriverRequest struct {
	FirstName            string   `json:"first_name" binding:"required"`
	LastName             string   `json:"last_name" binding:"required"`
	Age                  int      `json:"age" binding:"required,min=21,max=70"`
	Gender               string   `json:"gender" binding:"required,oneof=male female"`
	State                string   `json:"state" binding:"required"`
	Religion             *string  `json:"religion,omitempty"`
	Complexion           string   `json:"complexion" binding:"required"`
	Height               int      `json:"height" binding:"required,min=140,max=220"` // cm
	PhoneNumber          string   `json:"phone_number" binding:"required"`
	Email                *string  `json:"email,omitempty"`
	LicenseNumber        string   `json:"license_number" binding:"required"`
	LicenseExpiryDate    string   `json:"license_expiry_date" binding:"required"` // ISO 8601 format
	YearsOfExperience    int      `json:"years_of_experience" binding:"required,min=0"`
	Bio                  *string  `json:"bio,omitempty"`
	Languages            []string `json:"languages,omitempty"`
	ProfileImageURL      *string  `json:"profile_image_url,omitempty"`
	ProfileImagePublicID *string  `json:"profile_image_public_id,omitempty"`
}

// UpdateDriverRequest - admin updates driver details
type UpdateDriverRequest struct {
	FirstName               *string  `json:"first_name,omitempty"`
	LastName                *string  `json:"last_name,omitempty"`
	Age                     *int     `json:"age,omitempty"`
	Gender                  *string  `json:"gender,omitempty"`
	State                   *string  `json:"state,omitempty"`
	Religion                *string  `json:"religion,omitempty"`
	Complexion              *string  `json:"complexion,omitempty"`
	Height                  *int     `json:"height,omitempty"`
	PhoneNumber             *string  `json:"phone_number,omitempty"`
	Email                   *string  `json:"email,omitempty"`
	LicenseNumber           *string  `json:"license_number,omitempty"`
	LicenseExpiryDate       *string  `json:"license_expiry_date,omitempty"`
	YearsOfExperience       *int     `json:"years_of_experience,omitempty"`
	Bio                     *string  `json:"bio,omitempty"`
	Languages               []string `json:"languages,omitempty"`
	IsAvailable             *bool    `json:"is_available,omitempty"`
	Status                  *string  `json:"status,omitempty"`
	ProfileImageURL         *string  `json:"profile_image_url,omitempty"`
	ProfileImagePublicID    *string  `json:"profile_image_public_id,omitempty"`
	OldProfileImagePublicID *string  `json:"old_profile_image_public_id,omitempty"`
}

// SearchDriversRequest - Search and filter drivers
type SearchDriversRequest struct {
	// Filters
	Gender        *string  `form:"gender"`
	State         *string  `form:"state"`
	Religion      *string  `form:"religion"`
	Complexion    *string  `form:"complexion"`
	MinAge        *int     `form:"min_age"`
	MaxAge        *int     `form:"max_age"`
	MinHeight     *int     `form:"min_height"`
	MaxHeight     *int     `form:"max_height"`
	MinExperience *int     `form:"min_experience"`
	MinRating     *float64 `form:"min_rating"`
	IsAvailable   *bool    `form:"is_available"`
	Language      *string  `form:"language"`

	// Sorting
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=average_rating years_of_experience total_trips age"`
	OrderBy string `form:"order_by" binding:"omitempty,oneof=asc desc"`

	// Pagination
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=100"`
}

// DriverListResponse - Response with pagination
type DriverListResponse struct {
	Drivers    []*Driver              `json:"drivers"`
	Pagination PaginationMeta         `json:"pagination"`
	Filters    map[string]interface{} `json:"filters_applied"`
}
