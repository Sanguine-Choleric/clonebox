package models

import (
	"snippetbox/internal/assert"
	"testing"
)

func TestFileModel_GetByChecksum(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name     string
		checksum string
		want     *File
		wantErr  error
	}{
		{
			name:     "File exists in db",
			checksum: "abcdef",
			want: &File{
				FileName:    "test_file.pdf",
				FileUUID:    "123456",
				Checksum:    "abcdef",
				StoragePath: "/clonebox/upload",
			},
			wantErr: nil,
		},
		{
			name:     "File doesn't exist in db",
			checksum: "qwerty",
			want:     nil,
			wantErr:  ErrNoRecord,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := FileModel{db}
			res, err := m.GetByChecksum(tt.checksum)
			assert.Equal(t, err, tt.wantErr)
			if res != nil {
				assert.Equal(t, res.Checksum, tt.want.Checksum)
				assert.Equal(t, res.FileName, tt.want.FileName)
				assert.Equal(t, res.FileUUID, tt.want.FileUUID)
			}
		})
	}
}

func TestFileModel_GetByUUID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name    string
		uuid    string
		want    *File
		wantErr error
	}{
		{
			name: "File exists in db",
			uuid: "123456",
			want: &File{
				FileName:    "test_file.pdf",
				FileUUID:    "123456",
				Checksum:    "abcdef",
				StoragePath: "/clonebox/upload",
			},
			wantErr: nil,
		},
		{
			name:    "File doesn't exist in db",
			uuid:    "qwerty",
			want:    nil,
			wantErr: ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := FileModel{db}
			res, err := m.GetByUUID(tt.uuid)
			assert.Equal(t, err, tt.wantErr)
			if res != nil {
				assert.Equal(t, res.Checksum, tt.want.Checksum)
				assert.Equal(t, res.FileName, tt.want.FileName)
				assert.Equal(t, res.FileUUID, tt.want.FileUUID)
			}
		})
	}
}

func TestFileModel_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name        string
		fileName    string
		fileUUID    string
		fileSize    int
		checksum    string
		storagePath string
		wantErr     error
	}{
		{
			name:        "Insert Success",
			fileName:    "test_file2.pdf",
			fileUUID:    "987654",
			fileSize:    1000,
			checksum:    "qwerty",
			storagePath: "/clonebox/upload",
			wantErr:     nil,
		},
		{
			name:        "Insert Duplicate Checksum",
			fileName:    "test_file3.pdf",
			fileUUID:    "987654",
			fileSize:    1000,
			checksum:    "abcdef",
			storagePath: "/clonebox/upload",
			wantErr:     nil,
		},
		{
			name:        "Insert Duplicate UUID",
			fileName:    "test_file4.pdf",
			fileUUID:    "123456",
			fileSize:    1000,
			checksum:    "qwerty",
			storagePath: "/clonebox/upload",
			wantErr:     ErrDuplicateUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := FileModel{db}
			err := m.Insert(tt.fileName, tt.fileUUID, tt.fileSize, tt.checksum, tt.storagePath)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}
