package services

import (
	"errors"
	"fmt"
	"strings"
)

var ErrAuthRegistrationDisabled = errors.New("auth registration disabled")

type RegistrationMode string

const (
	RegistrationModeOpen   RegistrationMode = "open"
	RegistrationModeClosed RegistrationMode = "closed"
)

func ParseRegistrationMode(raw string) (RegistrationMode, error) {
	mode := RegistrationMode(strings.ToLower(strings.TrimSpace(raw)))
	if mode == "" {
		return RegistrationModeOpen, nil
	}
	if err := mode.Validate(); err != nil {
		return "", err
	}
	return mode, nil
}

func (mode RegistrationMode) Validate() error {
	switch mode {
	case RegistrationModeOpen, RegistrationModeClosed:
		return nil
	default:
		return fmt.Errorf("REGISTRATION_MODE must be one of: open, closed")
	}
}

func (mode RegistrationMode) IsOpen() bool {
	return mode == RegistrationModeOpen
}
