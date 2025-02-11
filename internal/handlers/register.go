package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/storage/postgres"
	"github.com/go-playground/validator/v10"
)

func RegisterHandler(st *postgres.Storage, logger *logging.ZapLogger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
			logger.Info(fmt.Sprintf("%v", validationErrors))
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = st.AddUser(user.Login, user.Password)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				msg := "user already exists"
				logger.Errorw(msg, "error", err)
				http.Error(w, msg, http.StatusConflict)
				return
			}
		}
		logger.Info("registered user", "user", user)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		if err != nil {
			return
		}
		//TODO CHECK DB USER ACCOUNT
		//
		//IF EXIST 409
		//ADD IF NOT EXIST 200
		//INTERNAL ELSE
	}
}
