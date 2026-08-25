package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// schema содержит содержимое файла schema.sql.
//
// Благодаря go:embed SQL-схема будет включена прямо
// в скомпилированный бинарник приложения.
//
//go:embed schema.sql
var schema string

// Open открывает базу данных SQLite.
// Если файла базы ещё нет, SQLite создаст его автоматически.
func Open(path string) (*sql.DB, error) {
	// Убеждаемся, что директория для базы существует.
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Настройки SQLite:
	//
	// foreign_keys=on
	//     включает поддержку FOREIGN KEY.
	//
	// journal_mode=WAL
	//     позволяет чтению и записи лучше сосуществовать.
	//
	// busy_timeout=5000
	//     если база временно занята, SQLite подождёт
	//     до 5 секунд вместо мгновенной ошибки.
	dsn := fmt.Sprintf(
		"file:%s?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000",
		filepath.ToSlash(path),
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Для нашего маленького приложения пяти соединений
	// более чем достаточно.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	// sql.Open сам по себе ещё не гарантирует,
	// что соединение с базой действительно работает.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Создаём отсутствующие таблицы.
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize database schema: %w", err)
	}

	return db, nil
}
