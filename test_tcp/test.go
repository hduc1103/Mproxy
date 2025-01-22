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
		"device_id": "device_new",
		"message":   "Hello tcp",
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
}
