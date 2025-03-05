package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/utils"
)

func WithdrawHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
		defer cancel()
		defer r.Body.Close()

		userIDstr := r.Header.Get("userID")
		if userIDstr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(userIDstr, 10, 64)
		if err != nil {
			h.Logger.Error(fmt.Sprintf("parse user id err: %v", err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.Logger.Debug(fmt.Sprintf("UserID: %v", userID))

		b, err := io.ReadAll(r.Body)
		if err != nil {
			h.Logger.Error(fmt.Sprintf("read body err: %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var wd models.Withdraw
		err = json.Unmarshal(b, &wd)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = utils.ValidateStruct(wd); err != nil {
			h.Logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.Logger.Debug(fmt.Sprintf("Withdraw: %v", wd))
		if ok := utils.ValidateLuhnAlgorithm(wd.Order); !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ub, err := h.Storage.Withdraw(ctx, userID, wd, *h.Logger)
		if err != nil {
			status := http.StatusBadRequest

			if errors.Is(err, models.ErrInsufficientFunds) {
				status = http.StatusPaymentRequired
			}

			h.Logger.Error(err.Error())
			w.WriteHeader(status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(*ub); err != nil {
			h.Logger.Error("failed to encode response", "error", err)
		}
	}
}
