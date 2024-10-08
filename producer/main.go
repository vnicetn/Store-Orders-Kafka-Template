package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Order struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	ItemName string `json:"itemname"`
	ItemID   int    `json:"itemid"`
}

func main() {
	http.HandleFunc("/order", placeOrder)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func ConnectProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(brokers, config)
}

func PushOrderToQueue(topic string, message []byte) error {
	brokers := []string{"localhost:9092"}
	// Create connection
	producer, err := ConnectProducer(brokers)
	if err != nil {
		return err
	}

	defer producer.Close()

	// Create a new message

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	// Send message
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Order is stored in topic(%s)/partition(%d)/offset(%d)\n",
		topic,
		partition,
		offset)
	return nil
}

// Place order handler

func placeOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse request body into order

	order := new(Order)
	if err := json.NewDecoder(r.Body).Decode(order); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Convert Body into bytes

	orderInBytes, err := json.Marshal(order)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Send the bytes to kafka

	err = PushOrderToQueue("store_orders", orderInBytes)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Write the data into the database

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))

	// 5. Respond back to the user
	orderId := string(order.ID)
	itemId := string(order.ItemID)

	response := map[string]interface{}{
		"success": true,
		"msg":     "Order for " + order.Name + " (ID: " + orderId + ") placed successfully with the chosen item of " + order.ItemName + " (ID: " + itemId + ")!",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
		http.Error(w, "Error placing order", http.StatusInternalServerError)
		return
	}
}
