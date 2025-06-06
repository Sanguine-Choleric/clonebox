package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"snippetbox/internal/assert"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	// Initialize a new httptest.ResponseRecorder
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Creates a mock HTTP handler that gets passed the secureHeaders() middleware, which writes a 200 status code and an "OK" response body
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Passes the mock HTTP handler to our secureHeaders middleware. Because secureHeaders returns a http.Handler, call its ServeHTTP() method,
	// Passing in the http.ResponseRecorder and dummy http.Request to execute it
	secureHeaders(next).ServeHTTP(rr, r)

	// Call the Result() method on the http.ResponseRecorder to get the results of the test
	rs := rr.Result()

	// Check that the middleware has correctly set the Content-Security-Policy header on the response
	expectedVal := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	assert.Equal(t, rs.Header.Get("Content-Security-Policy"), expectedVal)

	// Check that the middleware has correctly set the Referrer-Policy header on the response.
	expectedVal = "origin-when-cross-origin"
	assert.Equal(t, rs.Header.Get("Referrer-Policy"), expectedVal)

	// Check that the middleware has correctly set the X-Content-Type-Options header on the response.
	expectedVal = "nosniff"
	assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), expectedVal)

	// Check that the middleware has correctly set the X-Frame-Options header on the response.
	expectedVal = "deny"
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), expectedVal)

	// Check that the middleware has correctly set the X-XSS-Protection header on the response
	expectedVal = "0"
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), expectedVal)

	// Check that the middleware has correctly called the next handler in line and the response status code and body are as expected.
	assert.Equal(t, rs.StatusCode, http.StatusOK)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")
}
