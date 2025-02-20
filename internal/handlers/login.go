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
)

func LoginHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
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

		ue, err := h.storage.GetUser(ctx, user.Login)

		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				h.logger.Error(err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			return
		}

		if ok := auth.ComparePassword(ue.PasswordHash, user.Password); !ok {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		err = auth.CookieSetUserID(w, ue.ID, h.secret)
		if err != nil {
			return
		}

		h.logger.Info("registered user", "user", user.Login)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		if err != nil {
			return
		}
	}
}
