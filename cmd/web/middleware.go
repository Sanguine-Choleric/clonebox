package main

import (
	"context"
	"fmt"
	"github.com/justinas/nosurf"
	"log"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var remote string
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			remote = forwarded
		} else {
			remote = r.RemoteAddr
		}

		app.infoLog.Printf("%s - %s %s %s", remote, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Deferred function (which will always be run in the event of a panic as Go unwinds the stack)
		defer func() {

			// Use the builtin recover function to check if there has been a panic
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If the user is not authenticated, redirect them to the login page and return from the middleware chain so
		// that no subsequent handlers in the chain are executed.
		if !app.isAuthenticated(r) {
			app.sessionManager.Put(r.Context(), "originalPath", r.URL.Path)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Pages requiring authentication not stored in the users browser cache (or other intermediary cache)
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// CSRF Handling with noSurf
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})

	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CSRF validation failed for path: %s, reason: %v", r.URL.Path, nosurf.Reason(r))
		http.Error(w, "CSRF validation failed", http.StatusBadRequest)
		log.Println("Cookies on failed request:")
		for _, cookie := range r.Cookies() {
			log.Printf("  Name: %s, Value: %s, Path: %s, Secure: %t, HttpOnly: %t",
				cookie.Name, cookie.Value, cookie.Path, cookie.Secure, cookie.HttpOnly)
		}
	}))

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Retrieves the authenticatedUserID value from the session using GetInt(). This will return the zero value for
		// an int (0) if no "authenticatedUserID" value is in the session -- in which case we call the next handler in
		// the chain as normal and return.
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserId")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, check to see if a user with that ID exists in database.
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// If a matching user is found, we know that the request is coming from an authenticated user who exists in db.
		// Also creates a new copy of the request (with an isAuthenticatedContextKey value of true in the request context)
		// and assign it to r.
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
