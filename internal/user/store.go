package user

import (
	"database/sql"
	"errors"
)

var (
	ErrUsernameTaken = errors.New("username already taken")
	ErrNotFound      = errors.New("user not found")
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// Create создаёт нового пользователя.
func (s *Store) Create(username string, passwordHash string) (int64, error) {
	result, err := s.db.Exec(
		`
		INSERT INTO users (
			username,
			password_hash
		)
		VALUES (?, ?)
		ON CONFLICT(username) DO NOTHING
		`,
		username,
		passwordHash,
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, ErrUsernameTaken
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetByUsername возвращает пользователя по логину.
func (s *Store) GetByUsername(username string) (User, error) {
	var u User

	err := s.db.QueryRow(
		`
		SELECT
			id,
			username,
			password_hash
		FROM users
		WHERE username = ?
		`,
		username,
	).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}

	if err != nil {
		return User{}, err
	}

	return u, nil
}

// GetByID возвращает пользователя по ID.
func (s *Store) GetByID(id int64) (User, error) {
	var u User

	err := s.db.QueryRow(
		`
		SELECT
			id,
			username,
			password_hash
		FROM users
		WHERE id = ?
		`,
		id,
	).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}

	if err != nil {
		return User{}, err
	}

	return u, nil
}
