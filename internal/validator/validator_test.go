package validator

import (
	"testing"
)

func TestPermittedMimeType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		accepted []string
		want     bool
	}{
		{
			name:     "valid mime type in list",
			input:    "image/jpeg",
			accepted: []string{"image/jpeg", "image/png"},
			want:     true,
		},
		{
			name:     "valid mime type at end of list",
			input:    "application/json",
			accepted: []string{"image/jpeg", "image/png", "application/json"},
			want:     true,
		},
		{
			name:     "invalid mime type not in list",
			input:    "text/plain",
			accepted: []string{"image/jpeg", "image/png"},
			want:     false,
		},
		{
			name:     "empty mime type with empty list",
			input:    "",
			accepted: []string{},
			want:     false,
		},
		{
			name:     "empty mime type with non-empty list",
			input:    "",
			accepted: []string{"image/jpeg", "image/png"},
			want:     false,
		},
		{
			name:     "empty list with non-empty mime type",
			input:    "image/jpeg",
			accepted: []string{},
			want:     false,
		},
		{
			name:     "case insensitive match",
			input:    "image/jpeg",
			accepted: []string{"Image/Jpeg", "image/png"},
			want:     false,
		},
		{
			name:     "whitespace in mime type",
			input:    " image/jpeg ",
			accepted: []string{"image/jpeg", "image/png"},
			want:     true,
		},
		{
			name:     "parameterized mime type match",
			input:    "text/plain; charset=utf8",
			accepted: []string{"text/plain", "image/png"},
			want:     true,
		},
		{
			name:     "parameterized mime type mismatch",
			input:    "text/plain; charset=utf8",
			accepted: []string{"image/png"},
			want:     false,
		},

		{
			name:     "parameterized mime type extra spacing",
			input:    "text/plain; charset=utf8 ",
			accepted: []string{"text/plain", "image/png"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PermittedMimeType(tt.input, tt.accepted)
			if got != tt.want {
				t.Errorf("PermittedMimeType(%q, %v) = %v; want %v", tt.input, tt.accepted, got, tt.want)
			}
		})
	}
}
