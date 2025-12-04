package main

import (
	"net/http"
)

func (app *application) handlerReset(w http.ResponseWriter, r *http.Request) {
	if app.cfg.env != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset only allowed on dev"))
		return
	}

	if err := app.db.Reset(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database reset to initial state"))
}
