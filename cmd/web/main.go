package main

import (
	"crypto/tls"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/msdoshack/mycodedairy/internal/models"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippetss      *models.SnippetModel
	users          *models.UserModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

// USED LOCALLY
// func init() {
// 	if err := godotenv.Load(".env"); err != nil {
// 		log.Fatal("error loading env file")
// 	}
// }

func main() {
	// CREATING CUSTOM LOG
	dsn := os.Getenv("DSN") // Ensure this is formatted correctly for PostgreSQL
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(dsn)
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	// Use pgstore for PostgreSQL session management
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		infoLog:        infoLog,
		errorLog:       errLog,
		snippetss:      &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}


	port := os.Getenv("PORT")
	if port == "" {
		port = "4000" // Default to port 4000 if PORT is not set
	}

	srv := &http.Server{
		Addr:         ":"+port,
		ErrorLog:     errLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Println("Server running on port "+port)


	app.hitEndPoint()


	// WITHOUT TLS CERTIFICATE
	err = srv.ListenAndServe()

	// WITH TLS CERTIFICATE (USE ONLY LOCALLY)
	// err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")


	if err != nil {
		errLog.Fatal(err)
	}

}


func openDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}



// WRITING THE LOG TO A DISK
// go run ./cmd/web >>./temp/info.log




