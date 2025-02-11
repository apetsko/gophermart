package handlers

import (
	"time"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type URLHandler struct {
	baseURL string
	storage *postgres.Storage
	secret  string
	//ToDelete chan models.BatchDeleteRequest
	logger *logging.ZapLogger
}

func New(baseURL string, s *postgres.Storage, l *logging.ZapLogger, secret string) *URLHandler {
	return &URLHandler{
		baseURL: baseURL,
		storage: s,
		logger:  l,
		secret:  secret,
		//ToDelete: make(chan models.BatchDeleteRequest),
	}
}

func SetupRouter(handler *URLHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))

	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Recoverer)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", RegisterHandler(handler.storage, handler.logger)) // регистрация пользователя;
		//r.Post("/login", handler.Login)            // аутентификация пользователя;
		//r.Get("/withdrawals", handler.Withdrawals) //  получение информации о выводе средств с накопительного счёта пользователем.
		//
		//r.Post("/orders", handler.AddOrder) //  загрузка пользователем номера заказа для расчёта;
		//r.Get("/orders", handler.GetOrder)  // получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
		//
		//r.Get("/balance", handler.Balance)            // получение текущего баланса счёта баллов лояльности пользователя;
		//r.Post("/balance/withdraw", handler.Withdraw) // запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	})

	return r
}

//func SetupRouter(handler *URLHandler) *chi.Mux {
//	r := chi.NewRouter()
//
//	r.Use(middleware.RequestID)
//	r.Use(middleware.RealIP)
//	r.Use(middleware.Recoverer)
//	r.Use(mw.LoggingMiddleware(handler.logger))
//	r.Use(mw.GzipMiddleware(handler.logger))
//
//	r.Post("/", handler.ShortenURL)
//	r.Post("/api/shorten", handler.ShortenJSON)
//	r.Post("/api/shorten/batch", handler.ShortenBatchJSON)
//	r.Get("/api/user/urls", handler.AllUserURLs)
//	r.Delete("/api/user/urls", handler.DeleteUserURLs)
//	r.Get("/{id}", handler.ExpandURL)
//	r.Get("/ping", handler.PingDB)
//
//	return r
//}
