package models

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"
)

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (string, error)
	Get(public_id string) (*Snippet, error)
	Latest() ([]*Snippet, error)
}

// Snippet defines a struct type to hold the data for an individual snippet. Struct fields correspond with MySQL fields
type Snippet struct {
	ID       int
	PublicID string
	Title    string
	Content  string
	Created  time.Time
	Expires  time.Time
}

// SnippetModel defines a type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title string, content string, expires int) (string, error) {
	stmt := `INSERT INTO snippets (public_id, title, content, created, expires) 
				VALUES(?, ?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Quick random id generation
	const letters = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
	random := make([]byte, 5)
	for i := range random {
		random[i] = letters[rand.Intn(len(letters))]
	}
	public_id := string(random)
	if len(random) != 5 {
		return "", errors.New("bad public id generation")
	}

	result, err := m.DB.Exec(stmt, public_id, title, content, expires)
	if err != nil {
		return "", errors.New("bad query exec")
	}

	// Gets ID of newly inserted record in the snippets table
	_, err = result.LastInsertId()
	if err != nil {
		return "", errors.New("bad latest insert id")
	}

	return public_id, nil
}

// Get returns a specific snippet based on its ID.
func (m *SnippetModel) Get(public_id string) (*Snippet, error) {
	stmt := `SELECT ID, public_id, title, content, created, expires FROM snippets
				WHERE expires > UTC_TIMESTAMP() AND public_id = ?`

	row := m.DB.QueryRow(stmt, public_id)

	// Zeroed Snippet struct.
	s := &Snippet{}

	// row.Scan() copies the values from each field in sql.Row to the corresponding field in struct
	err := row.Scan(&s.ID, &s.PublicID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

// Latest returns the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT ID, public_id, title, created, expires FROM snippets
				WHERE expires > UTC_TIMESTAMP() ORDER BY ID DESC LIMIT 10`

	// Query() returns a sql.Rows resultset (potentially multiple rows)
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// Defer sql.Rows resultset gets properly closed before Latest() method returns. Closing after error check as go
	// panics from trying to close a nil resultset.
	defer rows.Close()

	snippets := []*Snippet{}

	// rows.Next() iterates through the rows in the resultset. If all rows are iterated over, then resultset auto closes
	for rows.Next() {
		s := &Snippet{}

		err = rows.Scan(&s.ID, &s.PublicID, &s.Title, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		// Appended to slice of snippets.
		snippets = append(snippets, s)
	}

	// rows.Err() retrieves any error that was encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
