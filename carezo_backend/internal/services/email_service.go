package services

import (
	"crypto/tls"
	"fmt"
	"strconv"
    "gopkg.in/gomail.v2"
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
    port, err := strconv.Atoi(s.cfg.SMTPPort)
    if err != nil {
        return fmt.Errorf("invalid SMTP port: %w", err)
    }

    m := gomail.NewMessage()
    m.SetHeader("From", fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromEmail))
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)

    d := gomail.NewDialer(s.cfg.SMTPHost, port, s.cfg.SMTPUser, s.cfg.SMTPPassword)
    d.TLSConfig = &tls.Config{ServerName: s.cfg.SMTPHost}
    d.SSL = true // ✅ port 465 uses SSL directly

    if err := d.DialAndSend(m); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    return nil
}