package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/RiMaBo/go_http_server/internal/auth"
	"github.com/RiMaBo/go_http_server/internal/database"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string    `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Email) < 1 || len(params.Password) < 1 {
		respondWithError(w, http.StatusBadRequest, `Usage: { "email": "user@example.com", "password": "pa$$word" }`, nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	usr, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        usr.ID,
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
		Email:     usr.Email,
	})
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string            `json:"email"`
		Password string         `json:"password"`
		ExpiresIn time.Duration `json:"expires_in_seconds"`
	}
	type response struct {
		User
		Token string `json:"token"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Email) < 1 || len(params.Password) < 1 {
		respondWithError(w, http.StatusBadRequest, `Usage: { "email": "user@example.com", "password": "pa$$word", "expires_in_seconds": x (optional) }`, nil)
		return
	}

	usr, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, usr.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	expiresIn := params.ExpiresIn * time.Second
	if expiresIn.Seconds() == 0 || expiresIn.Seconds() > 3600 {
		expiresIn = time.Hour
	}

	token, err := auth.MakeJWT(usr.ID, cfg.secret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating token", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        usr.ID,
			CreatedAt: usr.CreatedAt,
			UpdatedAt: usr.UpdatedAt,
			Email:     usr.Email,
		},
		Token: token,
	})
}
