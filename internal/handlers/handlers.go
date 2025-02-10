package handlers

import (
	"github.com/go-chi/chi/v5"
)

type URLHandler struct {
	baseURL  string
	storage  Storage
	secret   string
	ToDelete chan models.BatchDeleteRequest
	logger   *logging.ZapLogger
}

func SetupRouter(handler *URLHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.Register)      // регистрация пользователя;
		r.Post("/login", handler.Login)            // аутентификация пользователя;
		r.Get("/withdrawals", handler.Withdrawals) //  получение информации о выводе средств с накопительного счёта пользователем.

		r.Post("/orders", handler.AddOrder) //  загрузка пользователем номера заказа для расчёта;
		r.Get("/orders", handler.GetOrder)  // получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;

		r.Get("/balance", handler.Balance)            // получение текущего баланса счёта баллов лояльности пользователя;
		r.Post("/balance/withdraw", handler.Withdraw) // запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	})

	return r
}
