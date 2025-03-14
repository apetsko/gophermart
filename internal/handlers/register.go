package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/apetsko/gophermart/internal/auth"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var user models.User
		err = json.Unmarshal(b, &user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = utils.ValidateStruct(user); err != nil {
			h.Logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		u := &models.UserEntry{
			Username:     user.Login,
			PasswordHash: string(hash),
		}

		userID, err := h.Storage.AddUser(ctx, u)
		if err != nil {
			if errors.Is(err, models.ErrUserExists) {
				h.Logger.Error(err.Error())
				w.WriteHeader(http.StatusConflict)
				return
			}
			h.Logger.Error("failed to add user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = auth.CookieSetUserID(w, userID, h.Secret); err != nil {
			h.Logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		h.Logger.Info("registered user", "user", user.Login)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
