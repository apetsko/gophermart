package handlers

import (
	"time"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type URLHandler struct {
	storage *postgres.Storage
	secret  string
	//ToDelete chan models.BatchDeleteRequest
	logger *logging.ZapLogger
}

func New(s *postgres.Storage, l *logging.ZapLogger) *URLHandler {
	return &URLHandler{
		storage: s,
		logger:  l,
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
		r.Post("/register", RegisterHandler(handler)) // регистрация пользователя;
		r.Post("/login", LoginHandler(handler))       // аутентификация пользователя;
		// TODO@@@@@@@@@@
		// TODO@@@@@@@@@@
		// TODO@@@@@@@@@@
		// TODO@@@@@@@@@@
		r.Post("/orders", AddOrderHandler(handler))  //  загрузка пользователем номера заказа для расчёта;
		r.Get("/orders", ListOrdersHandler(handler)) // получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
		//
		//r.Get("/balance", handler.Balance)            // получение текущего баланса счёта баллов лояльности пользователя;
		//r.Post("/balance/withdraw", handler.Withdraw) // запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
		//r.Get("/withdrawals", handler.Withdrawals) //  получение информации о выводе средств с накопительного счёта пользователем.
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
