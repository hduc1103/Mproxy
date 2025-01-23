package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"project/database"
	"project/models"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		db, err := database.ConnectDB()
		if err != nil {
			log.Printf("Error connecting to the database: %v", err)
			return
		}
		defer db.Close()
	
		var received models.ServerMessage
	
		err = json.Unmarshal(msg.Payload(), &received)
		if err != nil {
			log.Printf("Error parsing message payload: %v", err)
			return
		}
	
		query := `INSERT INTO messages (message, device_id, timestamp) VALUES (?, ?, ?)`
		_, err = db.Exec(query, received.Message, received.DeviceId, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Printf("Error inserting into database: %v", err)
			return
		}
	
		fmt.Printf("Received message: %s from topic: %s\n", received.Message, msg.Topic())
	}
	

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://mqtt-broker:1883") 
	opts.SetClientID("mqtt-server")
	opts.SetDefaultPublishHandler(messageHandler)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()

	if token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	log.Println("Connected to MQTT broker")

	subscribeToken := client.Subscribe("mproxy/topic", 0, nil)
	subscribeToken.Wait()

	if subscribeToken.Error() != nil {
		log.Fatalf("Failed to subscribe to topic: %v", subscribeToken.Error())
	}

	log.Println("Subscribed to topic: test/topic")

	select {}
}

