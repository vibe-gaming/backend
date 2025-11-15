package task

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	SendEmailTaskName  = "sendEmailTask"
	SendEmailQueueName = "sendEmailQueue"
)

type SendEmail struct {
	Email            string `json:"email"`
	VerificationCode string `json:"verification_code"`
}

func NewSendEmailTask(email string, verificationCode string) (*asynq.Task, error) {
	var data SendEmail
	data.Email = email
	data.VerificationCode = verificationCode

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("json data marshal failed: %w", err)
	}

	return asynq.NewTask(
		SendEmailTaskName,
		payload,
		asynq.MaxRetry(5),
		asynq.Queue(SendEmailQueueName),
	), nil
}
