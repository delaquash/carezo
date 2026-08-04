package services

import (
	// "uuid"

	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/delaquash/carezo/internal/database"
	models "github.com/delaquash/carezo/internal/model"
	"github.com/google/uuid"
)

type PushNotificationService struct{}

func NewPushNotificationService() *PushNotificationService {
	return &PushNotificationService{}
}

func (s *PushNotificationService) RegisterPushToken(userID string, req *models.RegisterPushTokenRequest) error {
	_, err := database.DB.Exec(`
		INSERT INTO push_tokens(id, user_id, token, platform)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT( user_id, token) DO UPDATE SET updated_at = CURRENT_TIMESTAMP

	`, uuid.New().String(), userID, req.Token, req.Platform)

	if err != nil {
		return fmt.Errorf("failed to register push token: %w", err)
	}
	return nil
}

type expoPushMessage struct {
	To    string                 `json:"to"`
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
	Sound string                 `json:"sound,omitempty"`
}

func (s *PushNotificationService) SendPushNotification(userID, title, body string, data map[string]interface{}) {
	var tokens []models.PushTokens

	err := database.DB.Select(&tokens, `SELECT * FROM push_tokens WHERE user_id = $1`, userID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("failed to fetch push tokens for user %s: %v", userID, err)
		}
		return
	}

	for _, t := range tokens {
		msg := expoPushMessage {
			To:    t.Token,
			Title: title,
			Body:  body,
			Sound: "default",
			Data:  data,
		}

		payloadBytes, err := json.Marshal(msg)
		if err != nil {
			log.Printf("failed to marshal push payload: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", "https://exp.host/--/api/v2/push/send",bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Printf("failed to create push request: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")


		client := &http.Client{Timeout: 10 * time.Second} 
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("failed to send push notification to token %s: %v", t.Token, err)
			continue // one failed device shouldn't stop the others from receiving theirs
		}
		resp.Body.Close()

	}
}