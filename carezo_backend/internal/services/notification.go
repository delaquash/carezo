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
	return  nil
}


func (s *NotificationService) GetUserNotification(userID string)([]models.Notification, error) {
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

// to get the number of unread messages
func (s *NotificationService) GetUnreadCount(userID string)(int, error) {
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
		return  fmt.Errorf("Failed to mark notification as read: %w", err)
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