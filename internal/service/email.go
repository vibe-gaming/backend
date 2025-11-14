package service

import (
	"fmt"

	"github.com/vibe-gaming/backend/internal/config"
	emailProvider "github.com/vibe-gaming/backend/pkg/email"
)

type EmailService struct {
	sender  emailProvider.Sender
	config  config.EmailConfig
	enabled bool
}

func newEmailsService(sender emailProvider.Sender, config config.EmailConfig) *EmailService {
	return &EmailService{
		enabled: config.Enabled,
		sender:  sender,
		config:  config,
	}
}

type verificationEmailInput struct {
	VerificationCode string
}

type VerificationEmailInput struct {
	Email            string
	VerificationCode string
}

func (s *EmailService) SendUserVerificationEmail(input VerificationEmailInput) error {
	if !s.enabled {
		return nil
	}

	subject := "Код подтверждения"

	templateInput := verificationEmailInput{input.VerificationCode}
	sendInput := emailProvider.SendEmailInput{Subject: subject, To: input.Email}

	if err := sendInput.GenerateBodyFromHTML(s.config.Templates.Verification, templateInput); err != nil {
		return fmt.Errorf("generate email failed: %w", err)
	}

	return s.sender.Send(sendInput)
}
