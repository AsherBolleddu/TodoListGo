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
	jwtSecret string
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

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	cfg := &config{
		jwtSecret: jwtSecret,
	}

	app := &application{
		db:  dbQueries,
		cfg: *cfg,
	}

	mux.HandleFunc("POST /register", app.handlerUserRegister)
	mux.HandleFunc("POST /login", app.handlerUserLogin)

	log.Printf("Starting server at http://localhost%s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
