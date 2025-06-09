package mocks

import "snippetbox/internal/models"

var mockFile = &models.File{
	ID:          1,
	FileName:    "test_file.pdf",
	FileUUID:    "123456",
	FileSize:    10,
	Checksum:    "abcdef",
	StoragePath: "/clonebox/upload",
}

type FileModel struct{}

func (f *FileModel) Insert(fileName string, uuid string, fileSize int, checksum string, storagePath string) error {
	//TODO implement me
	panic("implement me")
}

func (f *FileModel) GetByUUID(uuid string) (*models.File, error) {
	switch uuid {
	case "123456":
		return mockFile, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (f *FileModel) GetByChecksum(checksum string) (*models.File, error) {
	switch checksum {
	case "abcdef":
		return mockFile, nil
	default:
		return nil, models.ErrNoRecord
	}
}
