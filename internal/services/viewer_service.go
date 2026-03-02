package services

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type ViewerDayReader interface {
	FetchLogByDate(userID uint, day time.Time, location *time.Location) (models.DailyLog, error)
}

type ViewerSymptomReader interface {
	FetchSymptoms(userID uint) ([]models.SymptomType, error)
}

type ViewerService struct {
	days     ViewerDayReader
	symptoms ViewerSymptomReader
}

func NewViewerService(days ViewerDayReader, symptoms ViewerSymptomReader) *ViewerService {
	return &ViewerService{
		days:     days,
		symptoms: symptoms,
	}
}

func (service *ViewerService) FetchSymptomsForViewer(user *models.User) ([]models.SymptomType, error) {
	if !ShouldExposeSymptomsForViewer(user) {
		return []models.SymptomType{}, nil
	}
	return service.symptoms.FetchSymptoms(user.ID)
}

func (service *ViewerService) FetchDayLogForViewer(user *models.User, day time.Time, location *time.Location) (models.DailyLog, []models.SymptomType, error) {
	logEntry, err := service.days.FetchLogByDate(user.ID, day, location)
	if err != nil {
		return models.DailyLog{}, nil, err
	}

	symptoms, err := service.FetchSymptomsForViewer(user)
	if err != nil {
		return models.DailyLog{}, nil, err
	}

	return SanitizeLogForViewer(user, logEntry), symptoms, nil
}
