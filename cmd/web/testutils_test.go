package main

import (
	"bytes"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"html"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"snippetbox/internal/models/mocks"
	"strings"
	"testing"
	"time"
)

// Defines a custom testServer type which embeds a httptest.Server instance.
type testServer struct {
	*httptest.Server
}

// Returns an instance of the application struct containing mocked dependencies
func newTestApplication(t *testing.T) *application {
	// Create an instance of the template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	// Create an instance of the form decoder
	formDecoder := form.NewDecoder()

	// Create a sessionManager instance
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true
	//sessionManager.Cookie.Secure = false
	return &application{
		errorLog:       log.New(io.Discard, "", 0),
		infoLog:        log.New(io.Discard, "", 0),
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		links:          &mocks.LinkMappingModel{},
		files:          &mocks.FileModel{},
	}
}

// Initalizes and returns a new instance of the custom testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {
	// Initialize the test server as normal.
	ts := httptest.NewTLSServer(h)
	//ts := httptest.NewServer(h)

	// Initialize a new cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the test server client. Any response cookies will
	// now be stored and sent with subsequent requests when using this client.
	ts.Client().Jar = jar

	// Disables redirect-following for the test server client by setting a custom CheckRedirect function.
	// This function will be called whenever a 3xx response is received by the client, and by always returning a
	// http.ErrUseLastResponse error it forces the client to immediately return the received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// Makes a GET request to a given url path using the test server client, and returns the response status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

//// Makes a POST request to a given url with a provided form. Returns response status code, headers and body
//func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
//	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Read the response body from the test server.
//	defer rs.Body.Close()
//	body, err := io.ReadAll(rs.Body)
//	if err != nil {
//		t.Fatal(err)
//	}
//	bytes.TrimSpace(body)
//
//	// Return the response status, headers and body.
//	return rs.StatusCode, rs.Header, string(body)
//}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	req, err := http.NewRequest("POST", ts.URL+urlPath, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	// Need to manually add headers for some reason?
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", ts.URL+urlPath)

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("No CSRF token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}
