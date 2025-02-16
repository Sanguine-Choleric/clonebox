package mocks

import (
	"snippetbox/internal/models"
)

type LinkMappingModel struct{}

func (m *LinkMappingModel) Insert(original, short string) error {
	//TODO implement me
	panic("implement me")
}

func (m *LinkMappingModel) Latest() ([]*models.LinkMapping, error) {
	//TODO implement me
	panic("implement me")
}

func (m *LinkMappingModel) GetOriginal(short string) (string, error) {
	switch short {
	case "abcde":
		return "https://existent.com", nil
	default:
		return "", models.ErrNoRecord
	}
}

func (m *LinkMappingModel) GetShort(original string) (string, error) {
	switch original {
	case "https://existent.com":
		return "abcde", nil
	default:
		return "", models.ErrNoRecord
	}
}

func (m *LinkMappingModel) Exists(original string) (bool, error) {
	switch original {
	case "https://nonexistent.com":
		return false, nil
	default:
		return true, nil
	}
}
