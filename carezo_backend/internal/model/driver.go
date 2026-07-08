package models

import (
	"time"
)

type Driver struct {
	ID                   string     `db:"id"                    json:"id"`
	FirstName            string     `db:"first_name"            json:"first_name"`
	LastName             string     `db:"last_name"             json:"last_name"`
	Age                  int        `db:"age"                   json:"age"`
	Gender               string     `db:"gender"                json:"gender"`
	State                string     `db:"state"                 json:"state"`
	Religion             *string    `db:"religion"              json:"religion,omitempty"`
	Complexion           *string    `db:"complexion"            json:"complexion,omitempty"`
	Nationality          *string    `db:"nationality"           json:"nationality,omitempty"` // ← was string, NULL breaks scan
	Height               int        `db:"height"                json:"height"`
	PhoneNumber          string     `db:"phone_number"          json:"phone_number"`
	TotalTrips           *int       `db:"total_trips" json:"total_trips,omitempty"`
	Email                string     `db:"email"                 json:"email"`
	LicenseNumber        string     `db:"license_number"        json:"license_number"`
	LicenseExpiryDate    time.Time  `db:"license_expiry_date"   json:"license_expiry_date"`
	YearsOfExperience    int        `db:"years_of_experience"   json:"years_of_experience"`
	Bio                  *string    `db:"bio"                   json:"bio,omitempty"`
	Languages            JSONB      `db:"languages"             json:"languages"`
	IsAvailable          bool       `db:"is_available"          json:"is_available"`
	Status               string     `db:"status"                json:"status"`
	AverageRating        *float64   `db:"average_rating"        json:"average_rating,omitempty"`
	TotalReviews         *int       `db:"total_reviews"         json:"total_reviews,omitempty"`
	ProfileImageURL      *string    `db:"profile_image_url"     json:"profile_image_url,omitempty"`
	ProfileImagePublicID *string    `db:"profile_image_public_id" json:"profile_image_public_id,omitempty"`
	CreatedAt            time.Time  `db:"created_at"            json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"            json:"updated_at"`
	DeletedAt            *time.Time `db:"deleted_at"            json:"deleted_at,omitempty"`
}

// CreateDriverRequest - for admin
type CreateDriverRequest struct {
	FirstName            string   `json:"first_name" binding:"required"`
	LastName             string   `json:"last_name" binding:"required"`
	Age                  int      `json:"age" binding:"required,min=21,max=70"`
	Gender               string   `json:"gender" binding:"required,oneof=male female"`
	State                string   `json:"state" binding:"required"`
	Nationality          string   `db:"nationality" json:"nationality"`
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
