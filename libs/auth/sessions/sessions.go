package sessions

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tobyrushton/padel-stats/libs/db/models"
)

var (
	ErrInvalidToken    = errors.New("invalid jwt token")
	ErrExpiredToken    = errors.New("jwt token has expired")
	ErrRevokedSession  = errors.New("session has been revoked")
	ErrSessionNotFound = errors.New("session not found")
)

type Claims struct {
	UserID int64 `json:"uid"`

	jwt.RegisteredClaims
}

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../../fakes/session-store.go . SessionStore
type SessionStore interface {
	Create(ctx context.Context, session *models.Session) error
	FindByTokenID(ctx context.Context, tokenID string) (*models.Session, error)
	RevokeByTokenID(ctx context.Context, tokenID string, revokedAt time.Time) error
}

type Service struct {
	store     SessionStore
	secret    []byte
	issuer    string
	sessionTT time.Duration
}

func NewService(store SessionStore, secret, issuer string, sessionTTL time.Duration) (*Service, error) {
	if store == nil {
		return nil, errors.New("session store is required")
	}
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	if issuer == "" {
		return nil, errors.New("jwt issuer is required")
	}
	if sessionTTL <= 0 {
		return nil, errors.New("session ttl must be greater than 0")
	}

	return &Service{
		store:     store,
		secret:    []byte(secret),
		issuer:    issuer,
		sessionTT: sessionTTL,
	}, nil
}

func (s *Service) Create(ctx context.Context, userID int64) (string, error) {
	tokenID, err := newTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.sessionTT)

	session := &models.Session{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}

	if err := s.store.Create(ctx, session); err != nil {
		return "", err
	}

	token, err := s.SignJWT(session.UserID, session.TokenID, session.ExpiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Validate(ctx context.Context, tokenString string) (*models.Session, error) {
	claims, err := s.parseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	session, err := s.store.FindByTokenID(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	now := time.Now().UTC()
	if session.RevokedAt != nil {
		return nil, ErrRevokedSession
	}
	if now.After(session.ExpiresAt.UTC()) {
		return nil, ErrExpiredToken
	}

	return session, nil
}

func (s *Service) Revoke(ctx context.Context, tokenID string) error {
	now := time.Now().UTC()
	err := s.store.RevokeByTokenID(ctx, tokenID, now)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SignJWT(userID int64, tokenID string, expiresAt time.Time) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expiresAt.UTC()),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

func (s *Service) parseJWT(tokenString string) (*Claims, error) {
	claims := new(Claims)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func newTokenID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
