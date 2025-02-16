package models

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
)

type LinkMappingModelInterface interface {
	//
	Insert(original, short string) error
	GetOriginal(short string) (string, error)
	GetShort(original string) (string, error)
	Exists(original string) (bool, error)
	Latest() ([]*LinkMapping, error)
}

type LinkMappingModel struct {
	DB *sql.DB
}

func (m *LinkMappingModel) Latest() ([]*LinkMapping, error) {
	stmt := `SELECT original_link, short_link FROM link_mapping
				ORDER BY ID DESC LIMIT 5`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []*LinkMapping{}

	for rows.Next() {
		l := &LinkMapping{}

		err = rows.Scan(&l.OriginalLink, &l.ShortLink)
		if err != nil {
			return nil, err
		}

		links = append(links, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (m *LinkMappingModel) Exists(original string) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT TRUE FROM link_mapping WHERE original_link = ?)`

	err := m.DB.QueryRow(stmt, original).Scan(&exists)
	return exists, err
}

func (m *LinkMappingModel) GetOriginal(hash string) (string, error) {
	var originalLink string
	stmt := `SELECT original_link FROM link_mapping WHERE short_link = ?`
	err := m.DB.QueryRow(stmt, hash).Scan(&originalLink)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoRecord
		}
		return "", err
	}

	return originalLink, nil
}

func (m *LinkMappingModel) GetShort(link string) (string, error) {
	var originalLink string
	stmt := `SELECT short_link FROM link_mapping WHERE original_link = ?`
	err := m.DB.QueryRow(stmt, link).Scan(&originalLink)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoRecord
		}
		return "", err
	}

	return originalLink, nil
}

type LinkMapping struct {
	// Add validation for links later
	ID           int
	OriginalLink string
	ShortLink    string
}

func (m *LinkMappingModel) Insert(original, short string) error {
	stmt := `INSERT INTO link_mapping (original_link, short_link)
				VALUES (?, ?)`
	_, err := m.DB.Exec(stmt, original, short)
	if err != nil {
		var mySqlErr *mysql.MySQLError
		if errors.As(err, &mySqlErr) {
			if mySqlErr.Number == 1062 {
				return ErrDuplicateLink
			}
		}
		return err
	}

	return nil
}
