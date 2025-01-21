package mocks

import "snippetbox/internal/models"

type UserModel struct{}

func (m *UserModel) PasswordUpdate(id int, currentPassword string, newPassword string) error {
	//TODO implement me
	panic("implement me")
}

func (m *UserModel) Get(id int) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@mock.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "p@ssw0rd" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}
