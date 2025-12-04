package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/AsherBolleddu/TodoListGo/internal/auth"
	"github.com/AsherBolleddu/TodoListGo/internal/database"
	"github.com/AsherBolleddu/TodoListGo/internal/validation"
	"github.com/google/uuid"
)

type ToDo struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UserID      uuid.UUID `json:"-"`
}

func (app *application) handlerTodoCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userID, err := auth.ValidateJWT(token, app.cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	var params parameters
	if err = json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if !validation.IsNotEmpty(params.Title, params.Description) {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	todo, err := app.db.CreateTodo(r.Context(), database.CreateTodoParams{
		Title:       params.Title,
		Description: params.Description,
		UserID:      userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, ToDo{
		ID:          todo.ID,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
		Title:       todo.Title,
		Description: todo.Description,
		UserID:      todo.UserID,
	})
}

func (app *application) handlerTodoUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	userID, err := auth.ValidateJWT(token, app.cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	var params parameters
	if err = json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if !validation.IsNotEmpty(params.Title, params.Description) {
		respondWithError(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	todoID, err := uuid.Parse(r.PathValue("todoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID", err)
		return
	}

	updatedTodo, err := app.db.UpdateTodo(r.Context(), database.UpdateTodoParams{
		ID:          todoID,
		UserID:      userID,
		Title:       params.Title,
		Description: params.Description,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "ToDo not found", nil)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, ToDo{
		ID:          updatedTodo.ID,
		CreatedAt:   updatedTodo.CreatedAt,
		UpdatedAt:   updatedTodo.UpdatedAt,
		Title:       updatedTodo.Title,
		Description: updatedTodo.Description,
		UserID:      updatedTodo.UserID,
	})
}
