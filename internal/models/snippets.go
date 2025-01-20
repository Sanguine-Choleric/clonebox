package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
}

// Snippet defines a struct type to hold the data for an individual snippet. Struct fields correspond with MySQL fields
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// SnippetModel defines a type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {

	// The SQL statement to execute
	stmt := `INSERT INTO snippets (title, content, created, expires) 
				VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Exec() method on the embedded connection pool executes the statement.
	// The first parameter is the SQL statement, followed by the title, content and expiry values for the placeholder
	// parameters.
	// Returns a sql.Result type, which contains some basic information about what happened when the statement was executed.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the result to get the ID of our newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

// Get returns a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
				WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// QueryRow() method on the connection pool executes our SQL statement, passing in the untrusted id variable as
	// the value for the placeholder parameter.
	// This returns a pointer to a sql.Row object which holds the result from the database.
	row := m.DB.QueryRow(stmt, id)

	// Initialize a pointer to a new zeroed Snippet struct.
	s := &Snippet{}

	// row.Scan() copies the values from each field in sql.Row to the corresponding field in the Snippet struct.
	// Arguments to row.Scan are pointers to the place to copy the data into, and the number of arguments must be
	// exactly the same as the number of columns returned by the statement.
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything OK then return the Snippet object.
	return s, nil
}

// Latest returns the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, created, expires FROM snippets
				WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	// Query() returns a sql.Rows resultset containing the result of our query.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// Defer rows.Close() is to ensure the sql.Rows resultset is always properly closed before the Latest() method
	// returns. Should come *after* you check for an error from the Query() method. Otherwise, if Query() returns an
	// error, go panics from trying to close a nil resultset.
	defer rows.Close()

	// Holds Snippet structs.
	snippets := []*Snippet{}

	// rows.Next iterates through the rows in the resultset. This prepares the first (and then each subsequent) row to
	// be acted on by the rows.Scan() method. If iteration over all the rows completes then the resultset automatically
	// closes itself and frees-up the underlying database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed Snippet struct.
		s := &Snippet{}

		// rows.Scan() copies the values from each field in the row to the new Snippet object that we created. The
		// arguments to row.Scan() must be pointers to the place you want to copy the data into, and the number of
		// arguments must be exactly the same as the number of columns returned by the statement.
		err = rows.Scan(&s.ID, &s.Title, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		// Append it to the slice of snippets.
		snippets = append(snippets, s)
	}

	// When loop has finished, rows.Err() retrieves any error that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK then return the Snippets slice.
	return snippets, nil
}
