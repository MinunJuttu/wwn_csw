package session

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("session not found")

const Lifetime = 30 * 24 * time.Hour

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// Create создаёт новую сессию.
//
// В cookie браузера отправляется случайный token.
// В базе хранится только его SHA-256 хеш.
func (s *Store) Create(userID int64) (string, time.Time, error) {
	token, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}

	tokenHash := hashToken(token)

	expiresAt := time.Now().
		UTC().
		Add(Lifetime)

	_, err = s.db.Exec(
		`
		INSERT INTO sessions (
			user_id,
			token_hash,
			expires_at
		)
		VALUES (?, ?, ?)
		`,
		userID,
		tokenHash,
		expiresAt.Format(time.RFC3339),
	)

	if err != nil {
		return "", time.Time{}, fmt.Errorf(
			"insert session: %w",
			err,
		)
	}

	return token, expiresAt, nil
}

// GetUserID ищет действующую сессию
// и возвращает ID её пользователя.
func (s *Store) GetUserID(token string) (int64, error) {
	tokenHash := hashToken(token)

	var userID int64

	err := s.db.QueryRow(
		`
		SELECT user_id
		FROM sessions
		WHERE token_hash = ?
		  AND expires_at > ?
		`,
		tokenHash,
		time.Now().UTC().Format(time.RFC3339),
	).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}

	if err != nil {
		return 0, err
	}

	return userID, nil
}

// Delete уничтожает конкретную сессию.
func (s *Store) Delete(token string) error {
	tokenHash := hashToken(token)

	_, err := s.db.Exec(
		`
		DELETE FROM sessions
		WHERE token_hash = ?
		`,
		tokenHash,
	)

	return err
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf(
			"generate random token: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
