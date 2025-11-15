package service

import "errors"

var (
	ErrUserAlreadyExist         = errors.New("user already exist")
	ErrUserNotFound             = errors.New("user not found")
	ErrVerificationCodeNotFound = errors.New("verification code not found")
)
