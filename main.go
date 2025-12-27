package main

import (
	"loanBackedBitcoin/frontend"
	"log"
	"time"
)

func main() {
	restFrontend, err := frontend.NewRestFrontend(
		frontend.WithAddress("0.0.0.0"),
		frontend.WithIdleTimeout(time.Minute),
		frontend.WithPort(8088),
	)

	if err != nil {
		log.Fatal(err)
	}

	err = restFrontend.Start()
	if err != nil {
		log.Fatal(err)
	}
}
