package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AsherBolleddu/TodoListGo/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type config struct {
	env       string
	jwtSecret string
	port      string
}

type application struct {
	cfg config
	db  *database.Queries
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	env := os.Getenv("ENV")
	if env == "" {
		log.Fatal("ENV is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	cfg := &config{
		env:       env,
		jwtSecret: jwtSecret,
		port:      port,
	}

	app := &application{
		db:  dbQueries,
		cfg: *cfg,
	}

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":" + app.cfg.port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	mux.HandleFunc("POST /admin/reset", app.handlerReset)

	mux.HandleFunc("POST /register", app.handlerUserRegister)
	mux.HandleFunc("POST /login", app.handlerUserLogin)

	mux.HandleFunc("GET /todos", app.handlerTodoRetrieve)
	mux.HandleFunc("POST /todos", app.handlerTodoCreate)
	mux.HandleFunc("PUT /todos/{todoID}", app.handlerTodoUpdate)
	mux.HandleFunc("DELETE /todos/{todoID}", app.handlerTodoDelete)

	log.Printf("Starting server at http://localhost%s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
