package main

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nexpictora-pvt-ltd/cnx-backend/api"
	db "github.com/nexpictora-pvt-ltd/cnx-backend/db/sqlc"
	"github.com/nexpictora-pvt-ltd/cnx-backend/messaging"
	"github.com/nexpictora-pvt-ltd/cnx-backend/util"
)

// @title           Cnx-Backend API
// @version         1.0
// @description     This is a backend API for CTT_Back the Applicaation with integrated CRM + Ordering System.
// @termsOfService  http://swagger.io/terms/

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configuration:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start the server:", err)
	}

	consumer, err := messaging.NewConsumer("amqp://guest:guest@localhost:5672/", "new-orders")
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Start consuming messages
	newOrderMessages, err := consumer.ConsumeMessages("new-orders")
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	// Handle new order messages
	go func() {
		for msg := range newOrderMessages {
			// Handle the new order message here
			// Parse the message if necessary
			// Call your updateOrder function
			// ...
			log.Printf("Received new order message: %s", msg.Body)
		}
	}()

	var ctx *gin.Context
	// Handle new order messages
	go func() {
		for msg := range newOrderMessages {
			// Handle the new order message here
			// Parse the message if necessary
			// Call updateUserOrder function to update the user's order count
			var orderMessage api.UpdateUserOrderRequest
			err := json.Unmarshal(msg.Body, &orderMessage)
			if err != nil {
				log.Printf("Error parsing order message: %v", err)
				continue
			}

			// Call updateUserOrder function to update the user's order count
			server.UpdateUserOrder(ctx)
		}
	}()

}
