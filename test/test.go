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
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2VfaWQiOiJkZXZpY2VfMTIzIiwiZXhwIjoxNzM3NzMyNjI2fQ.r4nDLjV5iNBQ5ypeRZ18GC45FdEnXHPnoOufBPy2dTA",
		"message":   "12345678",
	}
	// message := map[string]string{
	// 	"device_id": "device_123",
	// 	"password":  "12345678",
	// }
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

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Response from server:", string(buf[:n]))
}
