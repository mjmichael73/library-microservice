package main

import (
	"log"

	"github.com/mjmichael73/library-microservice/loanservice/internal/database"
	"github.com/mjmichael73/library-microservice/loanservice/internal/server"
)

func main() {
	db, err := database.NewDatabaseClient()
	if err != nil {
		log.Fatalf("failed to initialize Database client: %s", err)
	}
	srv := server.NewEchoServer(db)
	if err := srv.Start(); err != nil {
		log.Fatal(err.Error())
	}
}
