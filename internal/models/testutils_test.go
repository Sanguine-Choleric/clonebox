package models

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	// Establishes a sql.DB connection pool for test database. Because setup and teardown scripts contains multiple SQL
	// statements, have to use the "multiStatements=true" parameter in our DSN. This instructs MySQL database driver to
	// support executing multiple SQL statements in one db.Exec() call.
	dsn := fmt.Sprintf("%s:%s@tcp(db)/%s?parseTime=true&multiStatements=true",
		"test_web",
		"pass",
		"test_snippetbox")
	//db, err := sql.Open("mysql", "test_web:pass@/test_snippetbox?parseTime=true&multiStatements=true")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}

	// Reads the setup SQL script from file and execute the statements.
	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// t.Cleanup() registers a function which will automatically be called by Go when the current test (or sub-test)
	// which calls newTestDB() has finished. Reads and executes the teardown script, and close the database connection pool.
	t.Cleanup(func() {
		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
		db.Close()
	})

	// Return the database connection pool.
	return db
}
