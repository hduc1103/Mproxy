package main

import (
	"project/proxies"
)

func main() {
	go proxies.TCPMProxy()
	go proxies.HTTPMProxy()

	select {}
}
