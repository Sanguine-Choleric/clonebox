package models

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"time"
)

type FilesModelInterface interface {
	Insert(fileName string, uuid string, fileSize int, checksum string, storagePath string) error
	GetByUUID(uuid string) (*File, error)
	GetByChecksum(checksum string) (*File, error)
}

type FileModel struct {
	DB *sql.DB
}

type File struct {
	ID          int
	FileName    string
	FileUUID    string
	FileSize    int
	Checksum    string
	UploadTime  time.Time
	StoragePath string
}

func (m *FileModel) Insert(fileName string, uuid string, fileSize int, checksum string, storagePath string) error {
	stmt := `INSERT INTO files (original_file_name, file_name, file_size, checksum, storage_path, upload_date) VALUES (
				?, ?, ?, ?, ?, UTC_TIMESTAMP())`

	_, err := m.DB.Exec(stmt, fileName, uuid, fileSize, checksum, storagePath)
	if err != nil {
		var mySqlErr *mysql.MySQLError
		if errors.As(err, &mySqlErr) {
			if mySqlErr.Number == 1062 {
				return ErrDuplicateUUID
			}
		}
		return err
	}

	return nil
}

func (m *FileModel) GetByUUID(uuid string) (*File, error) {
	stmt := `SELECT * FROM files WHERE file_name = ?`

	s := &File{}

	err := m.DB.QueryRow(stmt, uuid).Scan(&s.ID, &s.FileName, &s.FileUUID, &s.FileSize, &s.Checksum, &s.StoragePath, &s.UploadTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *FileModel) GetByChecksum(checksum string) (*File, error) {
	stmt := `SELECT * FROM files WHERE checksum = ?`

	s := &File{}

	err := m.DB.QueryRow(stmt, checksum).Scan(&s.ID, &s.FileName, &s.FileUUID, &s.FileSize, &s.Checksum, &s.StoragePath, &s.UploadTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}
