package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var numItems int

const broker = "localhost:9093"

type Delivery struct {
	DeliveryUID string `json:"delivery_uid"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Zip         string `json:"zip"`
	City        string `json:"city"`
	Address     string `json:"address"`
	Region      string `json:"region"`
	Email       string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

func randomString() string {
	return uuid.New().String()
}

func randomPhone() string {
	return fmt.Sprintf("+%d%d", rand.Intn(99), rand.Intn(1000000000))
}

func generateRandomItems(trackNumber string, num int) []Item {
	items := make([]Item, num)
	for i := 0; i < num; i++ {
		items[i] = Item{
			ChrtID:      rand.Intn(10000000),
			TrackNumber: trackNumber,
			Price:       rand.Intn(1000),
			RID:         randomString(),
			Name:        fmt.Sprintf("Test Item %d", i+1),
			Sale:        rand.Intn(90),
			Size:        "0",
			TotalPrice:  rand.Intn(1000),
			NmID:        rand.Intn(1000000),
			Brand:       "Test Brand",
			Status:      rand.Intn(400),
		}
	}
	return items
}

func generateRandomOrder() Order {
	skipFields := rand.Float32() < 0.2

	trackNumber := fmt.Sprintf("TRACK%d", rand.Intn(1000000))

	order := Order{
		OrderUID:    randomString(),
		TrackNumber: trackNumber,
		Entry:       "WBIL",
		Delivery: Delivery{
			DeliveryUID: randomString(),
			Name:        "Test Testov",
			Phone:       randomPhone(),
			Zip:         fmt.Sprintf("%d", rand.Intn(999999)),
			City:        "Test City",
			Address:     "Test Address",
			Region:      "Test Region",
			Email:       fmt.Sprintf("test%d@test.com", rand.Intn(1000)),
		},
		Payment: Payment{
			Transaction:  randomString(),
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(10000),
			PaymentDt:    int(time.Now().Unix()),
			Bank:         "alpha",
			DeliveryCost: rand.Intn(2000),
			GoodsTotal:   rand.Intn(5000),
		},
		Items:           generateRandomItems(trackNumber, numItems),
		Locale:          "en",
		CustomerID:      fmt.Sprintf("customer%d", rand.Intn(1000)),
		DeliveryService: "meest",
		Shardkey:        fmt.Sprintf("%d", rand.Intn(10)),
		SmID:            rand.Intn(100),
		DateCreated:     time.Now(),
		OofShard:        fmt.Sprintf("%d", rand.Intn(10)),
	}

	if skipFields {
		if rand.Float32() < 0.3 {
			order.InternalSignature = ""
		}
		if rand.Float32() < 0.3 {
			order.Payment.RequestID = ""
		}
		if rand.Float32() < 0.3 {
			order.Payment.CustomFee = 0
		}
	}

	return order
}

func main() {
	var numOrders int
	fmt.Println("Enter num of orders:")
	fmt.Scan(&numOrders)

	fmt.Println("Enter num of items:")
	fmt.Scan(&numItems)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   "orders",
	})
	defer writer.Close()

	for i := 0; i < numOrders; i++ {
		order := generateRandomOrder()

		value, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order %d: %v", i, err)
			continue
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(order.OrderUID),
				Value: value,
			},
		)

		if err != nil {
			log.Printf("Failed to write message %d: %v", i, err)
			continue
		}

		log.Printf("Successfully sent order %d: %s\n", i, order.OrderUID)
		time.Sleep(time.Second)
	}
}
