package proxies

import (
	"log"
	"net"

	"project/mqtt"
	"project/handlers"
)

func TCP_MProxy() {
	listener, err := net.Listen("tcp", "localhost:1884")
	if err != nil {
		log.Printf("Failed to start proxy server: %v", err)
		return
	}
	defer listener.Close()
	log.Print("TCP Proxy server listening on localhost:1884")

	for {
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
			broker := mqtt.GetBrokerClient("TCPMproxy")
			_, response := handlers.MessagesHandler(broker, buf[:n]) 

			_, err = conn.Write([]byte(response))
			if err != nil {
				log.Printf("Error sending response: %v", err)
				return
			}

			log.Printf("Response sent to client: %s\n", response)
		}(conn)
	}
}
