package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"

	"snippetbox/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	debug          *bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	links          models.LinkMappingModelInterface
	files          models.FilesModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	llmClient      *genai.Client
	llmConfig      *genai.GenerateContentConfig
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	addr := flag.String("addr", ":4000", "HTTP network address")

	raw_password, err := os.ReadFile("/run/secrets/db_password")
	if err != nil {
		errorLog.Printf("%s", err)
	}

	DB_PASS := strings.TrimSpace(string(raw_password))

	default_dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		DB_PASS,
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	// infoLog.Printf("%s", default_dsn)

	dsn := flag.String("dsn", default_dsn, "MySQL data source name")
	debug := flag.Bool("debug", false, "Enables debug mode (stack traces)")

	flag.Parse()

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// Gemini integration - doing this here so i don't create a new geminiClient for every request
	llm_key, err := os.ReadFile("/run/secrets/web_llm_api_key")
	API_KEY := strings.TrimSpace(string(llm_key))
	if err != nil {
		errorLog.Printf("%s", err)
	}
	ctx := context.Background()
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		// FIXME: Use an env
		APIKey:  API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		errorLog.Fatal(err)
	}

	geminiConfig := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name":     {Type: genai.TypeString},
					"price":    {Type: genai.TypeNumber},
					"quantity": {Type: genai.TypeInteger},
				},
				PropertyOrdering: []string{"name", "price", "quantity"},
			},
		},
	}

	app := &application{
		debug:          debug,
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		links:          &models.LinkMappingModel{DB: db},
		files:          &models.FileModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		llmClient:      geminiClient,
		llmConfig:      geminiConfig,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)

	//err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
