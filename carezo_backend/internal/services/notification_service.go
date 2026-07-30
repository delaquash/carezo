package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
)

type NotificationService struct{}

func NewNotification() *NotificationService {
	return &NotificationService{}
}

// saves a notification to the Db, this is called
// internally after booking creation and payment notification

func (s *NotificationService) CreateNotification(req *models.CreateNotificationRequest) error {
	dataJSON, err := json.Marshal(req.Data)

	if err != nil {
		return fmt.Errorf("failed to marshal notification data: %w", err)
	}

	query := `
		INSERT INTO notification(user_id, title, message, type, data)
		VALUES($1, $2, $3, $4, $5)
	`

	_, err = database.DB.Exec(query,
		req.UserID,
		req.Title,
		req.Message,
		req.Type,
		dataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (s *NotificationService) GetUserNotification(userID string) ([]models.Notification, error) {
	var notifications []models.Notification

	query := `
		SELECT * FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`

	err := database.DB.Select(&notifications, query, userID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch notification: %w", err)
	}

	return notifications, nil
}

func (s *NotificationService) SendBookingCreatedNotification(userID, bookingReference string, totalAmount float64) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Booking Confirmed",
		Message: fmt.Sprintf("Your booking %s has been created successfully.", bookingReference),
		Type:    models.NotificationTypeBookingCreated,
		Data: map[string]interface{}{
			"booking_reference": bookingReference,
			"total_amount":      totalAmount,
		},
	})
}

func (s *NotificationService) SendPaymentSuccessNotification(userID, bookingReference string, amountPaid float64) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Payment Confirmed",
		Message: fmt.Sprintf("We've received your payment of ₦%.2f for booking %s.", amountPaid, bookingReference),
		Type:    models.NotificationTypeBookingCreated,
		Data: map[string]interface{}{
			"booking_reference": bookingReference,
			"amount_paid":       amountPaid,
		},
	})
}

func (s *NotificationService) SendPickUpNotification(userID, bookingReference string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Driver has picked you up",
		Message: fmt.Sprintf("Your trip for booking %s has started.", bookingReference),
		Type:    models.NotificationtypePickup,
		Data: map[string]interface{}{
			"booking_reference": bookingReference,
		},
	})
}

func (s *NotificationService) SendDropoffNotification(userID, bookingReference string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "You've arrived",
		Message: fmt.Sprintf("Your trip for booking %s is complete. Thanks for riding with Carezo.", bookingReference),
		Type:    models.NotificationTypeDropoff,
		Data: map[string]interface{}{
			"booking_reference": bookingReference,
		},
	})
}

func (s *NotificationService) SendReturnedNotification(userID, bookingReference string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Vehicle returned",
		Message: fmt.Sprintf("Booking %s has been completed. We hope you enjoyed your ride!", bookingReference),
		Type:    models.NotificationTypeReturned,
		Data: map[string]interface{}{
			"booking_reference": bookingReference,
		},
	})
}

func (s *NotificationService) SendBookingCancelledNotification(userID, bookingReference, reason string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Booking Cancelled",
		Message: fmt.Sprintf("Your booking %s has been cancelled.", bookingReference),
		Type:    models.NotificationTypeBookingCancelled,
		Data: map[string]interface{}{
			"booking_reference": bookingReference,
			"reason":            reason,
		},
	})
}

// to get the number of unread messages
func (s *NotificationService) GetUnreadCount(userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND is_read = false
	`

	err := database.DB.Get(&count, query, userID)

	if err != nil {
		return 0, fmt.Errorf("failed to count unread notification: %w", err)
	}
	return count, nil
}

// to mark all notification for users as read
func (s *NotificationService) MarkAllRead(userID string) error {
	query := `
	UPDATE notifications
	SET is_read= true, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = $1 AND is_read = false
	`
	_, err := database.DB.Exec(query, userID)

	if err != nil {
		return fmt.Errorf("Failed to mark notifications as read: %w", err)
	}
	return nil
}

func (s *NotificationService) MarkOneread(notificationID, userID string) error {
	query := `
	UPDATE notifications
	SET is_read= true, updateD_at = CURRENT_TIMESTAMP
	WHERE id =$1 AND user_id = $2

	`
	_, err := database.DB.Exec(query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("Failed to mark notification as read: %w", err)
	}

	return nil
}

func (s *EmailService) SendBookingConfirmationEmail(
	to, bookingReference string,
	pickupDate, returnDate time.Time,
	totalAmount float64,
) error {
	subject := "Booking Confirmed — " + bookingReference

	body := fmt.Sprintf(`
		<html>
		<body style="font-family: sans-serif; color: #111;">
			<h2 style="color: #16A34A;">Booking Confirmed ✅</h2>
			<p>Your booking has been confirmed. Here are your details:</p>
 
			<table style="border-collapse: collapse; width: 100%%;">
				<tr>
					<td style="padding: 8px; font-weight: bold;">Booking Reference</td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr style="background: #f9f9f9;">
					<td style="padding: 8px; font-weight: bold;">Pickup Date</td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px; font-weight: bold;">Return Date</td>
					<td style="padding: 8px;">%s</td>
				</tr>
				<tr style="background: #f9f9f9;">
					<td style="padding: 8px; font-weight: bold;">Total Paid</td>
					<td style="padding: 8px;">₦%.2f</td>
				</tr>
			</table>
 
			<p style="margin-top: 24px;">Thank you for choosing Carezo. Have a safe trip!</p>
		</body>
		</html>
	`,
		bookingReference,
		pickupDate.Format("Mon, 02 Jan 2006 15:04"),
		returnDate.Format("Mon, 02 Jan 2006 15:04"),
		totalAmount,
	)

	return s.sendEmail(to, subject, body)
}
func (s *NotificationService) DeleteNotification(notificationID, userID string) error {
	query := `
	DELETE FROM notifications
	WHERE id = $1 
		AND user_id = $2
	`

	result, err := database.DB.Exec(query, notificationID, userID)

	if err != nil {
		return fmt.Errorf("Failed to delete notification: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("Notification not found or not owned by user")
	}

	return nil
}

func (s *NotificationService) SendDriverDocumentsReceivedNotification(userID, firstName string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Documents received",
		Message: "We've received your document and they are now under review.",
		Type:    models.NotificationTypeDriverDocumentsReceived,
		Data:    map[string]interface{}{},
	})
}

func (s *NotificationService) SendDriverApprovedNotification(userID string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "You're approved!",
		Message: "Your driver application has been approved. You can now add your payout details.",
		Type:    models.NotificationTypeDriverApproved,
		Data:    map[string]interface{}{},
	})
}

func (s *NotificationService) SendDriverRejectedNotification(userID, reason string) error {
	return s.CreateNotification(&models.CreateNotificationRequest{
		UserID:  userID,
		Title:   "Application update",
		Message: "Your driver application was not approved this time.",
		Type:    models.NotificationTypeDriverRejected,
		Data:    map[string]interface{}{"reason": reason},
	})
}

func (s *NotificationService) DeleteAllNotification(userID string) error {
	query := `
	DELETE FROM notifications
	WHERE user_id = $1
	`

	_, err := database.DB.Exec(query, userID)

	if err != nil {
		return fmt.Errorf("Failed to delete notifications: %w", err)
	}
	return nil
}
