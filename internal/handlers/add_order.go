package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/utils"
)

func AddOrderHandler(h *URLHandler) func(http.ResponseWriter, *http.Request) {
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

		orderBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.Logger.Error(fmt.Sprintf("read body err: %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var order = string(orderBytes)
		if len(orderBytes) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.Logger.Debug(fmt.Sprintf("Order: %s", order))
		if ok := utils.ValidateLuhnAlgorithm(string(orderBytes)); !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if err = h.Storage.AddOrder(ctx, userID, order); err != nil {
			status := http.StatusBadRequest

			if errors.Is(err, models.ErrOrderExistsForAnotherUser) {
				status = http.StatusConflict
			} else if errors.Is(err, models.ErrOrderAlreadyExists) {
				status = http.StatusOK
			}

			h.Logger.Error(err.Error())
			w.WriteHeader(status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, err = w.Write(orderBytes)
		if err != nil {
			return
		}
	}
}
