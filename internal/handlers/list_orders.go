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

func ListOrdersHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
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

		lo, err := h.Storage.ListOrders(ctx, userID)
		if err != nil {
			if errors.Is(err, models.ErrOrderNotFound) {
				h.Logger.Debug(err.Error())
				w.WriteHeader(http.StatusNoContent)
				return
			}

			h.Logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		h.Logger.Debugf("OrderList: %v", lo)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(lo); err != nil {
			h.Logger.Error("failed to encode response", "error", err)
		}
	}
}
