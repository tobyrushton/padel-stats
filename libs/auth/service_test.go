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
	"github.com/tobyrushton/padel-stats/libs/db/models"
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

// NewService tests

func (suite *AuthServiceTestSuite) TestNewService_Success() {
	svc, err := auth.NewService(suite.userRepo, suite.sessionSvc)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), svc)
}

func (suite *AuthServiceTestSuite) TestNewService_NilUserRepository() {
	svc, err := auth.NewService(nil, suite.sessionSvc)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), svc)
	assert.Equal(suite.T(), "user repository is required", err.Error())
}

func (suite *AuthServiceTestSuite) TestNewService_NilSessionService() {
	svc, err := auth.NewService(suite.userRepo, nil)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), svc)
	assert.Equal(suite.T(), "session service is required", err.Error())
}

// Signup tests

func (suite *AuthServiceTestSuite) TestSignup_Success() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	// Mock user not found
	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return nil, auth.ErrUserNotFound
	}

	// Mock CreateUser to set the ID
	suite.userRepo.CreateUserStub = func(ctx context.Context, user *models.User) error {
		user.ID = 1
		return nil
	}

	// Mock session creation
	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (*models.Session, string, error) {
		return &models.Session{ID: 1, UserID: userID}, "jwt-token-123", nil
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.NotEmpty(suite.T(), token)
	assert.Equal(suite.T(), "John", user.FirstName)
	assert.Equal(suite.T(), "Doe", user.LastName)
	assert.Equal(suite.T(), "johndoe", user.Username)
	assert.Equal(suite.T(), "jwt-token-123", token)

	// Verify repository was called
	assert.Equal(suite.T(), 1, suite.userRepo.FindUserByUsernameCallCount())
	assert.Equal(suite.T(), 1, suite.userRepo.CreateUserCallCount())
	assert.Equal(suite.T(), 1, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignup_UserExists() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	existingUser := &models.User{
		ID:             1,
		FirstName:      "Jane",
		LastName:       "Doe",
		Username:       "johndoe",
		HashedPassword: "hashed",
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return existingUser, nil
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrUserExists, err)

	// CreateUser should not have been called
	assert.Equal(suite.T(), 0, suite.userRepo.CreateUserCallCount())
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignup_InvalidInput_EmptyFirstName() {
	input := &auth.SignupInput{
		FirstName: "",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrInvalidFirstName, err)
	assert.Equal(suite.T(), 0, suite.userRepo.FindUserByUsernameCallCount())
}

func (suite *AuthServiceTestSuite) TestSignup_InvalidInput_EmptyLastName() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrInvalidLastName, err)
}

func (suite *AuthServiceTestSuite) TestSignup_InvalidInput_InvalidUsername() {
	testCases := []string{
		"ab",                               // too short
		"a",                                // too short
		"toolong1234567890123456789012345", // too long
		"user@name",                        // invalid chars
		"user name",                        // space not allowed
	}

	for _, username := range testCases {
		input := &auth.SignupInput{
			FirstName: "John",
			LastName:  "Doe",
			Username:  username,
			Password:  "securepassword123",
		}

		user, token, err := suite.service.Signup(suite.ctx, input)

		assert.Error(suite.T(), err, "username should be invalid: %s", username)
		assert.Nil(suite.T(), user)
		assert.Empty(suite.T(), token)
	}
}

func (suite *AuthServiceTestSuite) TestSignup_InvalidInput_PasswordTooShort() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "short",
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrPasswordTooShort, err)
}

func (suite *AuthServiceTestSuite) TestSignup_RepositoryError() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	expectedErr := errors.New("database error")

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return nil, auth.ErrUserNotFound
	}

	suite.userRepo.CreateUserStub = func(ctx context.Context, user *models.User) error {
		return expectedErr
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignup_SessionCreationError() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	expectedErr := errors.New("session error")

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return nil, auth.ErrUserNotFound
	}

	suite.userRepo.CreateUserStub = func(ctx context.Context, user *models.User) error {
		user.ID = 1
		return nil
	}

	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (*models.Session, string, error) {
		return nil, "", expectedErr
	}

	user, token, err := suite.service.Signup(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), expectedErr, err)
}

