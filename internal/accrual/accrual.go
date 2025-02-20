package accrual

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/go-resty/resty/v2"
	"github.com/theplant/luhn"
)

type Tovar struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
type Buyback struct {
	Match      string `json:"match"`
	Reward     int    `json:"reward"`
	RewardType string `json:"reward_type"`
}
type orda struct {
	Order string  `json:"order"`
	Goods []Tovar `json:"goods"`
}
type OrderStatus struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var Accrualhost = "localhost:8080"

var marks = []Buyback{
	{Match: "Acer", Reward: 20, RewardType: "pt"},
	{Match: "Bork", Reward: 10, RewardType: "%"},
	{Match: "Asus", Reward: 20, RewardType: "pt"},
	{Match: "Samsung", Reward: 25, RewardType: "%"},
	{Match: "Apple", Reward: 35, RewardType: "%"},
}

func LoadGood(num int, goodIdx int, price float64) error {
	ord := orda{Order: strconv.Itoa(Luhner(num)), Goods: []Tovar{
		{Description: "Smth " + marks[goodIdx].Match + " " + strconv.Itoa(num), Price: price}}}
	buyM, _ := json.Marshal(ord)
	fmt.Printf("%s\n", buyM)
	err := poster("/api/orders", buyM)
	return err
}

func InitAccrualForTests(logger *logging.ZapLogger) error {
	for _, r := range marks { // load to accrual good's type and buybacks
		buyM, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		err = poster("/api/goods", buyM)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	for idx := range 1000 { // затарим ордерами
		err := LoadGood(idx+1, int(rand.Int63n(5)), float64(rand.Int63n(100000)))
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	logger.Info(fmt.Sprintf("Loaded %d goods", len(marks)*999))
	return nil
}

func poster(postCMD string, wts []byte) error {
	httpc := resty.New() //
	httpc.SetBaseURL("http://" + Accrualhost)
	req := httpc.R().
		SetHeader("Content-Type", "application/json").
		SetBody(wts)
	_, err := req.
		SetDoNotParseResponse(false).
		Post(postCMD) //
	return err
}

func Luhner(numb int) int {
	// if luhn.Valid(numb) {	// если возвращать неизменённым, возникнут коллизии, типа у 2 Лун 26, и у 26 тоже 26
	// 	return numb
	// }
	return 10*numb + luhn.CalculateLuhn(numb)
}
