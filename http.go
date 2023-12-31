package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"fmt"
	"net/http"
	"time"
	"strings"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)
type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction   string `json:"transaction"`
	RequestId     string `json:"request_id"`
	Currency      string `json:"currency"`
	Provider      string `json:"provider"`
	Amount        int    `json:"amount"`
	PaymentDt     int64  `json:"payment_dt"`
	Bank          string `json:"bank"`
	DeliveryCost  int    `json:"delivery_cost"`
	GoodsTotal    int    `json:"goods_total"`
	CustomFee     int    `json:"custom_fee"`
}

type Item struct {
	ChrtID     int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price      int    `json:"price"`
	RID        string `json:"rid"`
	Name       string `json:"name"`
	Sale       int    `json:"sale"`
	Size       string `json:"size"`
	TotalPrice int    `json:"total_price"`
	NmID       int    `json:"nm_id"`
	Brand      string `json:"brand"`
	Status     int    `json:"status"`
}

type RequestData struct {
	OrderUID        string    `json:"order_uid"`
	TrackNumber     string    `json:"track_number"`
	Entry           string    `json:"entry"`
	Delivery        Delivery  `json:"delivery"`
	Payment         Payment   `json:"payment"`
	Items           []Item    `json:"items"`
	Locale          string    `json:"locale"`
	InternalSig     string    `json:"internal_signature"`
	CustomerID      string    `json:"customer_id"`
	DeliveryService string    `json:"delivery_service"`
	ShardKey        string    `json:"shardkey"`
	SMID            int       `json:"sm_id"`
	DateCreated     string    `json:"date_created"`
	OOFShard        string    `json:"oof_shard"`
}

func sendCache() map[string]string{

	cache := make(map[string]string)

	db, err := sql.Open("postgres","host=localhost port=5432 user=postgres password=mother545 dbname=golang_nats sslmode=disable")
	if err != nil {
		log.Print(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id,full_json FROM request_data")

	for rows.Next() {
		var value string
		var id string
		if err := rows.Scan(&id, &value); err != nil {
			log.Fatal(err)
		}
		cache[id] = value
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return cache
}


func main() {

	
	http.HandleFunc("/sendData", handleRequest)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		GetValue(w, r, sendCache())})

	log.Println("Сервер запущен на http://localhost:8080")
	sendCache()
	log.Fatal(http.ListenAndServe(":8080", nil))

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	subject := "myChannel"
	sub, err := nc.SubscribeSync(subject)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			msg, err := sub.NextMsg(time.Second)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Получено сообщение на сервере NATS: %s\n", string(msg.Data))
		}

	}()

}

func GetValue(w http.ResponseWriter, r *http.Request, cache map[string]string) {
	
	path := r.URL.Path
	parts := strings.Split(path, "/")
	
	if len(parts) > 1 && parts[1] != "" {
		number := parts[1]
		fmt.Fprintf(w, "Вы запросили информацию для числа %s\n", number)
		fmt.Fprintf(w, cache[number])
		return
	}
	
	http.Error(w, "Необходимо указать число после /", http.StatusBadRequest)
}


func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Некорректный формат JSON. Ожидается структура RequestData.", http.StatusBadRequest)
		return
	}else{

		responseJSON, err := json.Marshal(requestData)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			log.Fatal(err)
		}

		subject := "myChannel"

		sub, err := nc.SubscribeSync(subject)
		if err != nil {
			log.Fatal(err)
		}

		go func(){
			var result RequestData

			msg, err := sub.NextMsg(time.Second)
			if err != nil {
				log.Fatal(err)
			}
			if err := json.Unmarshal(msg.Data, &result); err != nil {
				log.Printf("Ошибка при разборе JSON:", err)
				return
			}
			log.Print("Данные приняты с Nats сервера")

			sendDataToDB(result, string(msg.Data))

		}()

		err = nc.Publish(subject, responseJSON)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("Данные отправлены на Nats сервер")
	}
}


func sendDataToDB(requestData RequestData, json_data string) error{
	db, err := sql.Open("postgres","host=localhost port=5432 user=postgres password=mother545 dbname=golang_nats sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	query := `
		INSERT INTO item (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

	`	

	query1 := `

		INSERT INTO delivery (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7);

	`
	query2 := `
		INSERT INTO payment (transaction, request_id,currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`

	query3 := `

		INSERT INTO request_data (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, delivery_id, payment_id, item_id, full_json) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);
	`

	items := requestData.Items


	for _, item := range items {
		_, err := db.Exec(query, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = db.Exec(query1, requestData.Delivery.Name, requestData.Delivery.Phone, requestData.Delivery.Zip, requestData.Delivery.City, requestData.Delivery.Address, requestData.Delivery.Region, requestData.Delivery.Email)

	_, err = db.Exec(query2, requestData.Payment.Transaction,requestData.Payment.RequestId, requestData.Payment.Currency, requestData.Payment.Provider, requestData.Payment.Amount, requestData.Payment.PaymentDt, requestData.Payment.Bank, requestData.Payment.DeliveryCost, requestData.Payment.GoodsTotal, requestData.Payment.CustomFee)

	
	var lastDeliveryID, lastPaymentID, lastItemID int
	row := db.QueryRow("SELECT id FROM delivery ORDER BY id DESC LIMIT 1")
	err = row.Scan(&lastDeliveryID)
	if err != nil {
		log.Fatal(err)
	}

	row = db.QueryRow("SELECT id FROM payment ORDER BY id DESC LIMIT 1")
	err = row.Scan(&lastPaymentID)
	if err != nil {
		log.Fatal(err)
	}

	row = db.QueryRow("SELECT id FROM item ORDER BY id DESC LIMIT 1")
	err = row.Scan(&lastItemID)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(query3, requestData.OrderUID, requestData.TrackNumber, requestData.Entry, requestData.Locale, requestData.InternalSig, requestData.CustomerID, requestData.DeliveryService, requestData.ShardKey, requestData.SMID, requestData.DateCreated, requestData.OOFShard, lastDeliveryID, lastPaymentID, lastItemID, json_data)


	if err != nil {
		log.Printf("Error: ", err)
		return nil
	}

	log.Println("Данные успешно отправлены в базу данных")
	return nil
}


