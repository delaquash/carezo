package services

import (
	"fmt"
	"net/smtp"

	"github.com/delaquash/carezo/configs"
)

type EmailService struct {
	cfg *configs.Config
}

func NewEmailService(cfg *configs.Config) *EmailService {
	return &EmailService{cfg: cfg}
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

// sendEmail is the actual email sending function
func (s *EmailService) sendEmail(to, subject, body string) error {
	// Set up authentication
	auth := smtp.PlainAuth(
		"",
		s.cfg.SMTPUser,
		s.cfg.SMTPPassword,
		s.cfg.SMTPHost,
	)

	// Compose email message
	from := fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromEmail)
	message := []byte(
		fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body,
	)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, s.cfg.SMTPPort)
	err := smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}