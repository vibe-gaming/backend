package task

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	CheckSocialGroupTaskName  = "checkSocialGroupTask"
	CheckSocialGroupQueueName = "checkSocialGroupQueue"
)

type CheckSocialGroup struct {
	UserID uuid.UUID `json:"user_id"`
	SNILS  string    `json:"snils"`
	Groups []string  `json:"groups"` // Список типов групп для проверки
}

// NewCheckSocialGroupTask создает новую задачу для проверки социальной группы
func NewCheckSocialGroupTask(userID uuid.UUID, snils string, groups []string) (*asynq.Task, error) {
	data := CheckSocialGroup{
		UserID: userID,
		SNILS:  snils,
		Groups: groups,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("json data marshal failed: %w", err)
	}

	return asynq.NewTask(
		CheckSocialGroupTaskName,
		payload,
		asynq.MaxRetry(3),
		asynq.Queue(CheckSocialGroupQueueName),
	), nil
}
