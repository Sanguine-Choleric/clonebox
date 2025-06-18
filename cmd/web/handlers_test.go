package main

import (
	"net/http"
	"net/url"
	"snippetbox/internal/assert"
	"testing"
)

func TestPing(t *testing.T) {
	t.Parallel()
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}

func TestSnippetView(t *testing.T) {
	t.Parallel()
	// Create a new instance of our application struct which uses the mocked dependencies
	app := newTestApplication(t)

	// Establish a new test server for running end-to-end tests
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Set up some table-driven tests to check the responses sent by our application for different URLs
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/view/A",
			wantCode: http.StatusOK,
			wantBody: "Test content",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/view/ABC",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/snippet/view/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/snippet/view/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/snippet/view/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/snippet/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestUserSignup(t *testing.T) {
	t.Parallel()
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)
	const (
		validName     = "Bob"
		validPassword = "validPa$$word"
		validEmail    = "bob@example.com"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "dupe@mock.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)
			code, _, body := ts.postForm(t, "/user/signup", form)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantFormTag != "" {
				if tt.wantFormTag != "" {
					assert.StringContains(t, body, tt.wantFormTag)
				}
			}
		})
	}
}

func TestSnippetCreate(t *testing.T) {
	t.Parallel()
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("Unauthorized Snippet Create", func(t *testing.T) {
		code, header, _ := ts.get(t, "/snippet/create")
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, header.Get("Location"), "/user/login")
	})

	t.Run("Authorized Snippet Create", func(t *testing.T) {
		// Grab csrf token
		_, _, body := ts.get(t, "/user/signup")
		validCSRFToken := extractCSRFToken(t, body)

		// Prep and make authn request
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "p@ssw0rd")
		form.Add("csrf_token", validCSRFToken)

		code, _, _ := ts.postForm(t, "/user/login", form)
		assert.Equal(t, code, http.StatusSeeOther)

		code, _, body = ts.get(t, "/snippet/create")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, "<form action='/snippet/create' method='POST'>")

	})
}

func TestLinkShorten(t *testing.T) {
	t.Parallel()
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("Unauthorized Link Shorten", func(t *testing.T) {
		code, header, _ := ts.get(t, "/shorten")
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, header.Get("Location"), "/user/login")
	})

	t.Run("Authorized Link Shorten", func(t *testing.T) {
		// Logging in user
		_, _, body := ts.get(t, "/user/signup")
		validCSRFToken := extractCSRFToken(t, body)
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "p@ssw0rd")
		form.Add("csrf_token", validCSRFToken)
		code, _, _ := ts.postForm(t, "/user/login", form)
		assert.Equal(t, code, http.StatusSeeOther)

		code, _, body = ts.get(t, "/shorten")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, "<form action='/shorten' method='POST'")

		form2 := url.Values{}
		form2.Add("original_link", "https://existent.com")
		form2.Add("csrf_token", validCSRFToken)

		code, _, _ = ts.postForm(t, "/shorten", form2)
		assert.Equal(t, code, http.StatusOK)
	})

	t.Run("Bad Form", func(t *testing.T) {
		_, _, body := ts.get(t, "/user/signup")
		validCSRFToken := extractCSRFToken(t, body)
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "p@ssw0rd")
		form.Add("csrf_token", validCSRFToken)
		_, _, _ = ts.postForm(t, "/user/login", form)

		form = url.Values{}
		form.Add("original_link", "?nval?durl")
		form.Add("csrf_token", validCSRFToken)

		code, _, body := ts.postForm(t, "/shorten", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)

	})
}

func TestFileUpload(t *testing.T) {
	t.Parallel()
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("Unauthorized File Upload", func(t *testing.T) {
		code, header, _ := ts.get(t, "/file")
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, header.Get("Location"), "/user/login")
	})

	t.Run("Authorized File Upload", func(t *testing.T) {
		// Logging in user
		_, _, body := ts.get(t, "/user/signup")
		validCSRFToken := extractCSRFToken(t, body)
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "p@ssw0rd")
		form.Add("csrf_token", validCSRFToken)
		code, _, _ := ts.postForm(t, "/user/login", form)
		assert.Equal(t, code, http.StatusSeeOther)

		// Upload test
		code, _, body = ts.get(t, "/file")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, "<form action=\"/file\" method=\"post\" enctype=\"multipart/form-data\">")

		// TODO: Multipart form upload test
	})
}

func TestFileView(t *testing.T) {
	t.Parallel()
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid UUID",
			urlPath:  "/file/view/123456",
			wantCode: http.StatusOK,
			wantBody: "test_file.pdf", // Contains test, don't need exact
		},
		{
			name:     "Non-existent UUID",
			urlPath:  "/file/view/987654",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty UUID",
			urlPath:  "/file/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})

	}
}
