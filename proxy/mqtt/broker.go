package mqtt

import (
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	client mqtt.Client
	once   sync.Once
)

func GetBrokerClient(mproxy string) mqtt.Client {
	once.Do(func() {
		opts := mqtt.NewClientOptions()
		opts.AddBroker("mqtt://localhost:1883")
		opts.SetClientID("mproxy-client")
		opts.SetKeepAlive(2 * time.Second)
		opts.SetPingTimeout(1 * time.Second)

		newClient := mqtt.NewClient(opts)
		token := newClient.Connect()
		token.Wait()

		if token.Error() != nil {
			log.Printf("%v failed to connect to MQTT broker: %v", mproxy, token.Error())
		}

		log.Printf("%v connected to shared MQTT broker!", mproxy)
		client = newClient
	})

	return client
}
