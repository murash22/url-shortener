package sqlite

import (
	"database/sql"
	"errors"
	"github.com/mattn/go-sqlite3"
	"url-shortener/internal/models"
	"url-shortener/internal/storage"
)

func (s *Storage) SaveURL(urlShortener models.UrlShortener) (int64, error) {
	stmt, err := s.db.Prepare("INSERT INTO url (alias, url) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(urlShortener.Alias, urlShortener.Url)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, storage.ErrUrlExists
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, err
}

func (s *Storage) GetURL(alias string) (string, error) {
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", err
	}
	var queryUrl string
	err = stmt.QueryRow(alias).Scan(&queryUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrUrlNotFound
	}
	if err != nil {
		return "", err
	}
	return queryUrl, nil
}

func (s *Storage) DeleteURL(alias string) error {

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrUrlNotFound
	}
	if err != nil {
		return err
	}
	return nil
}
