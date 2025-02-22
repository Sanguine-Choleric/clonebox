package models

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	PasswordUpdate(id int, currentPassword string, newPassword string) error
}

// User field names and types align with the columns in the database "users" table
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

// UserModel wraps a database connection pool.
type UserModel struct {
	DB *sql.DB
}

// Insert adds a new record to the "users" table.
func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plaintext password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
			 VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {

		// If this returns an error, errors.As() checks whether the error has the type *mysql.MySQLError. If it does, the
		// error will be assigned to the mySQLError variable. Then check whether the error relates to the users_uc_email
		// key by checking if the error code equals 1062 and the contents of the error message string. If it does,
		// returns an ErrDuplicateEmail error.
		var mySqlErr *mysql.MySQLError
		if errors.As(err, &mySqlErr) {
			if mySqlErr.Number == 1062 && strings.Contains(mySqlErr.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
	}

	return nil
}

// Authenticate verifies whether a user exists with provided email address and password. This will return the relevant
// user ID if they do.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Retrieve the ID and hashed password associated with the given email. If no matching email exists, returns ErrInvalidCredentials
	var id int
	var hashedPassword []byte

	stmt := `SELECT ID, hashed_password FROM users WHERE email = ?`

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Check whether the hashed password and plain-text password provided match. If not, return the ErrInvalidCredentials
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

// Exists checks if a user exists with a specific ID.
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE ID = ?)"

	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

func (m *UserModel) Get(id int) (*User, error) {
	var user User

	stmt := `SELECT ID, name, email, created FROM users WHERE ID = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	//if errors.Is(err, sql.ErrNoRows) {
	//	return nil, ErrNoRecord
	//} else if err != nil {
	//	return nil, err
	//}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return &user, nil

}

func (m *UserModel) PasswordUpdate(id int, currentPassword string, newPassword string) error {
	// Retrieve user from ID
	var user User
	stmt := `SELECT * FROM users WHERE ID = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.Created)
	//if errors.Is(err, sql.ErrNoRows) {
	//	return ErrNoRecord
	//} else if err != nil {
	//	return err
	//}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		} else {
			return err
		}
	}

	// Check if currentPassword matches password stored in db
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}

	// Update
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt = `UPDATE users SET hashed_password = ? WHERE ID = ?`
	_, err = m.DB.Exec(stmt, string(hashedNewPassword), id)
	if err != nil {
		return err
	}

	return nil
}
