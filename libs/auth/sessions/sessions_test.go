package sessions_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tobyrushton/padel-stats/libs/auth/sessions"
	"github.com/tobyrushton/padel-stats/libs/db/models"
	"github.com/tobyrushton/padel-stats/libs/fakes"
)

type ServiceTestSuite struct {
	suite.Suite
	store   *fakes.FakeSessionStore
	service *sessions.Service
	ctx     context.Context
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.store = new(fakes.FakeSessionStore)
	suite.ctx = context.Background()

	var err error
	suite.service, err = sessions.NewService(
		suite.store,
		"test-secret-key-32-bytes-long!!",
		"test-issuer",
		24*time.Hour,
	)
	require.NoError(suite.T(), err)
}

// NewService tests

func (suite *ServiceTestSuite) TestNewService_Success() {
	service, err := sessions.NewService(
		suite.store,
		"secret",
		"issuer",
		1*time.Hour,
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), service)
}

func (suite *ServiceTestSuite) TestNewService_NilStore() {
	service, err := sessions.NewService(
		nil,
		"secret",
		"issuer",
		1*time.Hour,
	)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), service)
	assert.Equal(suite.T(), "session store is required", err.Error())
}

func (suite *ServiceTestSuite) TestNewService_EmptySecret() {
	service, err := sessions.NewService(
		suite.store,
		"",
		"issuer",
		1*time.Hour,
	)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), service)
	assert.Equal(suite.T(), "jwt secret is required", err.Error())
}

func (suite *ServiceTestSuite) TestNewService_EmptyIssuer() {
	service, err := sessions.NewService(
		suite.store,
		"secret",
		"",
		1*time.Hour,
	)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), service)
	assert.Equal(suite.T(), "jwt issuer is required", err.Error())
}

func (suite *ServiceTestSuite) TestNewService_InvalidTTL() {
	service, err := sessions.NewService(
		suite.store,
		"secret",
		"issuer",
		0*time.Hour,
	)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), service)
	assert.Equal(suite.T(), "session ttl must be greater than 0", err.Error())
}

// Create tests

func (suite *ServiceTestSuite) TestCreate_Success() {
	userID := int64(42)
	suite.store.CreateStub = func(ctx context.Context, session *models.Session) error {
		session.ID = 1
		return nil
	}

	session, token, err := suite.service.Create(suite.ctx, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), session)
	assert.NotEmpty(suite.T(), token)
	assert.Equal(suite.T(), userID, session.UserID)
	assert.Equal(suite.T(), 1, suite.store.CreateCallCount())

	// Verify the session was passed correctly
	ctx, passedSession := suite.store.CreateArgsForCall(0)
	assert.Equal(suite.T(), suite.ctx, ctx)
	assert.Equal(suite.T(), userID, passedSession.UserID)
	assert.NotEmpty(suite.T(), passedSession.TokenID)
	assert.NotZero(suite.T(), passedSession.ExpiresAt)
}

func (suite *ServiceTestSuite) TestCreate_StoreError() {
	userID := int64(42)
	expectedErr := sql.ErrConnDone
	suite.store.CreateStub = func(ctx context.Context, session *models.Session) error {
		return expectedErr
	}

	session, token, err := suite.service.Create(suite.ctx, userID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), session)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), expectedErr, err)
}

func (suite *ServiceTestSuite) TestCreate_TokenGenerated() {
	userID := int64(99)
	var tokenID string
	suite.store.CreateStub = func(ctx context.Context, session *models.Session) error {
		tokenID = session.TokenID
		session.ID = 2
		return nil
	}

	session, token, err := suite.service.Create(suite.ctx, userID)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
	assert.NotEmpty(suite.T(), tokenID)
	// Token should be a valid JWT
	assert.Contains(suite.T(), token, ".")
	assert.Equal(suite.T(), session.TokenID, tokenID)
}

// Validate tests

func (suite *ServiceTestSuite) TestValidate_Success() {
	userID := int64(55)
	now := time.Now().UTC()
	tokenID := "test-token-id"
	expiresAt := now.Add(24 * time.Hour)

	session := &models.Session{
		ID:        1,
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	suite.store.FindByTokenIDStub = func(ctx context.Context, id string) (*models.Session, error) {
		return session, nil
	}

	// Create a valid token
	token, err := suite.service.SignJWT(userID, tokenID, expiresAt)
	require.NoError(suite.T(), err)

	// Validate the token
	validatedSession, err := suite.service.Validate(suite.ctx, token)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), validatedSession)
	assert.Equal(suite.T(), userID, validatedSession.UserID)
	assert.Equal(suite.T(), tokenID, validatedSession.TokenID)
	assert.Equal(suite.T(), 1, suite.store.FindByTokenIDCallCount())
}

