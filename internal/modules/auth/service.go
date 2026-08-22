package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"go-fiber-starter/internal/common"
	"go-fiber-starter/internal/config"
)

// Identity는 인증된 주체의 최소 정보다. 핸들러까지 전파해 권한 검사에 사용한다.
type Identity struct {
	Username string
}

// Authenticator는 자격증명 검증만 정의하는 소비자 인터페이스다.
// 실제 사용자 스토어 연동 시 이 인터페이스를 구현으로 교체한다(구현 교체 지점).
type Authenticator interface {
	Verify(ctx context.Context, username, password string) (Identity, error)
}

// demoAuthenticator는 개발용 자격증명 검사기다. env 값과 단순 비교만 한다.
// prod에서는 DB 기반 구현(비밀번호 해시 비교)으로 교체할 것 — README 참고.
type demoAuthenticator struct {
	username, password string
}

func NewDemoAuthenticator(cfg *config.Config) Authenticator {
	return &demoAuthenticator{username: cfg.AuthDemoUsername, password: cfg.AuthDemoPassword}
}

func (d *demoAuthenticator) Verify(_ context.Context, username, password string) (Identity, error) {
	if username != d.username || password != d.password {
		return Identity{}, common.ErrInvalidCredential
	}
	return Identity{Username: username}, nil
}

type Service struct {
	auth   Authenticator
	secret []byte
	ttl    time.Duration
	now    func() time.Time // 테스트에서 시간 제어용
}

func NewService(auth Authenticator, secret string, ttl time.Duration) *Service {
	return &Service{auth: auth, secret: []byte(secret), ttl: ttl, now: time.Now}
}

// Issue는 자격증명을 검증하고 HS256 JWT를 발급한다.
func (s *Service) Issue(ctx context.Context, username, password string) (string, time.Time, error) {
	id, err := s.auth.Verify(ctx, username, password)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := s.now().Add(s.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   id.Username,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(s.now()),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return token, expiresAt, nil
}

var errInvalidToken = errors.New("invalid token")

// Verify는 토큰 서명/만료를 검증하고 Identity를 반환한다.
func (s *Service) Verify(tokenString string) (Identity, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), // 알고리즘 혼동(alg=none 등) 방지
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Identity{}, common.ErrUnauthorized.WithCause(fmt.Errorf("%w: %v", errInvalidToken, err))
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return Identity{}, common.ErrUnauthorized.WithCause(errInvalidToken)
	}
	return Identity{Username: claims.Subject}, nil
}
