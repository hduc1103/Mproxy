package main

import (
	"project/proxies"
)

func main() {
	go proxies.TCP_MProxy()
	go proxies.HTTP_MProxy()

	select {}
}
