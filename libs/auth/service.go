package auth

import (
	"context"
	"errors"
	"strings"
)

//go:generate go tool counterfeiter -generate

//counterfeiter:generate -o ../fakes/user-repository.go . UserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, user *UserRecord) error
	FindUserByID(ctx context.Context, userID int64) (*UserRecord, error)
	FindUserByUsername(ctx context.Context, username string) (*UserRecord, error)
	ApproveUserByID(ctx context.Context, userID int64) error
	SearchUsersByQuery(ctx context.Context, query string) ([]*UserRecord, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

//counterfeiter:generate -o ../fakes/auth-session-service.go . SessionService
type SessionService interface {
	Create(ctx context.Context, userID int64) (string, error)
}

type Service struct {
	userRepo   UserRepository
	sessionSvc SessionService
}

func NewService(userRepo UserRepository, sessionSvc SessionService) (*Service, error) {
	if userRepo == nil {
		return nil, errors.New("user repository is required")
	}
	if sessionSvc == nil {
		return nil, errors.New("session service is required")
	}

	return &Service{
		userRepo:   userRepo,
		sessionSvc: sessionSvc,
	}, nil
}

func (s *Service) Signup(ctx context.Context, input *SignupInput) (*AuthResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Check if user exists
	_, err := s.userRepo.FindUserByUsername(ctx, input.Username)
	if err == nil {
		// User exists
		return nil, ErrUserExists
	}
	if err != ErrUserNotFound {
		// Some other error occurred
		return nil, err
	}

	// Hash password
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &UserRecord{
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Username:          input.Username,
		HashedPassword:    hashedPassword,
		IsAcceptedByAdmin: false,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &AuthResult{User: userFromRecord(user), Token: ""}, nil
}

func (s *Service) Signin(ctx context.Context, input *SigninInput) (*AuthResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Find user
	user, err := s.userRepo.FindUserByUsername(ctx, input.Username)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidPassword // Don't leak username existence
		}
		return nil, err
	}

	// Verify password
	if !verifyPassword(user.HashedPassword, input.Password) {
		return nil, ErrInvalidPassword
	}
	if !user.IsAcceptedByAdmin {
		return nil, ErrUserPendingApproval
	}

	// Create session
	token, err := s.sessionSvc.Create(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: userFromRecord(user), Token: token}, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return userFromRecord(user), nil
}

func (s *Service) ApproveUser(ctx context.Context, adminUserID, userID int64) (*User, error) {
	isAdmin, err := s.userRepo.IsAdmin(ctx, adminUserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrAdminAccessRequired
	}

	if _, err := s.userRepo.FindUserByID(ctx, userID); err != nil {
		return nil, err
	}

	if err := s.userRepo.ApproveUserByID(ctx, userID); err != nil {
		return nil, err
	}

	updatedUser, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return userFromRecord(updatedUser), nil
}

func (s *Service) SearchPlayers(ctx context.Context, query string) (*SearchPlayersResult, error) {
	trimmedQuery := strings.TrimSpace(query)
	users, err := s.userRepo.SearchUsersByQuery(ctx, trimmedQuery)
	if err != nil {
		return nil, err
	}

	players := make([]*User, 0, len(users))
	for _, user := range users {
		players = append(players, userFromRecord(user))
	}

	return &SearchPlayersResult{Players: players}, nil
}
