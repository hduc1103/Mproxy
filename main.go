package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	DeviceID string `json:"device_id"`
	Message  string `json:"message"`
}

var (
	accessMutex  sync.Mutex
	once         sync.Once
	sharedClient mqtt.Client
)

func getBrokerClient(mproxy string) mqtt.Client {
	once.Do(func() {
		opts := mqtt.NewClientOptions()
		opts.AddBroker("mqtt://localhost:1883")
		opts.SetClientID("shared-broker-forwarder")
		opts.SetKeepAlive(2 * time.Second)
		opts.SetPingTimeout(1 * time.Second)

		client := mqtt.NewClient(opts)
		token := client.Connect()
		token.Wait()

		if token.Error() != nil {
			log.Printf("%v failed to connect to MQTT broker: %v", mproxy, token.Error())
		}

		log.Printf("%v connected to shared MQTT broker!", mproxy)
		sharedClient = client
	})

	return sharedClient
}

func checkSpam(db *sql.DB, deviceID string) (bool, error) {
	query := `SELECT timestamp FROM messages WHERE device_id = ? ORDER BY timestamp DESC LIMIT 5`

	rows, err := db.Query(query, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No previous messages for device %s. Not spamming.\n", deviceID)
			return false, nil
		}
		log.Printf("Error querying timestamp for device %s: %v\n", deviceID, err)
		return false, err
	}
	defer rows.Close()

	var timestamps []time.Time
	for rows.Next() {
		var timestamp time.Time
		if err := rows.Scan(&timestamp); err != nil {
			log.Printf("Failed to scan row: %v", err)
		}

		timestamps = append(timestamps, timestamp)
	}

	if len(timestamps) < 2 {
		log.Printf("Device %s is not spamming. Average gap: %v\n", deviceID, 0)
		return false, nil
	}

	var totalGap time.Duration
	for i := 0; i < len(timestamps)-1; i++ {
		gap := timestamps[i].Sub(timestamps[i+1]) 
		totalGap += gap
	}

	averageGap := totalGap / time.Duration(len(timestamps)-1)

	threshold := 5 * time.Second

	isSpamming := averageGap < threshold

	if isSpamming {
		log.Printf("Device %s is spamming. Average gap: %v\n", deviceID, averageGap)
	} else {
		log.Printf("Device %s is not spamming. Average gap: %v\n", deviceID, averageGap)
	}

	return isSpamming, nil
}

func handleClientMessages(brokerClient mqtt.Client, payload []byte) {
	db, err := sql.Open("mysql", "root:123456789@tcp(127.0.0.1:3306)/proxy?parseTime=true&loc=Local")
	if err != nil {
		log.Printf("Error connecting to the database: %v", err)
		return
	}
	defer db.Close()

	var msg Message
	err = json.Unmarshal(payload, &msg)
	if err != nil {
		log.Printf("Invalid JSON format: %v", err)
		return
	}
	log.Printf("Received message - DeviceID: %s, Message: %s\n", msg.DeviceID, msg.Message)

	accessMutex.Lock()
	isSpamming, err := checkSpam(db, msg.DeviceID)
	accessMutex.Unlock()
	if err != nil {
		log.Printf("Error checking for spam: %v", err)
		return
	}
	if isSpamming {
		log.Printf("Device %s is spamming. Blocking.\n", msg.DeviceID)
		return
	}

	currentTime := time.Now()
	query := `INSERT INTO messages (message, device_id, timestamp) VALUES (?, ?, ?)`
	_, err = db.Exec(query, msg.Message, msg.DeviceID, currentTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		return
	}

	token := brokerClient.Publish("test/message", 0, false, payload)
	token.Wait()

	if token.Error() != nil {
		log.Printf("Error forwarding message to broker: %v\n", token.Error())
	} else {
		log.Print("Message successfully forwarded to broker.")
	}
}

func startTCPMproxy() {
	listener, err := net.Listen("tcp", "localhost:1884")
	if err != nil {
		log.Printf("Failed to start proxy server: %v", err)
	}
	defer listener.Close()
	log.Print("TCPProxy server listening on localhost:1884")

	brokerClient := getBrokerClient("TCPMproxy") 
	for {
		log.Print("Waiting for client connection...")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		log.Print("Client connected!")
		go func(conn net.Conn) {
			defer conn.Close()

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("Error reading data: %v", err)
				return
			}

			log.Printf("Received data: %s\n", string(buf[:n]))
			handleClientMessages(brokerClient, buf[:n]) 
		}(conn)
	}
}

func startHTTPMproxy() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Handler invoked for /message")

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var msg Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		payload, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, "Failed to process message", http.StatusInternalServerError)
			return
		}

		brokerClient := getBrokerClient("HTTPMproxy")
		handleClientMessages(brokerClient, payload)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message processed successfully"))
	})

	log.Print("HTTP server listening on localhost:1885")
	log.Print(http.ListenAndServe("localhost:1885", nil))
}


func main() {
	go startTCPMproxy()
	go startHTTPMproxy()

	select {}
}
