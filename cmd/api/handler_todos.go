package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userID, err := auth.ValidateJWT(token, app.cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	todoID, err := uuid.Parse(r.PathValue("todoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID", err)
		return
	}

	existingTodo, err := app.db.GetTodoByID(r.Context(), todoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "ToDo not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if existingTodo.UserID != userID {
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

	updatedTodo, err := app.db.UpdateTodo(r.Context(), database.UpdateTodoParams{
		ID:          existingTodo.ID,
		UserID:      existingTodo.UserID,
		Title:       params.Title,
		Description: params.Description,
	})
	if err != nil {
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

func (app *application) handlerTodoDelete(w http.ResponseWriter, r *http.Request) {
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

	todoID, err := uuid.Parse(r.PathValue("todoID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid todo ID", err)
		return
	}

	existingTodo, err := app.db.GetTodoByID(r.Context(), todoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "ToDo not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	if existingTodo.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	if err = app.db.DeleteTodo(r.Context(), database.DeleteTodoParams{
		ID:     existingTodo.ID,
		UserID: existingTodo.UserID,
	}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handlerTodoRetrieve(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Data  []ToDo `json:"data"`
		Page  int    `json:"page"`
		Limit int    `json:"limit"`
		Total int    `json:"total"`
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

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	dbTodos, err := app.db.GetTodosByUserID(r.Context(), database.GetTodosByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	count, err := app.db.GetTodosCountByUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
		return
	}

	todos := make([]ToDo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = ToDo{
			ID:          dbTodo.ID,
			CreatedAt:   dbTodo.CreatedAt,
			UpdatedAt:   dbTodo.UpdatedAt,
			Title:       dbTodo.Title,
			Description: dbTodo.Description,
			UserID:      dbTodo.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, response{
		Data:  todos,
		Page:  page,
		Limit: limit,
		Total: int(count),
	})
}
