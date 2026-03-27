package auth

import (
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidUsername  = errors.New("invalid username format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidFirstName = errors.New("invalid first name")
	ErrInvalidLastName  = errors.New("invalid last name")
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
	bcryptCost        = 12
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

func (in *SignupInput) Validate() error {
	if in.FirstName == "" {
		return ErrInvalidFirstName
	}
	if in.LastName == "" {
		return ErrInvalidLastName
	}
	if !usernameRegex.MatchString(in.Username) {
		return fmt.Errorf("%w: must be 3-32 characters, alphanumeric, dash, or underscore", ErrInvalidUsername)
	}
	if len(in.Password) < minPasswordLength || len(in.Password) > maxPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

func (in *SigninInput) Validate() error {
	if in.Username == "" {
		return ErrInvalidUsername
	}
	if in.Password == "" {
		return ErrPasswordTooShort
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func verifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
