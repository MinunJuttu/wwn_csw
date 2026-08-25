package character

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("character not found")

type Character struct {
	ID           int64
	UserID       int64
	Name         string
	Level        int
	Class        string
	SheetVersion int
	Data         string
	CreatedAt    string
	UpdatedAt    string
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

// Create создаёт нового персонажа,
// принадлежащего конкретному пользователю.
func (s *Store) Create(
	userID int64,
	name string,
) (int64, error) {
	result, err := s.db.Exec(
		`
		INSERT INTO characters (
			user_id,
			name,
			level,
			class,
			sheet_version,
			data
		)
		VALUES (?, ?, 1, '', 1, '{}')
		`,
		userID,
		name,
	)

	if err != nil {
		return 0, fmt.Errorf(
			"insert character: %w",
			err,
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(
			"get character id: %w",
			err,
		)
	}

	return id, nil
}

// ListByUserID возвращает только персонажей
// конкретного пользователя.
func (s *Store) ListByUserID(
	userID int64,
) ([]Character, error) {
	rows, err := s.db.Query(
		`
		SELECT
			id,
			user_id,
			name,
			level,
			class,
			sheet_version,
			data,
			created_at,
			updated_at
		FROM characters
		WHERE user_id = ?
		ORDER BY id DESC
		`,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"query characters: %w",
			err,
		)
	}
	defer rows.Close()

	var characters []Character

	for rows.Next() {
		var c Character

		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.Level,
			&c.Class,
			&c.SheetVersion,
			&c.Data,
			&c.CreatedAt,
			&c.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"scan character: %w",
				err,
			)
		}

		characters = append(
			characters,
			c,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate characters: %w",
			err,
		)
	}

	return characters, nil
}

// GetByIDForUser возвращает персонажа,
// только если он принадлежит указанному пользователю.
func (s *Store) GetByIDForUser(
	characterID int64,
	userID int64,
) (Character, error) {
	var c Character

	err := s.db.QueryRow(
		`
		SELECT
			id,
			user_id,
			name,
			level,
			class,
			sheet_version,
			data,
			created_at,
			updated_at
		FROM characters
		WHERE id = ?
		  AND user_id = ?
		`,
		characterID,
		userID,
	).Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.Level,
		&c.Class,
		&c.SheetVersion,
		&c.Data,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Character{}, ErrNotFound
	}

	if err != nil {
		return Character{}, fmt.Errorf(
			"get character: %w",
			err,
		)
	}

	return c, nil
}

// Update сохраняет основные поля персонажа и JSON-лист.
//
// characterID и userID используются вместе,
// чтобы нельзя было изменить чужого персонажа.
func (s *Store) Update(
	characterID int64,
	userID int64,
	name string,
	level int,
	class string,
	data string,
) error {
	result, err := s.db.Exec(
		`
		UPDATE characters
		SET
			name = ?,
			level = ?,
			class = ?,
			data = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND user_id = ?
		`,
		name,
		level,
		class,
		data,
		characterID,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"update character: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) Delete(
	characterID int64,
	userID int64,
) error {
	result, err := s.db.Exec(
		`
		DELETE FROM characters
		WHERE id = ? AND user_id = ?
		`,
		characterID,
		userID,
	)

	if err != nil {
		return fmt.Errorf(
			"delete character: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get deleted character count: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
