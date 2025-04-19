package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database
type User struct {
	ID          uuid.UUID
	Name        string
	Email       string
	Gender      string
	Batch       string
	PhoneNumber string
	Role        string
	CreatedAt   time.Time
}

// CreateUser inserts a new user into the database
func CreateUser(db *sql.DB, user *User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()

	query := `
        INSERT INTO users (id, name, email, gender, batch, phone_number, role, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := db.ExecContext(context.Background(), query,
		user.ID, user.Name, user.Email, user.Gender,
		user.Batch, user.PhoneNumber, user.Role, user.CreatedAt,
	)

	return err
}

// GetUserByEmail fetches a user from the database by email
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	query := `SELECT id, name, email, gender, batch, phone_number, role, created_at FROM users WHERE email = $1`
	row := db.QueryRowContext(context.Background(), query, email)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Gender, &user.Batch, &user.PhoneNumber, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // not found
		}
		return nil, err
	}

	return &user, nil
}

//Handles user DB queries: CreateUser, FindByEmail, etc.
