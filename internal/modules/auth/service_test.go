package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-fiber-starter/internal/apperror"
	"go-fiber-starter/internal/config"
)

const (
	testSecret = "test-secret-0123456789abcdef-0123456789" // >= 32 bytes
	testTTL    = time.Hour
)

func demoCfg() *config.Config {
	return &config.Config{AuthDemoUsername: "admin", AuthDemoPassword: "admin123"}
}

func newTestService(ttl time.Duration) *Service {
	return NewService(NewDemoAuthenticator(demoCfg()), testSecret, ttl)
}

func TestIssueAndVerify_Roundtrip(t *testing.T) {
	svc := newTestService(testTTL)

	token, exp, err := svc.Issue(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if token == "" || exp.Before(time.Now()) {
		t.Fatalf("unexpected issue result: exp=%v", exp)
	}

	id, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if id.Username != "admin" {
		t.Fatalf("want subject admin, got %q", id.Username)
	}
}

func TestIssue_WrongPassword_InvalidCredentials(t *testing.T) {
	svc := newTestService(testTTL)
	_, _, err := svc.Issue(context.Background(), "admin", "wrong")
	if !errors.Is(err, apperror.ErrInvalidCredential) {
		t.Fatalf("want ErrInvalidCredential, got %v", err)
	}
}

func TestVerify_ExpiredToken_Unauthorized(t *testing.T) {
	svc := newTestService(-time.Minute) // 이미 만료된 토큰 발급
	token, _, err := svc.Issue(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := svc.Verify(token); !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized for expired token, got %v", err)
	}
}

func TestVerify_TamperedToken_Unauthorized(t *testing.T) {
	svc := newTestService(testTTL)
	token, _, _ := svc.Issue(context.Background(), "admin", "admin123")

	other := NewService(NewDemoAuthenticator(demoCfg()), "another-secret-0123456789abcdef-99", testTTL)
	if _, err := other.Verify(token); !errors.Is(err, apperror.ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized for foreign-secret token, got %v", err)
	}
}
