package main

import (
	"log"

	"github.com/mjmichael73/library-microservice/apigatewayservice/internal/server"
)

func main() {
	srv := server.NewEchoServer()
	if err := srv.Start(); err != nil {
		log.Fatal(err.Error())
	}
}
