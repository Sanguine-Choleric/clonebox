package models

import (
	"snippetbox/internal/assert"
	"testing"
)

func TestLinkMappingModel_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name         string
		originalLink string
		want         bool
	}{
		{
			name:         "Link exists in db",
			originalLink: "https://existent.com",
			want:         true,
		},
		{
			name:         "Link doesn't exist in db",
			originalLink: "https://nonexistent.com",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := LinkMappingModel{db}

			exists, err := m.Exists(tt.originalLink)
			assert.Equal(t, exists, tt.want)
			assert.NilError(t, err)
		})
	}
}

func TestLinkMappingModel_GetOriginal(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name    string
		short   string
		want    string
		wantErr error
	}{
		{
			name:    "Short Link Exists",
			short:   "123456",
			want:    "https://existent.com",
			wantErr: nil,
		},
		{
			name:    "Short Link Does Not Exist",
			short:   "abcdef",
			want:    "",
			wantErr: ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := LinkMappingModel{db}
			original, err := m.GetOriginal(tt.short)
			assert.Equal(t, original, tt.want)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}

func TestLinkMappingModel_GetShort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name     string
		original string
		want     string
		wantErr  error
	}{
		{
			name:     "Original Link Exists",
			original: "https://existent.com",
			want:     "123456",
			wantErr:  nil,
		},
		{
			name:     "Original Link Does Not Exist",
			original: "https://nonexistent.com",
			want:     "",
			wantErr:  ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := LinkMappingModel{db}
			short, err := m.GetShort(tt.original)
			assert.Equal(t, short, tt.want)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}

func TestLinkMappingModel_Insert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := []struct {
		name     string
		original string
		short    string
		wantErr  error
	}{
		{
			name:     "Insert Success",
			original: "https://newLink.com",
			short:    "abcdef",
			wantErr:  nil,
		},
		{
			name:     "Insert Duplicate",
			original: "https://existent.com",
			short:    "123456",
			wantErr:  ErrDuplicateLink,
		},
		{
			name:     "Insert Hash Collision",
			original: "https://hashCollision.com",
			short:    "123456",
			wantErr:  ErrDuplicateLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			m := LinkMappingModel{db}

			err := m.Insert(tt.original, tt.short)
			assert.Equal(t, err, tt.wantErr)
			if err == nil {
				o, err2 := m.GetOriginal(tt.short)
				assert.Equal(t, o, tt.original)
				assert.NilError(t, err2)

				s, err2 := m.GetShort(tt.original)
				assert.Equal(t, s, tt.short)
				assert.NilError(t, err2)
			}
		})
	}
}
