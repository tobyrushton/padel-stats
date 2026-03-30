package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"github.com/tobyrushton/padel-stats/libs/auth"
	"github.com/tobyrushton/padel-stats/libs/fakes"
)

type AuthServiceTestSuite struct {
	suite.Suite
	userRepo   *fakes.FakeUserRepository
	sessionSvc *fakes.FakeSessionService
	service    *auth.Service
	ctx        context.Context
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.userRepo = new(fakes.FakeUserRepository)
	suite.sessionSvc = new(fakes.FakeSessionService)
	suite.ctx = context.Background()

	var err error
	suite.service, err = auth.NewService(suite.userRepo, suite.sessionSvc)
	require.NoError(suite.T(), err)
}

func (suite *AuthServiceTestSuite) TestSignup_Success() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*auth.UserRecord, error) {
		return nil, auth.ErrUserNotFound
	}
	suite.userRepo.CreateUserStub = func(ctx context.Context, user *auth.UserRecord) error {
		user.ID = 1
		return nil
	}
	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (string, error) {
		return "jwt-token-123", nil
	}

	result, err := suite.service.Signup(suite.ctx, input)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotNil(suite.T(), result.User)
	assert.Equal(suite.T(), "johndoe", result.User.Username)
	assert.Equal(suite.T(), "jwt-token-123", result.Token)
	assert.Equal(suite.T(), 1, suite.userRepo.CreateUserCallCount())
	assert.Equal(suite.T(), 1, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignup_UserExists() {
	input := &auth.SignupInput{FirstName: "John", LastName: "Doe", Username: "johndoe", Password: "securepassword123"}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*auth.UserRecord, error) {
		return &auth.UserRecord{ID: 1, Username: "johndoe"}, nil
	}

	result, err := suite.service.Signup(suite.ctx, input)

	assert.ErrorIs(suite.T(), err, auth.ErrUserExists)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.userRepo.CreateUserCallCount())
}

func (suite *AuthServiceTestSuite) TestSignup_ValidationError() {
	input := &auth.SignupInput{FirstName: "", LastName: "Doe", Username: "johndoe", Password: "securepassword123"}

	result, err := suite.service.Signup(suite.ctx, input)

	assert.ErrorIs(suite.T(), err, auth.ErrInvalidFirstName)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.userRepo.FindUserByUsernameCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_Success() {
	hash, err := bcrypt.GenerateFromPassword([]byte("securepassword123"), 12)
	require.NoError(suite.T(), err)

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*auth.UserRecord, error) {
		return &auth.UserRecord{ID: 1, Username: username, HashedPassword: string(hash)}, nil
	}
	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (string, error) {
		return "jwt-token-456", nil
	}

	result, err := suite.service.Signin(suite.ctx, &auth.SigninInput{Username: "johndoe", Password: "securepassword123"})

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "johndoe", result.User.Username)
	assert.Equal(suite.T(), "jwt-token-456", result.Token)
}

func (suite *AuthServiceTestSuite) TestSignin_UserNotFound() {
	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*auth.UserRecord, error) {
		return nil, auth.ErrUserNotFound
	}

	result, err := suite.service.Signin(suite.ctx, &auth.SigninInput{Username: "missing", Password: "password123"})

	assert.ErrorIs(suite.T(), err, auth.ErrInvalidPassword)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_WrongPassword() {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), 12)
	require.NoError(suite.T(), err)

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*auth.UserRecord, error) {
		return &auth.UserRecord{ID: 1, Username: username, HashedPassword: string(hash)}, nil
	}

	result, err := suite.service.Signin(suite.ctx, &auth.SigninInput{Username: "johndoe", Password: "wrong-password"})

	assert.ErrorIs(suite.T(), err, auth.ErrInvalidPassword)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_SessionError() {
	hash, err := bcrypt.GenerateFromPassword([]byte("securepassword123"), 12)
	require.NoError(suite.T(), err)

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*auth.UserRecord, error) {
		return &auth.UserRecord{ID: 1, Username: username, HashedPassword: string(hash)}, nil
	}
	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (string, error) {
		return "", errors.New("session create failed")
	}

	result, err := suite.service.Signin(suite.ctx, &auth.SigninInput{Username: "johndoe", Password: "securepassword123"})

	assert.EqualError(suite.T(), err, "session create failed")
	assert.Nil(suite.T(), result)
}

func (suite *AuthServiceTestSuite) TestSearchPlayers_Success() {
	suite.userRepo.SearchUsersByQueryStub = func(ctx context.Context, query string) ([]*auth.UserRecord, error) {
		return []*auth.UserRecord{{
			ID:        12,
			FirstName: "Jane",
			LastName:  "Doe",
			Username:  "jane",
		}}, nil
	}

	result, err := suite.service.SearchPlayers(suite.ctx, "ja")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Players, 1)
	assert.Equal(suite.T(), int64(12), result.Players[0].ID)
	assert.Equal(suite.T(), "jane", result.Players[0].Username)
}

func (suite *AuthServiceTestSuite) TestSearchPlayers_EmptyQueryReturnsDefaultList() {
	suite.userRepo.SearchUsersByQueryStub = func(ctx context.Context, query string) ([]*auth.UserRecord, error) {
		assert.Equal(suite.T(), "", query)
		return []*auth.UserRecord{{
			ID:        2,
			FirstName: "Default",
			LastName:  "Player",
			Username:  "default-player",
		}}, nil
	}

	result, err := suite.service.SearchPlayers(suite.ctx, "   ")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Players, 1)
	assert.Equal(suite.T(), "default-player", result.Players[0].Username)
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
