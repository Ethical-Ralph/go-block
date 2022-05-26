package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Wallet Server: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP port for wallet server")
	gateway := flag.String("gateway", "http://localhost:9000", "blockchain Gateway")

	app := NewWalletServer(uint16(*port), *gateway)

	done := make(chan bool)
	go app.Run()
	log.Print("Wallet UI running at: ", *port)
	<-done
}
