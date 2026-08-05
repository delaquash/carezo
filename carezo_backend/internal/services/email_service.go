package services

import (
	// "crypto/tls"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/delaquash/carezo/configs"
)

type brevoRecipient struct {
	Email string `json:"email"`
}

type brevoSender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type EmailService struct {
	cfg *configs.Config
}

func NewEmailService(cfg *configs.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

type brevoEmailPayload struct {
	Sender      brevoSender      `json:"sender"`
	To          []brevoRecipient `json:"to"`
	Subject     string           `json:"subject"`
	HTMLContent string           `json:"htmlContent"`
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	payload := brevoEmailPayload{
		Sender: brevoSender{
			Name:  s.cfg.FromName,
			Email: s.cfg.FromEmail, // must be a verified sender in Brevo — see the DKIM note below
		},
		To:          []brevoRecipient{{Email: to}},
		Subject:     subject,
		HTMLContent: body,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-key", s.cfg.BrevoAPIKey) // the API key, NOT the SMTP password

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	// Brevo returns 201 Created on success, not 200 — checking >= 400
	// (rather than != 200) correctly treats 201 as success without
	// needing to hardcode the exact expected status.
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
func (s *EmailService) SendOTPEmail(to, otp string) error {
	subject := "Your Carezo Verification Code"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to Carezo!</h2>
			<p>Your verification code is: <strong style="font-size: 24px;">%s</strong></p>
			<p>This code will expire in %d minutes.</p>
			<p>If you didn't request this code, please ignore this email.</p>
		</body>
		</html>
	`, otp, s.cfg.OTPExpirationMinutes)

	return s.sendEmail(to, subject, body)
}

func (s *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	// In production, this would be your actual website URL
	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", resetToken)

	subject := "Reset Your Carezo Password"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request this, please ignore this email.</p>
		</body>
		</html>
	`, resetLink)

	return s.sendEmail(to, subject, body)
}

func (s *EmailService) SendDriverDocumentsReceivedEmail(to, firstName string) error {
	subject := "We've received your documents"
	body := fmt.Sprintf(`
		<html><body>
		<h2>Hi %s,</h2>
		<p>We've received your NIN and driver's license. Our team will review them and get back to you soon.</p>
		</body></html>
	`, firstName)
	return s.sendEmail(to, subject, body)
}

func (s *EmailService) SendDriverApprovedEmail(to, firstName string) error {
	subject := "You're approved to drive with Carezo!"
	body := fmt.Sprintf(`
		<html><body>
		<h2>Congratulations, %s!</h2>
		<p>Your driver application has been approved. Log in to the app to add your payout account details.</p>
		</body></html>
	`, firstName)
	return s.sendEmail(to, subject, body)
}

func (s *EmailService) SendDriverRejectedEmail(to, firstName, reason string, reapplyDate time.Time) error {
	subject := "Update on your Carezo driver application"
	body := fmt.Sprintf(`
		<html><body>
		<h2>Hi %s,</h2>
		<p>After review, we're unable to approve your driver application at this time.</p>
		<p><strong>Reason:</strong> %s</p>
		<p>You're welcome to reapply starting %s.</p>
		</body></html>
	`, firstName, reason, reapplyDate.Format("January 2, 2006"))
	return s.sendEmail(to, subject, body)
}
