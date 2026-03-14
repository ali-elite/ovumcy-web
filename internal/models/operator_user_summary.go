package models

import "time"

type OperatorUserSummary struct {
	ID                  uint
	DisplayName         string
	Email               string
	Role                string
	OnboardingCompleted bool
	CreatedAt           time.Time
}
