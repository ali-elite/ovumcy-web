package db

import "gorm.io/gorm"

type Repositories struct {
	Users          *UserRepository
	OIDCIdentities *OIDCIdentityRepository
	DailyLogs      *DailyLogRepository
	Symptoms       *SymptomRepository
}

func NewRepositories(database *gorm.DB) *Repositories {
	return &Repositories{
		Users:          NewUserRepository(database),
		OIDCIdentities: NewOIDCIdentityRepository(database),
		DailyLogs:      NewDailyLogRepository(database),
		Symptoms:       NewSymptomRepository(database),
	}
}
