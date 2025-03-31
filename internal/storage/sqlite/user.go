package sqlite

import (
	"database/sql"
	"errors"
	"github.com/mattn/go-sqlite3"
	"url-shortener/internal/models"
	"url-shortener/internal/storage"
)

func (s *Storage) SaveUser(user models.User) (int64, error) {
	stmt, err := s.db.Prepare("INSERT INTO users(email, password) VALUES (?, ?)")

	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(user.Email, user.Password)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, storage.ErrUserExists
		}
		return 0, err
	}

	id, err := res.LastInsertId()
	return id, err
}

func (s *Storage) GetUserByEmail(email string) (*models.User, error) {
	stmt, err := s.db.Prepare("SELECT * FROM users WHERE email = ?")
	if err != nil {
		return nil, err
	}
	var user models.User
	err = stmt.QueryRow(email).Scan(&user.Id, &user.Email, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
