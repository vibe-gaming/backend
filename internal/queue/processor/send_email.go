package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vibe-gaming/backend/internal/queue/task"
	"github.com/vibe-gaming/backend/internal/worker"

	"github.com/hibiken/asynq"
)

type sendEmailProcessor struct {
	workers *worker.Workers
}

func NewSendEmailProcessor(workers *worker.Workers) *sendEmailProcessor {
	return &sendEmailProcessor{
		workers: workers,
	}
}

func (p *sendEmailProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var data task.SendEmail
	err := json.Unmarshal(t.Payload(), &data)
	if err != nil {
		return fmt.Errorf("process send email task json unmarshal failed: %w", err)
	}

	if err = p.workers.EmailSender.SendUserVerificationEmail(ctx, data.Email, data.VerificationCode); err != nil {
		return fmt.Errorf("send user verification email failed: %w", err)
	}

	return nil
}
