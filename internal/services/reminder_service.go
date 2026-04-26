package services

import (
	"fmt"
	"time"

	"github.com/ovumcy/ovumcy-web/internal/db"
	"github.com/ovumcy/ovumcy-web/internal/i18n"
	"github.com/ovumcy/ovumcy-web/internal/models"
)

type ReminderService struct {
	userRepo *db.UserRepository
	dayRepo  *db.DailyLogRepository
	pushSvc  *PushService
	i18n     *i18n.Manager
	location *time.Location
}

func NewReminderService(userRepo *db.UserRepository, dayRepo *db.DailyLogRepository, pushSvc *PushService, i18n *i18n.Manager, location *time.Location) *ReminderService {
	return &ReminderService{
		userRepo: userRepo,
		dayRepo:  dayRepo,
		pushSvc:  pushSvc,
		i18n:     i18n,
		location: location,
	}
}

// CheckAndSendDailyReminders evaluates all active users. If they haven't logged anything today, send a reminder.
func (s *ReminderService) CheckAndSendDailyReminders() error {
	if !s.pushSvc.IsConfigured() {
		return nil
	}

	users, err := s.userRepo.ListAllWithRole(models.RoleOwner)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	loc := s.location
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.AddDate(0, 0, 1)

	for _, user := range users {
		_, found, err := s.dayRepo.FindByUserAndDayRange(user.ID, dayStart, dayEnd)
		if err != nil {
			// Skip this user on error, continue with others
			continue
		}
		if found {
			// Already logged today — no reminder needed
			continue
		}

		payload := PushPayload{
			Title: "Time to log your day! 🌸",
			Body:  "Don't forget to track your period, symptoms, and mood today.",
			URL:   "/dashboard",
		}
		_ = s.pushSvc.SendNotification(user.ID, payload)
	}

	return nil
}
