package services

import (
	"encoding/json"
	"fmt"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
)


type NotificationService struct{}

func NewNotification() *NotificationService {
	return &NotificationService{}
}

// saves a notification to the Db, this is called
// internally after booking creation and payment notification

func (s *NotificationService) CreateNotification(req models.CreateNotificationRequest) error {
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