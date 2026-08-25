package character

import (
	"database/sql"
	"fmt"
)

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
