package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AsherBolleddu/TodoListGo/internal/auth"
	"github.com/AsherBolleddu/TodoListGo/internal/database"
	"github.com/lib/pq"
)

func (app *application) handlerUserRegister(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		Token string `json:"token"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := app.db.CreateUser(r.Context(), database.CreateUserParams{
		Name:           params.Name,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				respondWithError(w, http.StatusConflict, "Email already exists", nil)
				return
			}
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, app.cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Token: token,
	})
}
