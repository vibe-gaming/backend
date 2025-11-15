package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vibe-gaming/backend/internal/queue/task"
	"github.com/vibe-gaming/backend/internal/worker"

	"github.com/hibiken/asynq"
)

type checkSocialGroupProcessor struct {
	workers *worker.Workers
}

func NewCheckSocialGroupProcessor(workers *worker.Workers) *checkSocialGroupProcessor {
	return &checkSocialGroupProcessor{
		workers: workers,
	}
}

func (p *checkSocialGroupProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var data task.CheckSocialGroup
	err := json.Unmarshal(t.Payload(), &data)
	if err != nil {
		return fmt.Errorf("process check social group task json unmarshal failed: %w", err)
	}

	if err = p.workers.SocialGroupChecker.CheckAndUpdateUserGroups(ctx, data.UserID, data.SNILS, data.Groups); err != nil {
		return fmt.Errorf("check and update user groups failed: %w", err)
	}

	return nil
}
