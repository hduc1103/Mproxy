package proxies

import (
	"encoding/json"
	"log"
	"net/http"

	"project/mqtt"
	"project/handlers"
	"project/models"
)

func HTTP_MProxy() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var msg models.Message
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
		broker := mqtt.GetBrokerClient("HTTPMproxy")
		status, response := handlers.MessagesHandler(broker, payload)

		if status {
			w.WriteHeader(http.StatusOK) 
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
		w.Write([]byte(response))
	})

	log.Print("HTTP proxy server listening on localhost:1885")
	http.ListenAndServe("0.0.0.0:1885", nil)
}