func (suite *ServiceTestSuite) TestValidate_InvalidToken() {
	validatedSession, err := suite.service.Validate(suite.ctx, "invalid-token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedSession)
	assert.Equal(suite.T(), sessions.ErrInvalidToken, err)
}

func (suite *ServiceTestSuite) TestValidate_ExpiredToken() {
	userID := int64(66)
	tokenID := "expired-token"
	now := time.Now().UTC()
	expiresAt := now.Add(-1 * time.Hour) // Expired 1 hour ago

	// Create an expired token
	token, err := suite.service.SignJWT(userID, tokenID, expiresAt)
	require.NoError(suite.T(), err)

	// Validate the expired token
	validatedSession, err := suite.service.Validate(suite.ctx, token)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedSession)
	assert.Equal(suite.T(), sessions.ErrExpiredToken, err)
}

func (suite *ServiceTestSuite) TestValidate_SessionNotFound() {
	userID := int64(77)
	tokenID := "missing-token"
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	suite.store.FindByTokenIDStub = func(ctx context.Context, id string) (*models.Session, error) {
		return nil, sql.ErrNoRows
	}

	token, err := suite.service.SignJWT(userID, tokenID, expiresAt)
	require.NoError(suite.T(), err)

	validatedSession, err := suite.service.Validate(suite.ctx, token)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedSession)
	assert.Equal(suite.T(), sessions.ErrSessionNotFound, err)
}

func (suite *ServiceTestSuite) TestValidate_StoreError() {
	userID := int64(88)
	tokenID := "store-error-token"
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	suite.store.FindByTokenIDStub = func(ctx context.Context, id string) (*models.Session, error) {
		return nil, sql.ErrConnDone
	}

	token, err := suite.service.SignJWT(userID, tokenID, expiresAt)
	require.NoError(suite.T(), err)

	validatedSession, err := suite.service.Validate(suite.ctx, token)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedSession)
	assert.Equal(suite.T(), sql.ErrConnDone, err)
}

func (suite *ServiceTestSuite) TestValidate_RevokedSession() {
	userID := int64(99)
	now := time.Now().UTC()
	tokenID := "revoked-token"
	expiresAt := now.Add(24 * time.Hour)

	revokedAt := now.Add(-1 * time.Hour)
	session := &models.Session{
		ID:        2,
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		RevokedAt: &revokedAt,
		CreatedAt: now,
	}

	suite.store.FindByTokenIDStub = func(ctx context.Context, id string) (*models.Session, error) {
		return session, nil
	}

	token, err := suite.service.SignJWT(userID, tokenID, expiresAt)
	require.NoError(suite.T(), err)

	validatedSession, err := suite.service.Validate(suite.ctx, token)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedSession)
	assert.Equal(suite.T(), sessions.ErrRevokedSession, err)
}

func (suite *ServiceTestSuite) TestValidate_SessionExpiredInDB() {
	userID := int64(111)
	now := time.Now().UTC()
	tokenID := "expired-in-db-token"
	expiresAt := now.Add(-1 * time.Hour) // Past expiry

	session := &models.Session{
		ID:        3,
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	suite.store.FindByTokenIDStub = func(ctx context.Context, id string) (*models.Session, error) {
		return session, nil
	}

	// Create a token that appears valid (JWT checks issuer, etc.) but session is expired
	token, err := suite.service.SignJWT(userID, tokenID, expiresAt)
	require.NoError(suite.T(), err)

	validatedSession, err := suite.service.Validate(suite.ctx, token)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedSession)
	assert.Equal(suite.T(), sessions.ErrExpiredToken, err)
}

// Revoke tests

func (suite *ServiceTestSuite) TestRevoke_Success() {
	tokenID := "revoke-me"
	suite.store.RevokeByTokenIDStub = func(ctx context.Context, id string, revokedAt time.Time) error {
		return nil
	}

	err := suite.service.Revoke(suite.ctx, tokenID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, suite.store.RevokeByTokenIDCallCount())

	ctx, passedTokenID, revokedAt := suite.store.RevokeByTokenIDArgsForCall(0)
	assert.Equal(suite.T(), suite.ctx, ctx)
	assert.Equal(suite.T(), tokenID, passedTokenID)
	assert.WithinDuration(suite.T(), time.Now().UTC(), revokedAt, 1*time.Second)
}

func (suite *ServiceTestSuite) TestRevoke_StoreError() {
	tokenID := "error-token"
	expectedErr := sql.ErrConnDone
	suite.store.RevokeByTokenIDStub = func(ctx context.Context, id string, revokedAt time.Time) error {
		return expectedErr
	}

	err := suite.service.Revoke(suite.ctx, tokenID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedErr, err)
}

// Run the test suite
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
