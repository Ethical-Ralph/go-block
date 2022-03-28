package main

import (
	"flag"
	"log"

	blockchainserver "github.com/Ethical-Ralph/go-block/blockchain_server"
)

func init() {
	log.SetPrefix("BlockChain: ")
}


func main(){
	port := flag.Uint("port", 8080, "port to listen on")
	flag.Parse()

	app := blockchainserver.NewBlockchainServer((uint16(*port)))

	app.Start()
}
