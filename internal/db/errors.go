package db

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type UniqueConstraintError struct {
	Constraint string
	Err        error
}

func (err *UniqueConstraintError) Error() string {
	if strings.TrimSpace(err.Constraint) == "" {
		return "unique constraint violation"
	}
	return "unique constraint violation: " + err.Constraint
}

func (err *UniqueConstraintError) Unwrap() error {
	return err.Err
}

func (err *UniqueConstraintError) UniqueConstraint() string {
	return err.Constraint
}

type SymptomSeedError struct {
	Err error
}

func (err *SymptomSeedError) Error() string {
	return "symptom seed write failed"
}

func (err *SymptomSeedError) Unwrap() error {
	return err.Err
}

func (err *SymptomSeedError) SymptomSeedFailure() bool {
	return true
}

func classifyUserCreateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return &UniqueConstraintError{
			Constraint: "users.email",
			Err:        err,
		}
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(message, "unique constraint failed") {
		const marker = "unique constraint failed:"
		constraint := "users.email"
		index := strings.Index(message, marker)
		if index >= 0 {
			extracted := strings.TrimSpace(message[index+len(marker):])
			if extracted != "" {
				constraint = extracted
			}
		}
		return &UniqueConstraintError{
			Constraint: constraint,
			Err:        err,
		}
	}

	return err
}
