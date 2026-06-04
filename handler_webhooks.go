package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RiMaBo/go_http_server/internal/auth"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string    `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		}
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding body", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, "")
		return
	}

	_, err = cfg.db.UpgradeUser(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Error upgrading user", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
