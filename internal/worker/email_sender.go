package worker

import (
	"context"
	"fmt"

	"github.com/vibe-gaming/backend/internal/config"
	emailProvider "github.com/vibe-gaming/backend/pkg/email"
)

type emailSender struct {
	sender emailProvider.Sender
	config config.EmailConfig
}

func newEmailSender(
	sender emailProvider.Sender,
	config config.EmailConfig,
) *emailSender {
	return &emailSender{
		sender: sender,
		config: config,
	}
}

type verificationEmailInput struct {
	VerificationCode string
}

func (s *emailSender) SendUserVerificationEmail(ctx context.Context, email string, verificationCode string) error {
	subject := "Код подтверждения"

	templateInput := verificationEmailInput{verificationCode}
	sendInput := emailProvider.SendEmailInput{Subject: subject, To: email}

	if err := sendInput.GenerateBodyFromHTML(s.config.Templates.Verification, templateInput); err != nil {
		return fmt.Errorf("generate email failed: %w", err)
	}

	if err := s.sender.Send(sendInput); err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}

	return nil
}
