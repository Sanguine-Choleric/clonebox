package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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
	stmt := `INSERT INTO files (file_name, file_uuid, file_size, checksum, storage_path, upload_date) VALUES (
				$1, $2, $3, $4, $5, now() AT TIME ZONE 'UTC')`

	_, err := m.DB.Exec(stmt, fileName, uuid, fileSize, checksum, storagePath)
	if err != nil {
		//if mySqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
		//	if mySqlErr.Number == 1062 {
		//		return ErrDuplicateUUID
		//	}
		//}

		if postgresErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if postgresErr.Code == "23505" { // https://www.postgresql.org/docs/8.4/errcodes-appendix.html
				return ErrDuplicateUUID
			}
		}
		return err
	}

	return nil
}

func (m *FileModel) GetByUUID(uuid string) (*File, error) {
	stmt := `SELECT * FROM files WHERE file_uuid = $1`

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
	stmt := `SELECT * FROM files WHERE checksum = $1`

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
