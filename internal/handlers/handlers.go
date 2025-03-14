package handlers

import (
	"context"
	"time"

	"github.com/apetsko/gophermart/internal/logging"
	mw "github.com/apetsko/gophermart/internal/middleware"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Storage interface {
	AddOrder(ctx context.Context, userID int64, order string) error
	Balance(ctx context.Context, id int64) (*models.UserBalance, error)
	ListOrders(ctx context.Context, id int64) ([]models.UserOrderEntry, error)
	GetUser(ctx context.Context, login string) (*models.UserEntry, error)
	AddUser(ctx context.Context, u *models.UserEntry) (int, error)
	Withdraw(ctx context.Context, id int64, wd models.Withdraw, logger logging.Logger) (*models.UserBalance, error)
	Withdrawals(ctx context.Context, id int64) ([]models.Withdraw, error)
}

type URLHandler struct {
	Storage        Storage
	Secret         string
	AccrualProcess chan models.BatchDeleteRequest
	Logger         *logging.Logger
}

func New(s Storage, secret string, l *logging.Logger) *URLHandler {
	return &URLHandler{
		Storage:        s,
		Secret:         secret,
		Logger:         l,
		AccrualProcess: make(chan models.BatchDeleteRequest),
	}
}

func SetupRouter(handler *URLHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	compressionLevel := 5
	r.Use(middleware.Compress(compressionLevel))
	r.Use(middleware.Timeout(time.Minute))
	r.Use(middleware.StripSlashes)
	r.Use(mw.LoggingMiddleware(handler.Logger))
	r.Use(middleware.Recoverer)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", RegisterHandler(handler))
		r.Post("/login", LoginHandler(handler))

		r.Group(func(r chi.Router) {
			r.Use(mw.AuthMiddleware(handler.Secret, handler.Logger))
			r.Post("/orders", AddOrderHandler(handler))
			r.Get("/orders", ListOrdersHandler(handler))
			r.Get("/balance", BalanceHandler(handler))
			r.Post("/balance/withdraw", WithdrawHandler(handler))
			r.Get("/withdrawals", WithdrawalsHandler(handler))
		})
	})
	return r
}
