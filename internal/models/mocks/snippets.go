package mocks

import (
	"snippetbox/internal/models"
	"time"
)

var mockSnippet = &models.Snippet{
	ID:       1,
	PublicID: "A",
	Title:    "Test title",
	Content:  "Test content",
	Created:  time.Now(),
	Expires:  time.Now(),
}

type SnippetModel struct{}

func (m *SnippetModel) Insert(title, content string, expires int) (string, error) {
	return "B", nil
}

func (m *SnippetModel) Get(public_id string) (*models.Snippet, error) {
	switch public_id {
	case "A":
		return mockSnippet, nil
	default:
		return nil, models.ErrNoRecord
	}
}

func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}
