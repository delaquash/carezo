package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/delaquash/carezo/internal/utils"
	"github.com/google/uuid"
)

type DriverService struct {
	cfg                 *configs.Config
	otpService          *OTPService
	emailService        *EmailService
	notificationService *NotificationService
}

func NewDriverService(cfg *configs.Config) *DriverService {
	return &DriverService{
		cfg:                 cfg,
		otpService:          NewOTPService(cfg),
		emailService:        NewEmailService(cfg),
		notificationService: NewNotification(),
	}
}

// to create driver
func (s *DriverService) RegisterDriver(req *models.DriverRegisterRequest) (*models.Driver, error) {
	// treat these operations as one unit. Either all of them happen, or none of them happen.
	// if it doesnt happen then rollback
	tx, err := database.DB.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer tx.Rollback()

	var exists bool
	err = tx.Get(&exists, `
	  	SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`, req.Email)

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if exists {
		return nil, errors.New("an account with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)

	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	//  the LOGIN account. role='driver' is the single value that
	// makes RequireRole("driver") meaningful downstream — everything else
	// about this row is mechanically identical to a customer registering.
	userID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO users (
		id, 
		email, 
		password_hash, 
		first_name, 
		last_name, 
		role, 
		status, 
		email_verified
	)
		VALUES (
		$1, 
		$2, 
		$3, 
		$4, 
		$5, 
		'driver', 
		'active', 
		false
	)
	`, userID, req.Email, hashedPassword, req.FirstName, req.LastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver account: %w", err)
	}

	driverID := uuid.New().String()
	var driver models.Driver
	err = tx.Get(&driver, `
    INSERT INTO drivers (id, user_id, first_name, last_name, gender, phone_number, email, verification_status, is_available)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, false)
    RETURNING *
	`, driverID, userID, req.FirstName, req.LastName, req.Gender, req.PhoneNumber, req.Email, models.DriverVerificationPendingProfile)

	if err != nil {
		return nil, fmt.Errorf("failed to create driver profile: %w", err)
	}

	otp, err := s.otpService.GenerateAndStoreOTP(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate otp: %w", err)
	}

	// Commit BEFORE sending the email. The database write is the part
	// that must be correct and durable; sending an email is an external
	// side effect we can't roll back anyway. If Commit succeeds but the
	// email below fails, the account still exists correctly, and the
	// driver can recover via the EXISTING /api/auth/resend-otp endpoint —
	// no new recovery path needed.
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit driver registration: %w", err)
	}

	if err := s.emailService.SendOTPEmail(req.Email, otp); err != nil {
		log.Printf("failed to send registration OTP email to driver %s: %v", req.Email, err)
	}

	return &driver, nil

}

func (s *DriverService) CompleteDriverProfile(driverID string, req *models.CompleteDriverProfileRequest) (*models.Driver, error) {
	driver, err := s.GetDriverByID(driverID)

	if err != nil {
		return nil, err
	}

	if driver.VerificationStatus != models.DriverVerificationPendingProfile {
		return nil, fmt.Errorf("profile already completed or not eligible for this step (current status: %s)", driver.VerificationStatus)
	}

	expiryDate, err := time.Parse("2006-01-02", req.LicenseExpiryDate)
	if err != nil {
		return nil, errors.New("invalid license_expiry_date format, use YYYY-MM-DD")
	}

	if expiryDate.Before(time.Now()) {
		return nil, errors.New("driver license has expired")
	}
	languagesJSON, err := json.Marshal(req.Languages)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal languages: %w", err)
	}

	var updated models.Driver
	err = database.DB.Get(&updated, ` 
		UPDATE drivers 
		SET age=$1, gender=$2, state=$3, nationality=$4, religion=$5, 
			complexion=$6, height=$7, license_number=$8, license_expiry_date=$9, 
			years_of_experience=$10, bio=$11, languages=$12, 
			verification_status=$13, updated_at= CURRENT_TIMESTAMP
		WHERE id=$14 AND verification_status=$15 
		RETURNING *
		`,
		req.Age, req.Gender, req.State, req.Nationality, req.Religion,
		req.Complexion, req.Height, req.LicenseNumber, expiryDate,
		req.YearsOfExperience, req.Bio, languagesJSON,
		models.DriverVerificationPendingDocuments,
		driverID, models.DriverVerificationPendingProfile,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("profile could not be completed — status changed concurrently")
		}
		return nil, fmt.Errorf("failed to complete driver profile: %w", err)
	}

	return &updated, nil
}

func (s *DriverService) UploadDriverDocuments(driverID, nin, ninDocumentURL, licenseDocumentURL string) (*models.Driver, error) {
	driver, err := s.GetDriverByID(driverID)

	if err != nil {
		return nil, err
	}

	if driver.VerificationStatus == models.DriverVerificationPendingProfile {
		return nil, errors.New("please complete your profile before uploading document")
	}

	if driver.VerificationStatus != models.DriverVerificationPendingDocuments {
		return nil, fmt.Errorf("documents already submitted or not eligible for this step (current status: %s)", driver.VerificationStatus)
	}

	var updated models.Driver

	err = database.DB.Get(&updated, `
		UPDATE drivers 
		SET nin = $1, nin_document_url = $2, license_document_url = $3,
			verification_status = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND verification_status = $6
		RETURNING *
	`, nin, ninDocumentURL, licenseDocumentURL, models.DriverVerificationPendingReview,
		driverID, models.DriverVerificationPendingDocuments)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("documents could not be submitted — status changed concurrently")
		}
		return nil, fmt.Errorf("failed to upload documents: %w", err)
	}

	if driver.UserID != nil {
		if err := s.notificationService.SendDriverDocumentsReceivedNotification(*driver.UserID, driver.FirstName); err != nil {
			log.Printf("failed to send documents-received notification: %v", err)
		}
	}
	if err := s.emailService.SendDriverDocumentsReceivedEmail(driver.Email, driver.FirstName); err != nil {
		log.Printf("failed to send documents-received email: %v", err)
	}
	return &updated, nil
}

// get single driver details by ID
func (s *DriverService) GetDriverByID(driverID string) (*models.Driver, error) {
	var driver models.Driver
	query := `SELECT * FROM drivers WHERE id = $1 and deleted_at IS NULL`
	err := database.DB.Get(&driver, query, driverID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Driver not found")
		}
		return nil, fmt.Errorf("Database error: %w", err)
	}
	return &driver, nil

}

func (s *DriverService) GetDriverByUserID(userID string) (*models.Driver, error) {
	var driver models.Driver
	query := `SELECT * FROM drivers WHERE user_id = $1 AND deleted_at IS NULL`
	err := database.DB.Get(&driver, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no driver profile found for this account")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &driver, nil
}

func (s *DriverService) UpdateDriver(driverID string, req *models.UpdateDriverRequest) (*models.Driver, error) {
	// check if driver exist
	driver, err := s.GetDriverByID(driverID)

	if err != nil {
		return nil, err
	}

	if driver.VerificationStatus == models.DriverVerificationPendingProfile {
		return nil, errors.New("please complete your profile first")
	}

	// dynamic update query that only update provided fields
	var updates []string
	var args []interface{}

	argCount := 1

	if req.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argCount))
		args = append(args, *req.FirstName)
		argCount++
	}
	if req.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argCount))
		args = append(args, *req.LastName)
		argCount++
	}
	if req.Age != nil {
		updates = append(updates, fmt.Sprintf("age = $%d", argCount))
		args = append(args, *req.Age)
		argCount++
	}
	if req.Gender != nil {
		updates = append(updates, fmt.Sprintf("gender = $%d", argCount))
		args = append(args, *req.Gender)
		argCount++
	}
	if req.State != nil {
		updates = append(updates, fmt.Sprintf("state = $%d", argCount))
		args = append(args, *req.State)
		argCount++
	}
	if req.Religion != nil {
		updates = append(updates, fmt.Sprintf("religion = $%d", argCount))
		args = append(args, *req.Religion)
		argCount++
	}
	if req.Complexion != nil {
		updates = append(updates, fmt.Sprintf("complexion = $%d", argCount))
		args = append(args, *req.Complexion)
		argCount++
	}
	if req.Height != nil {
		updates = append(updates, fmt.Sprintf("height = $%d", argCount))
		args = append(args, *req.Height)
		argCount++
	}
	if req.PhoneNumber != nil {
		updates = append(updates, fmt.Sprintf("phone_number = $%d", argCount))
		args = append(args, *req.PhoneNumber)
		argCount++
	}
	if req.LicenseNumber != nil {
		updates = append(updates, fmt.Sprintf("license_number = $%d", argCount))
		args = append(args, *req.LicenseNumber)
		argCount++
	}
	if req.LicenseExpiryDate != nil {
		expiryDate, err := time.Parse("2006-01-02", *req.LicenseExpiryDate)

		if err != nil {
			return nil, errors.New("Invalid license expiry date format")
		}
		updates = append(updates, fmt.Sprintf("license_expiry_date = $%d", argCount))
		args = append(args, expiryDate)
		argCount++
	}
	if req.IsAvailable != nil {
		updates = append(updates, fmt.Sprintf("is_available = $%d", argCount))
		args = append(args, *req.IsAvailable)
		argCount++
	}
	if req.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *req.Status)
		argCount++
	}
	if req.Languages != nil {
		languagesJSON, _ := json.Marshal(req.Languages)
		updates = append(updates, fmt.Sprintf("languages = $%d", argCount))
		args = append(args, languagesJSON)
		argCount++
	}

	if req.YearsOfExperience != nil {
		updates = append(updates, fmt.Sprintf("years_of_experience = $%d", argCount))
		args = append(args, *req.YearsOfExperience)
		argCount++
	}

	if req.Nationality != nil {
		updates = append(updates, fmt.Sprintf("nationality = $%d", argCount))
		args = append(args, *req.Nationality)
		argCount++
	}

	if req.Bio != nil {
		updates = append(updates, fmt.Sprintf("bio = $%d", argCount))
		args = append(args, *req.Bio)
		argCount++
	}
	if len(updates) == 0 {
		return nil, errors.New("No fields to update")
	}

	// add updated_at
	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// add driverID to args
	args = append(args, driverID)

	// execute update fields
	query := fmt.Sprintf(
		`
		UPDATE drivers
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING *
		`, strings.Join(updates, ", "), argCount)

	var updated models.Driver
	err = database.DB.Get(&updated, query, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to update driver: %w", err)
	}
	return &updated, nil
}
// this is to review driver application for approval or rejection. 
// if rejected, a reason must be provided. 
// if approved, the driver will be notified via email and notification
func (s *DriverService) ReviewDriverApplication(driverID, adminID string, req *models.DriverReviewRequest) (*models.Driver, error) {
	driver, err := s.GetDriverByID(driverID)
	if err != nil {
		return nil, err
	}

	if driver.VerificationStatus != models.DriverVerificationPendingReview {
		return nil, fmt.Errorf("this application is not awaiting review(current status: %s)", driver.VerificationStatus)
	}

	if !req.Approved && strings.TrimSpace(req.RejectionReason) == "" {
		return nil, errors.New("a rejection reason is required when rejection an application")
	}

	newStatus := models.DriverVerificationRejected
	// rejectionReason stays a typed nil (not an empty string) on approval
	// — an approved driver's rejection_reason column should read NULL in
	// the DB, not the empty string "", since those mean different things
	// ("no reason was ever given" vs "a reason was given and it was blank")

	var rejectionReason interface{}
	if req.Approved {
		newStatus = models.DriverVerificationApproved
	} else {
		rejectionReason = req.RejectionReason
	}

	var updated models.Driver
	err = database.DB.Get(&updated, `
		UPDATE drivers
		SET verification_status = $1, rejection_reason = $2, reviewed_by = $3,
    		reviewed_at = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND verification_status = $6
		RETURNING *
	`, newStatus, rejectionReason, adminID, time.Now(), driverID, models.DriverVerificationPendingReview)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("application could not be reviewed, status changed concurrently")
		}
		return nil, fmt.Errorf("failed to review application: %w", err)
	}

	if driver.UserID == nil {
		return &updated, nil
	}

	if req.Approved {
		if err := s.notificationService.SendDriverApprovedNotification(*driver.UserID); err != nil {
			log.Printf("failed to send driver-approved notification: %v", err)
		}
		// driver.Email, driver.FirstName is used to extract driver first name and email to send it to
		if err := s.emailService.SendDriverApprovedEmail(driver.Email, driver.FirstName); err != nil {
			log.Printf("failed to send driver-approved email: %v", err)
		}
	} else {
		// driver can only reapply after 3months(thats why we have 3)
		reapplyDate := time.Now().AddDate(0, 3, 0)
		// Extract details
		if err := s.notificationService.SendDriverRejectedNotification(*driver.UserID, req.RejectionReason); err != nil {
			log.Printf("failed to send driver-rejected notification: %v", err)
		}
		// Extract driver details such as FirstName, email, rejectionreason, reapply date and put in in the mail
		// Send driver rejected email with driver.FirstName(Hello Tunde, not Hello User)
		//
		if err := s.emailService.SendDriverRejectedEmail(driver.Email, driver.FirstName, req.RejectionReason, reapplyDate); err != nil {
			log.Printf("failed to send driver-rejected email: %v", err)
		}
	}

	return &updated, nil
}

// Search for drivers and filter with pagination
func (s *DriverService) SearchDrivers(req *models.SearchDriversRequest) (*models.DriverListResponse, error) {
	var conditions []string
	var args []interface{}
	argCount := 1

	// exclude deleted acrs
	conditions = append(conditions, "deleted_at IS NULL")

	if req.Gender != nil {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argCount))
		args = append(args, *req.Gender)
		argCount++
	}
	if req.State != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(state) = LOWER($%d)", argCount))
		args = append(args, *req.State)
		argCount++
	}
	if req.Religion != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(religion) = LOWER($%d)", argCount))
		args = append(args, *req.Religion)
		argCount++
	}
	if req.Complexion != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(complexion) = LOWER($%d)", argCount))
		args = append(args, *req.Complexion)
		argCount++
	}
	if req.MinAge != nil {
		conditions = append(conditions, fmt.Sprintf("age >= $%d", argCount))
		args = append(args, *req.MinAge)
		argCount++
	}
	if req.MaxAge != nil {
		conditions = append(conditions, fmt.Sprintf("age <= $%d", argCount))
		args = append(args, *req.MaxAge)
		argCount++
	}
	if req.MinHeight != nil {
		conditions = append(conditions, fmt.Sprintf("height >= $%d", argCount))
		args = append(args, *req.MinHeight)
		argCount++
	}
	if req.MaxHeight != nil {
		conditions = append(conditions, fmt.Sprintf("height <= $%d", argCount))
		args = append(args, *req.MaxHeight)
		argCount++
	}
	if req.MinExperience != nil {
		conditions = append(conditions, fmt.Sprintf("years_of_experience >= $%d", argCount))
		args = append(args, *req.MinExperience)
		argCount++
	}
	if req.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("average_rating >= $%d", argCount))
		args = append(args, *req.MinRating)
		argCount++
	}
	if req.IsAvailable != nil {
		conditions = append(conditions, fmt.Sprintf("is_available = $%d", argCount))
		args = append(args, *req.IsAvailable)
		argCount++
	}
	if req.Language != nil {
		conditions = append(conditions, fmt.Sprintf("languages::text ILIKE $%d", argCount))
		args = append(args, "%"+*req.Language+"%")
		argCount++
	}
	// count total
	whereClause := strings.Join(conditions, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM drivers WHERE %s", whereClause)

	var total int
	err := database.DB.Get(&total, countQuery, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to count drivers: %w", err)
	}

	// Build order by

	orderBy := "created_at DESC"

	if req.SortBy != "" {
		order := "ASC"
		if req.OrderBy == "desc" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, order)
	}

	// 4. Pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	// query drivers
	query := fmt.Sprintf(`
		SELECT * FROM drivers
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argCount, argCount+1)

	args = append(args, perPage, offset)

	var drivers []*models.Driver

	err = database.DB.Select(&drivers, query, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch drivers: %w", err)
	}
	// calculate total pages
	totalPages := (total + perPage - 1) / perPage

	// response
	return &models.DriverListResponse{
		Drivers: drivers,
		Pagination: models.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Filters: map[string]interface{}{
			"gender":     req.Gender,
			"state":      req.State,
			"religion":   req.Religion,
			"complexion": req.Complexion,
		},
	}, nil
}

