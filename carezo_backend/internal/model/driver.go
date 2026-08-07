package models

import (
	"time"
)

const (
	DriverVerificationPendingProfile   = "pending_profile"
	DriverVerificationPendingDocuments = "pending_documents"
	DriverVerificationPendingReview    = "pending_review"
	DriverVerificationApproved         = "approved"
	DriverVerificationRejected         = "rejected"
)

type Driver struct {
	ID string `json:"id" db:"id"`

	UserID *string `json:"user_id,omitempty" db:"user_id"`

	FirstName   string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	Age         int    `json:"age" db:"age"`
	Gender      string `json:"gender" db:"gender"`
	State       string `json:"state" db:"state"`
	Nationality string `json:"nationality" db:"nationality"`
	Religion    string `json:"religion" db:"religion"`
	Complexion  string `json:"complexion" db:"complexion"`
	Height      int    `json:"height" db:"height"`

	PhoneNumber string `json:"phone_number" db:"phone_number"`
	Email       string `json:"email" db:"email"`

	LicenseNumber     string    `json:"license_number" db:"license_number"`
	LicenseExpiryDate time.Time `json:"license_expiry_date" db:"license_expiry_date"`

	NIN                *string `json:"nin,omitempty" db:"nin"`
	NINDocumentURL     *string `json:"nin_document_url,omitempty" db:"nin_document_url"`
	LicenseDocumentURL *string `json:"license_document_url,omitempty" db:"license_document_url"`

	YearsOfExperience int     `json:"years_of_experience" db:"years_of_experience"`
	Bio               *string `json:"bio,omitempty" db:"bio"`
	Languages         JSONB   `json:"languages" db:"languages"`

	ProfileImageURL      *string `json:"profile_image_url,omitempty" db:"profile_image_url"`
	ProfileImagePublicID *string `json:"profile_image_public_id,omitempty" db:"profile_image_public_id"`

	IsAvailable bool   `json:"is_available" db:"is_available"`
	Status      string `json:"status" db:"status"`

	VerificationStatus string     `json:"verification_status" db:"verification_status"`
	RejectionReason    *string    `json:"rejection_reason,omitempty" db:"rejection_reason"`
	ReviewedBy         *string    `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`

	BankAccountName   *string `json:"bank_account_name,omitempty" db:"bank_account_name"`
	BankAccountNumber *string `json:"bank_account_number,omitempty" db:"bank_account_number"`
	BankName          *string `json:"bank_name,omitempty" db:"bank_name"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateDriverRequest - for admin
type DriverRegisterRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Gender      string `json:"gender" binding:"required,oneof=male female"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email,omitempty"`
	Password    string `json:"password" binding:"required,min=8"`
}

type CompleteDriverProfileRequest struct {
	Age               int      `json:"age" binding:"required,min=21,max=70"`
	Gender            string   `json:"gender" binding:"required,oneof=male female"`
	State             string   `json:"state" binding:"required"`
	Nationality       string   `json:"nationality" binding:"required"`
	Religion          *string  `json:"religion,omitempty"`
	Complexion        string   `json:"complexion" binding:"required"`
	Height            int      `json:"height" binding:"required,min=140,max=220"`
	LicenseNumber     string   `json:"license_number" binding:"required"`
	LicenseExpiryDate string   `json:"license_expiry_date" binding:"required"` // ISO 8601, matches CreateDriverRequest's existing convention
	YearsOfExperience int      `json:"years_of_experience" binding:"required,min=0"`
	Bio               *string  `json:"bio,omitempty"`
	Languages         []string `json:"languages,omitempty"`
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
	LicenseNumber           *string  `json:"license_number,omitempty"`
	LicenseExpiryDate       *string  `json:"license_expiry_date,omitempty"`
	YearsOfExperience       *int     `json:"years_of_experience,omitempty"`
	Bio                     *string  `json:"bio,omitempty"`
	Nationality             *string  `json:"nationality,omitempty"`
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

type DriverBankDetailsRequest struct {
	BankAccountName   string `json:"bank_account_name" binding:"required"`
	BankAccountNumber string `json:"bank_account_number" binding:"required,len=10"`
	BankName          string `json:"bank_name" binding:"required"`
}

type DriverReviewRequest struct {
	Approved        bool   `json:"approved"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type DriverDocumentUploadRequest struct {
	NIN string `form:"nin" binding:"required,len=11"` // Nigerian NIN is 11 digits
}
