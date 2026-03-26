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
	"github.com/uptrace/bun"
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

type Service struct {
	db        *bun.DB
	secret    []byte
	issuer    string
	sessionTT time.Duration
}

func NewService(db *bun.DB, secret, issuer string, sessionTTL time.Duration) (*Service, error) {
	if db == nil {
		return nil, errors.New("db is required")
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
		db:        db,
		secret:    []byte(secret),
		issuer:    issuer,
		sessionTT: sessionTTL,
	}, nil
}

func (s *Service) Create(ctx context.Context, userID int64) (*models.Session, string, error) {
	tokenID, err := newTokenID()
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.sessionTT)

	session := &models.Session{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}

	if _, err := s.db.NewInsert().Model(session).Exec(ctx); err != nil {
		return nil, "", err
	}

	token, err := s.signJWT(session.UserID, session.TokenID, session.ExpiresAt)
	if err != nil {
		return nil, "", err
	}

	return session, token, nil
}

func (s *Service) Validate(ctx context.Context, tokenString string) (*models.Session, error) {
	claims, err := s.parseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	session := new(models.Session)
	err = s.db.NewSelect().
		Model(session).
		Where("token_id = ?", claims.ID).
		Limit(1).
		Scan(ctx)
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
	_, err := s.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked_at = ?", now).
		Set("updated_at = ?", now).
		Where("token_id = ?", tokenID).
		Where("revoked_at IS NULL").
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) signJWT(userID int64, tokenID string, expiresAt time.Time) (string, error) {
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
