package security

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type AuthService struct {
	db *DB
}

func NewAuthService(db *DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !ValidateEmail(email) {
		return "", fmt.Errorf("invalid email format")
	}

	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	id := uuid.New().String()
	_, err = s.db.Conn.Exec(
		`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		id, email, hash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return "", fmt.Errorf("user with this email already exists")
		}
		return "", err
	}

	return id, nil
}

func (s *AuthService) Login(email, password string) (string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var id, hash string
	err := s.db.Conn.QueryRow(
		`SELECT id, password_hash FROM users WHERE LOWER(email) = ?`,
		email,
	).Scan(&id, &hash)

	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("invalid email or password")
	} else if err != nil {
		return "", "", err
	}

	match, err := CheckPassword(password, hash)
	if err != nil || !match {
		return "", "", fmt.Errorf("invalid email or password")
	}

	token, err := GenerateToken(id, email)
	if err != nil {
		return "", "", err
	}

	return id, token, nil
}