func (suite *AuthServiceTestSuite) TestSignup_PasswordHashingWorks() {
	input := &auth.SignupInput{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "securepassword123",
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return nil, auth.ErrUserNotFound
	}

	var capturedUser *models.User
	suite.userRepo.CreateUserStub = func(ctx context.Context, user *models.User) error {
		capturedUser = user
		user.ID = 1
		return nil
	}

	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (*models.Session, string, error) {
		return &models.Session{ID: 1}, "token", nil
	}

	_, _, err := suite.service.Signup(suite.ctx, input)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), capturedUser)
	// Password should be hashed, not plaintext
	assert.NotEqual(suite.T(), "securepassword123", capturedUser.HashedPassword)
	// Hash should be reasonably long (bcrypt hashes are ~60 chars)
	assert.Greater(suite.T(), len(capturedUser.HashedPassword), 50)
}

// Signin tests

func (suite *AuthServiceTestSuite) TestSignin_Success() {
	input := &auth.SigninInput{
		Username: "johndoe",
		Password: "securepassword123",
	}

	// Create a real hash for the password
	hashedPassword, err := hashPassword("securepassword123")
	require.NoError(suite.T(), err)

	user := &models.User{
		ID:             1,
		FirstName:      "John",
		LastName:       "Doe",
		Username:       "johndoe",
		HashedPassword: hashedPassword,
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return user, nil
	}

	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (*models.Session, string, error) {
		return &models.Session{ID: 1}, "jwt-token-456", nil
	}

	returnedUser, token, err := suite.service.Signin(suite.ctx, input)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), returnedUser)
	assert.NotEmpty(suite.T(), token)
	assert.Equal(suite.T(), "johndoe", returnedUser.Username)
	assert.Equal(suite.T(), "jwt-token-456", token)
	assert.Equal(suite.T(), 1, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_UserNotFound() {
	input := &auth.SigninInput{
		Username: "nonexistent",
		Password: "anypassword",
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return nil, auth.ErrUserNotFound
	}

	user, token, err := suite.service.Signin(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	// Should return InvalidPassword to avoid username enumeration
	assert.Equal(suite.T(), auth.ErrInvalidPassword, err)
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_InvalidPassword() {
	input := &auth.SigninInput{
		Username: "johndoe",
		Password: "wrongpassword",
	}

	hashedPassword, err := hashPassword("correctpassword")
	require.NoError(suite.T(), err)

	user := &models.User{
		ID:             1,
		Username:       "johndoe",
		HashedPassword: hashedPassword,
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return user, nil
	}

	returnedUser, token, err := suite.service.Signin(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), returnedUser)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrInvalidPassword, err)
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_InvalidInput_EmptyUsername() {
	input := &auth.SigninInput{
		Username: "",
		Password: "password123",
	}

	user, token, err := suite.service.Signin(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrInvalidUsername, err)
	assert.Equal(suite.T(), 0, suite.userRepo.FindUserByUsernameCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_InvalidInput_EmptyPassword() {
	input := &auth.SigninInput{
		Username: "johndoe",
		Password: "",
	}

	user, token, err := suite.service.Signin(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), auth.ErrPasswordTooShort, err)
	assert.Equal(suite.T(), 0, suite.userRepo.FindUserByUsernameCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_RepositoryError() {
	input := &auth.SigninInput{
		Username: "johndoe",
		Password: "password123",
	}

	expectedErr := errors.New("database error")

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return nil, expectedErr
	}

	user, token, err := suite.service.Signin(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), expectedErr, err)
	assert.Equal(suite.T(), 0, suite.sessionSvc.CreateCallCount())
}

func (suite *AuthServiceTestSuite) TestSignin_SessionCreationError() {
	input := &auth.SigninInput{
		Username: "johndoe",
		Password: "securepassword123",
	}

	hashedPassword, err := hashPassword("securepassword123")
	require.NoError(suite.T(), err)

	user := &models.User{
		ID:             1,
		Username:       "johndoe",
		HashedPassword: hashedPassword,
	}

	suite.userRepo.FindUserByUsernameStub = func(ctx context.Context, username string) (*models.User, error) {
		return user, nil
	}

	expectedErr := errors.New("session creation failed")
	suite.sessionSvc.CreateStub = func(ctx context.Context, userID int64) (*models.Session, string, error) {
		return nil, "", expectedErr
	}

	returnedUser, token, err := suite.service.Signin(suite.ctx, input)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), returnedUser)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), expectedErr, err)
}

// Test suite runner
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// Helper to hash password for test setup
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}
