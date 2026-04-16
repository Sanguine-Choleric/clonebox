package models

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newTestDB(t *testing.T) *sql.DB {
	// Establishes a sql.DB connection pool for test database
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("test_web", "pass"),
		Host:   os.Getenv("DB_HOST") + ":5432",
		Path:   fmt.Sprintf("%s%s", "test_", os.Getenv("DB_NAME")),
	}
	q := dsn.Query()
	q.Set("sslmode", "disable")
	dsn.RawQuery = q.Encode()

	db, err := sql.Open("pgx", dsn.String())
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
		script, err = os.ReadFile("./testdata/teardown.sql")
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
