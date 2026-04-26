package db

import "gorm.io/gorm"

type Repositories struct {
	Users              *UserRepository
	OIDCIdentities     *OIDCIdentityRepository
	OIDCLogout         *OIDCLogoutStateRepository
	DailyLogs          *DailyLogRepository
	Symptoms           *SymptomRepository
	PartnerInvitations *PartnerInvitationRepository
	PartnerLinks       *PartnerLinkRepository
	PushSubscriptions  *PushSubscriptionRepository
}

func NewRepositories(database *gorm.DB) *Repositories {
	return &Repositories{
		Users:              NewUserRepository(database),
		OIDCIdentities:     NewOIDCIdentityRepository(database),
		OIDCLogout:         NewOIDCLogoutStateRepository(database),
		DailyLogs:          NewDailyLogRepository(database),
		Symptoms:           NewSymptomRepository(database),
		PartnerInvitations: NewPartnerInvitationRepository(database),
		PartnerLinks:       NewPartnerLinkRepository(database),
		PushSubscriptions:  NewPushSubscriptionRepository(database),
	}
}
