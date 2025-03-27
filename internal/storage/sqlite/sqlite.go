package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"runtime"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	_, op, line, _ := runtime.Caller(0)

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s:%d - %s", op, line, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS url (
		    id INTEGER PRIMARY KEY,
		    alias TEXT NOT NULL UNIQUE,
		    url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s:%d - %s", op, line, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s:%d - %s", op, line, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	_, op, line, _ := runtime.Caller(0)

	stmt, err := s.db.Prepare("INSERT INTO url (alias, url) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s:%d - %s", op, line, err)
	}

	res, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, storage.ErrUrlExists
		}
		return 0, fmt.Errorf("%s:%d - %s", op, line, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s:%d - %s", op, line, err)
	}
	return id, err
}

func (s *Storage) GetURL(alias string) (string, error) {
	_, op, line, _ := runtime.Caller(0)
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s:%d - %s", op, line, err)
	}
	var queryUrl string
	err = stmt.QueryRow(alias).Scan(&queryUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrUrlNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s:%d - %s", op, line, err)
	}
	return queryUrl, nil
}

func (s *Storage) DeleteURL(alias string) error {
	_, op, line, _ := runtime.Caller(0)

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s:%d - %s", op, line, err)
	}
	_, err = stmt.Exec(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrUrlNotFound
	}
	if err != nil {
		return fmt.Errorf("%s:%d - %s", op, line, err)
	}
	return nil
}
