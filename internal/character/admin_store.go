package character

import (
	"database/sql"
	"errors"
	"fmt"
)

type CharacterWithOwner struct {
	Character
	OwnerUsername string
}

func (s *Store) ListAllWithOwners() ([]CharacterWithOwner, error) {
	rows, err := s.db.Query(
		`SELECT
			c.id,
			c.user_id,
			c.name,
			c.level,
			c.class,
			c.sheet_version,
			c.data,
			c.created_at,
			c.updated_at,
			u.username
		 FROM characters c
		 JOIN users u ON u.id = c.user_id
		 ORDER BY u.username COLLATE NOCASE, c.name COLLATE NOCASE, c.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all characters: %w", err)
	}
	defer rows.Close()

	var characters []CharacterWithOwner

	for rows.Next() {
		var c CharacterWithOwner
		if err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.Level,
			&c.Class,
			&c.SheetVersion,
			&c.Data,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.OwnerUsername,
		); err != nil {
			return nil, fmt.Errorf("scan character with owner: %w", err)
		}

		characters = append(characters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all characters: %w", err)
	}

	return characters, nil
}

func (s *Store) GetByIDWithOwner(characterID int64) (CharacterWithOwner, error) {
	var c CharacterWithOwner

	err := s.db.QueryRow(
		`SELECT
			c.id,
			c.user_id,
			c.name,
			c.level,
			c.class,
			c.sheet_version,
			c.data,
			c.created_at,
			c.updated_at,
			u.username
		 FROM characters c
		 JOIN users u ON u.id = c.user_id
		 WHERE c.id = ?`,
		characterID,
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
		&c.OwnerUsername,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return CharacterWithOwner{}, ErrNotFound
	}
	if err != nil {
		return CharacterWithOwner{}, fmt.Errorf("get character with owner: %w", err)
	}

	return c, nil
}

func (s *Store) DeleteAny(characterID int64) error {
	result, err := s.db.Exec(
		`DELETE FROM characters WHERE id = ?`,
		characterID,
	)
	if err != nil {
		return fmt.Errorf("admin delete character: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get admin deleted character count: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
