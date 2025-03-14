package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/apetsko/gophermart/internal/models"
)

func WithdrawalsHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
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

		ww, err := h.Storage.Withdrawals(ctx, userID)
		if err != nil {
			status := http.StatusBadRequest

			if errors.Is(err, models.ErrWithdrawalsNotFound) {
				status = http.StatusNoContent
				h.Logger.Debug("withdrawals not found", "userID", userID)
			}

			h.Logger.Error(err.Error(), "userID", userID)
			w.WriteHeader(status)
			return
		}

		var wwr []models.WithdrawRequest
		for _, w := range ww {
			wr := models.WithdrawRequest{
				Order:       w.Order,
				Sum:         float64(w.SumMinor) / 100.0,
				ProcessedAt: w.ProcessedAt,
			}
			wwr = append(wwr, wr)
		}

		h.Logger.Debug("withdrawals found", "withdrawals", ww, "userID", userID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(wwr); err != nil {
			h.Logger.Error("failed to encode response", "error", err)
		}
	}
}
