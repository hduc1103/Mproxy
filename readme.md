# MProxy Server with MQTT 

This project implements two mproxy server which are **TCP Proxy Server** and **HTTP Server**. These two process incoming messages, check for spamming based on device IDs, store messages in a MySQL database, and forward valid messages to the MQTT broker.

## Requirement
- **Go Programming Language** (1.23.4)
- **MySQL Database**
- **MQTT Broker** (Mqtt Explorer)
- **Image Vernemq** (Run on Docker)
- Required Go modules:
  - `github.com/go-sql-driver/mysql`
  - `github.com/eclipse/paho.mqtt.golang`

## Features
- **TCP MProxy Server:** Listens for incoming client connections on port 1884.
- **HTTP Server:** Listens and Accept HTTP with method POST requests on `localhost:1885/message`.
- **Spam Checking:** Block deviceId in case of sending messages too frequently.
- **Database Storage:** Stores incoming messages in a MySQL database.
- **MQTT Forwarding:** Forwards valid messages to the MQTT broker.

## Installation
1. Clone the repository:
   ```bash
   git clone
   cd 
   ```

2. Install the required Go modules:
   ```bash
   go get github.com/go-sql-driver/mysql
   go get github.com/eclipse/paho.mqtt.golang
   go mod tidy
   ```

3. Set up a MySQL database and create a table for storing messages:
   ```sql
   CREATE DATABASE proxy;
   USE proxy;
   CREATE TABLE messages (
       id INT AUTO_INCREMENT PRIMARY KEY,
       message TEXT NOT NULL,
       device_id VARCHAR(255) NOT NULL,
       timestamp DATETIME NOT NULL
   );
   ```

4. Install Image Vernemq on Docker:
    ```bash
    docker pull vernemq/vernemq
    docker run -d --name vernemq-broker -p 1883:1883 -p 8080:8080 -e "DOCKER_VERNEMQ_ACCEPT_EULA=yes" -e "DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on" vernemq/vernemq
    docker start vernemq-broker
    ```
## Configuration
- The database connection string is hardcoded in the `handleClientMessages` function:
  ```go
  "root:<your_mysql_password>@tcp(127.0.0.1:3306)/proxy?parseTime=true&loc=Local"
  ```
- The MQTT broker address is configured in the `getBrokerClient` function:
  ```go
  opts.AddBroker("mqtt://localhost:1883")
  ```

## Running the Project
1. Start the proxy server:
   ```bash
   docker start vernemq-broker
   go run main.go
   ```

2. The server listens on `localhost:1884` for incoming TCP connections.
3. The HTTP server listens on `localhost:1885/message` for POST requests.
4. The broker listens on `localhost:1883` for receiving messages.

## How It Works
1. **Client Connection:**
   - Clients connect to the proxy server via TCP on `localhost:1884`.

2. **Message Handling:**
   - The server receives messages in JSON format containing a `device_id` and `message`.

3. **Spam Check:**
   - Checks the database for the last message timestamp for the given `device_id`.
   - Within 5 seconds, block any message from the device with that device_id

4. **Database Storage:**
   - Stores the message in the MySQL database.

5. **MQTT Forwarding:**
   - Forwards the valid message to the MQTT broker on the topic `test/message`.

## Example Message Format
```json
{
  "device_id": "device1",
  "message": "Hello, World!"
}
```

## Testing
1. **For TCP:**
- Simple Go code for testing:
   ```
package main
import (
	"encoding/json"
	"fmt"
	"net"
)

func main() {
	host := "127.0.0.1"
	port := 1884
	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	message := map[string]string{
		"device_id": "device123",
		"message":   "Hello1",
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	_, err = conn.Write(jsonMessage)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	fmt.Println("Message sent successfully")
}```

2. **For HTTP: Using Postman**
- Method: POST
- url: http://localhost:1885/message
- Header: 
    - Key: Content-Type
    - Value: application/json
- Body:
    ```json
    {
  "device_id": "device122",
  "message": "Hello!"
    }
    ```
- **Response:**
  - `200 OK`: Message processed successfully.
  - `400 Bad Request`: Invalid JSON format.
  - `404 Not Found`: The server might not have been run properly.
  - `405 Method Not Allowed`: Request method other than POST.

## Logging
The server logs events such as:
- Incoming connections and messages.
- Spam detection.
- Database operations.
- MQTT publishing results.
