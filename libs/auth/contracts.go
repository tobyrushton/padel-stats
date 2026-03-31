package auth

import "time"

// SignupInput is the request payload for user registration.
type SignupInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// SigninInput is the request payload for user authentication.
type SigninInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SearchPlayersResult is the response payload for player search.
type SearchPlayersResult struct {
	Players []*User `json:"players"`
}

// User is an API-safe user contract for OpenAPI generation.
type User struct {
	ID        int64     `json:"id" format:"int64"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt" format:"date-time"`
	UpdatedAt time.Time `json:"updatedAt" format:"date-time"`
}

// AuthResult is the response payload for signup/signin operations.
type AuthResult struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// UserRecord is a persistence-layer contract independent from DB model structs.
type UserRecord struct {
	ID             int64
	FirstName      string
	LastName       string
	Username       string
	HashedPassword string
	IsAdmin        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func userFromRecord(record *UserRecord) *User {
	if record == nil {
		return nil
	}

	return &User{
		ID:        record.ID,
		FirstName: record.FirstName,
		LastName:  record.LastName,
		Username:  record.Username,
		IsAdmin:   record.IsAdmin,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}
