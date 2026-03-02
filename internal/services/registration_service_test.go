package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type stubRegistrationAuthService struct {
	user         models.User
	recoveryCode string
	err          error
}

func (stub *stubRegistrationAuthService) RegisterOwner(string, string, string, time.Time) (models.User, string, error) {
	if stub.err != nil {
		return models.User{}, "", stub.err
	}
	return stub.user, stub.recoveryCode, nil
}

type stubRegistrationStore struct {
	err            error
	called         bool
	lastPersisted  models.User
	lastSymptomSet []models.SymptomType
}

func (stub *stubRegistrationStore) CreateUserWithSymptoms(user *models.User, symptoms []models.SymptomType) error {
	stub.called = true
	if user != nil {
		stub.lastPersisted = *user
	}
	stub.lastSymptomSet = make([]models.SymptomType, len(symptoms))
	copy(stub.lastSymptomSet, symptoms)
	return stub.err
}

type stubRegistrationUniqueConstraintError struct{}

func (stubRegistrationUniqueConstraintError) Error() string { return "unique constraint violation" }
func (stubRegistrationUniqueConstraintError) UniqueConstraint() string {
	return "users.email"
}

type stubRegistrationSeedWriteError struct{}

func (stubRegistrationSeedWriteError) Error() string { return "symptom seed write failed" }
func (stubRegistrationSeedWriteError) SymptomSeedFailure() bool {
	return true
}

func TestRegistrationServiceRegisterOwnerAccount(t *testing.T) {
	now := time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		auth := &stubRegistrationAuthService{
			user:         models.User{ID: 42, Email: "owner@example.com"},
			recoveryCode: "OVUM-ABCD-1234-EFGH",
		}
		store := &stubRegistrationStore{}
		service := NewRegistrationService(auth, store)

		user, recoveryCode, err := service.RegisterOwnerAccount("owner@example.com", "StrongPass1", "StrongPass1", now)
		if err != nil {
			t.Fatalf("RegisterOwnerAccount() unexpected error: %v", err)
		}
		if user.ID != 42 {
			t.Fatalf("expected user id 42, got %d", user.ID)
		}
		if recoveryCode != "OVUM-ABCD-1234-EFGH" {
			t.Fatalf("expected recovery code, got %q", recoveryCode)
		}
		if !store.called || store.lastPersisted.ID != 42 {
			t.Fatalf("expected CreateUserWithSymptoms to persist user 42")
		}
		if len(store.lastSymptomSet) == 0 {
			t.Fatalf("expected builtin symptoms batch for new registration")
		}
		for _, symptom := range store.lastSymptomSet {
			if symptom.UserID != 0 {
				t.Fatalf("expected symptoms without bound user id before persistence, got %d", symptom.UserID)
			}
			if !symptom.IsBuiltin {
				t.Fatalf("expected builtin symptom in seed set")
			}
		}
	})

	t.Run("auth error propagated", func(t *testing.T) {
		authErr := ErrAuthEmailExists
		auth := &stubRegistrationAuthService{err: authErr}
		store := &stubRegistrationStore{}
		service := NewRegistrationService(auth, store)

		if _, _, err := service.RegisterOwnerAccount("owner@example.com", "StrongPass1", "StrongPass1", now); !errors.Is(err, authErr) {
			t.Fatalf("expected auth error %v, got %v", authErr, err)
		}
		if store.called {
			t.Fatalf("did not expect persistence call on auth error")
		}
	})

	t.Run("seed error mapped", func(t *testing.T) {
		auth := &stubRegistrationAuthService{
			user:         models.User{ID: 55, Email: "owner@example.com"},
			recoveryCode: "OVUM-ABCD-1234-EFGH",
		}
		store := &stubRegistrationStore{err: stubRegistrationSeedWriteError{}}
		service := NewRegistrationService(auth, store)

		if _, _, err := service.RegisterOwnerAccount("owner@example.com", "StrongPass1", "StrongPass1", now); !errors.Is(err, ErrRegistrationSeedSymptoms) {
			t.Fatalf("expected ErrRegistrationSeedSymptoms, got %v", err)
		}
		if !store.called || store.lastPersisted.ID != 55 {
			t.Fatalf("expected persistence call for user 55")
		}
	})

	t.Run("unique violation mapped to email exists", func(t *testing.T) {
		auth := &stubRegistrationAuthService{
			user:         models.User{ID: 56, Email: "owner@example.com"},
			recoveryCode: "OVUM-ABCD-1234-EFGH",
		}
		store := &stubRegistrationStore{err: stubRegistrationUniqueConstraintError{}}
		service := NewRegistrationService(auth, store)

		if _, _, err := service.RegisterOwnerAccount("owner@example.com", "StrongPass1", "StrongPass1", now); !errors.Is(err, ErrAuthEmailExists) {
			t.Fatalf("expected ErrAuthEmailExists, got %v", err)
		}
	})

	t.Run("non-unique persistence error mapped to register failed", func(t *testing.T) {
		auth := &stubRegistrationAuthService{
			user:         models.User{ID: 57, Email: "owner@example.com"},
			recoveryCode: "OVUM-ABCD-1234-EFGH",
		}
		store := &stubRegistrationStore{err: errors.New("db down")}
		service := NewRegistrationService(auth, store)

		if _, _, err := service.RegisterOwnerAccount("owner@example.com", "StrongPass1", "StrongPass1", now); !errors.Is(err, ErrAuthRegisterFailed) {
			t.Fatalf("expected ErrAuthRegisterFailed, got %v", err)
		}
	})
}