func (s *DriverService) GetDriverReviews(driverID string) ([]*models.Review, error) {
	var reviews []*models.Review
	query := `
		SELECT * FROM reviews
		WHERE driver_id = $1 AND status = 'published'
		ORDER BY created_at DESC
	`

	err := database.DB.Select(&reviews, query, driverID)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch reviews: %w", err)
	}
	return reviews, nil
}

// SubmitBankDetails — only reachable once approved.
func (s *DriverService) DriverSubmitBankDetails(driverID string, req *models.DriverBankDetailsRequest) (*models.Driver, error) {
	driver, err := s.GetDriverByID(driverID)

	if err != nil {
		return nil, err
	}

	if driver.VerificationStatus != models.DriverVerificationApproved {
		return nil, errors.New("bank details can only be submitted after your application is approved")
	}

	var updated models.Driver
	err = database.DB.Get(&updated, `
		UPDATE drivers
		SET bank_account_name = $1, bank_account_number = $2, bank_name = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND verification_status = $5
		RETURNING *
	`, req.BankAccountName, req.BankAccountNumber, req.BankName, driverID, models.DriverVerificationApproved)
	if err != nil {
		return nil, fmt.Errorf("failed to submit bank details: %w", err)
	}
	return &updated, nil
}

// Soft delete driver
func (s *DriverService) DeleteDriver(driverID string) error {

	// check if driver exist
	_, err := s.GetDriverByID(driverID)

	if err != nil {
		return err
	}

	// soft delete if driver exist
	query := `UPDATE drivers SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`

	result, err := database.DB.Exec(query, driverID)

	if err != nil {
		return fmt.Errorf("Failed to delete driver: %w", err)
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		return errors.New("Driver not found")
	}
	return nil
}
