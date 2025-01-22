package handlers

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"project/database"
	"project/models"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mutex sync.Mutex

func MessagesHandler(broker mqtt.Client, payload []byte) (bool, string) {
	db, err := database.ConnectDB()
	if err != nil {
		log.Printf("Error connecting to the database: %v", err)
		return false, "Error connecting to the database"
	}
	defer db.Close()

	var msg models.Message
	err = json.Unmarshal(payload, &msg)
	if err != nil {
		log.Printf("Invalid JSON format: %v", err)
		return false, "Invalid JSON format: " + err.Error()
	}

	if msg.Token == "" {
		var device models.Device
		if err := json.Unmarshal(payload, &device); err != nil {
			log.Printf("Error extracting device ID and password: %v", err)
			return false, "Please sign in first"
		}

		token, authErr := AuthenticateAndGenerateToken(db, device.DeviceID, device.Password, time.Hour*24)
		if authErr != nil {
			log.Printf("Authentication failed for device %s: %v", device.DeviceID, authErr)
			return false, "Invalid device ID or password"
		}

		log.Printf("Device %s successfully authenticated", device.DeviceID)
		return true, "Token issued: " + token
	}

	log.Printf("Received message - Token: %s, Message: %s\n", msg.Token, msg.Message)
	deviceID, signinErr := VerifyToken(msg.Token)
	if signinErr != nil {
		log.Printf("Token verification failed: %v", signinErr)
		return false, "Invalid or expired token"
	}

	mutex.Lock()
	isSpamming, err := CheckSpam(db, deviceID)
	mutex.Unlock()
	if err != nil {
		log.Printf("Error checking for spam: %v", err)
		return false, "Error checking for spam"
	}
	if isSpamming {
		log.Printf("Device %s is spamming. Blocking.\n", deviceID)
		return false, "You are sending messages too frequently. Please slow down."
	}

	curTime := time.Now()
	query := `INSERT INTO messages (message, device_id, timestamp) VALUES (?, ?, ?)`
	_, err = db.Exec(query, msg.Message, deviceID, curTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		return false, "Error saving message to database"
	}

	token := broker.Publish("test/message", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("Error forwarding message to broker: %v\n", token.Error())
		return false, "Error forwarding message to broker"
	}

	log.Print("Message successfully forwarded to broker.")
	return true, "Message successfully processed and forwarded"
}
