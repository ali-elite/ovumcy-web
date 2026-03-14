package services

import (
	"strings"
	"testing"
	"time"
)

func TestAuthAttemptPolicyKeysUseScopedHMACFingerprint(t *testing.T) {
	policy := NewAuthAttemptPolicy("login", NewAttemptLimiter(), 2, time.Hour)
	secretKey := []byte("test-secret-key")

	keys := policy.keys(secretKey, "10.0.0.1", "owner@example.com")
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "login:client:10.0.0.1" {
		t.Fatalf("unexpected client key: %q", keys[0])
	}
	if !strings.HasPrefix(keys[1], "login:identity:") {
		t.Fatalf("expected scoped identity key, got %q", keys[1])
	}
	if strings.Contains(keys[1], "owner@example.com") {
		t.Fatalf("identity key should not contain the raw email: %q", keys[1])
	}

	repeated := policy.keys(secretKey, "10.0.0.9", "owner@example.com")
	if repeated[1] != keys[1] {
		t.Fatalf("expected deterministic identity fingerprint for same secret and identity")
	}

	withDifferentSecret := policy.keys([]byte("different-secret-key"), "10.0.0.9", "owner@example.com")
	if withDifferentSecret[1] == keys[1] {
		t.Fatalf("expected identity fingerprint to change when the secret changes")
	}
}

func TestAuthAttemptPolicyKeysOmitIdentityFingerprintForBlankIdentity(t *testing.T) {
	policy := NewAuthAttemptPolicy("recovery", NewAttemptLimiter(), 2, time.Hour)

	keys := policy.keys([]byte("test-secret-key"), "127.0.0.1", "   ")
	if len(keys) != 1 {
		t.Fatalf("expected only client key for blank identity, got %d", len(keys))
	}
	if keys[0] != "recovery:client:127.0.0.1" {
		t.Fatalf("unexpected client key: %q", keys[0])
	}
}
