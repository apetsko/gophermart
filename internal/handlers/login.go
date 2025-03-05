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
)

func LoginHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
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

		ue, err := h.Storage.GetUser(ctx, user.Login)

		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				h.Logger.Error(err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			return
		}

		if ok := utils.ComparePassword(ue.PasswordHash, user.Password); !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = auth.CookieSetUserID(w, ue.ID, h.Secret)
		if err != nil {
			return
		}

		h.Logger.Info("success login", "user", user.Login)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(b); err != nil {
			h.Logger.Error(err.Error())
		}
	}
}
