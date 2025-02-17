package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/apetsko/gophermart/internal/auth"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		var user models.User
		err = json.Unmarshal(b, &user)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		validate := validator.New()
		if err = validate.Struct(user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			h.logger.Info(fmt.Sprintf("%v", validationErrors))
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		u := &models.UserEntry{
			ID:           0,
			Username:     user.Login,
			PasswordHash: string(hash),
		}

		userID, err := h.storage.AddUser(ctx, u)
		if err != nil {
			if errors.Is(err, models.ErrUserExists) {
				h.logger.Error(err.Error())
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			h.logger.Error("failed to add user", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err = auth.SetUserIDCookie(w, userID, h.secret); err != nil {
			h.logger.Error(err.Error())
			http.Error(w, "", http.StatusInternalServerError)
		}

		h.logger.Info("registered user", "user", user.Login)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(user); err != nil {
			h.logger.Error("failed to encode response", "error", err)
		}
	}
}
