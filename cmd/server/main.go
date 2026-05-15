package main

import (
	"log"

	"infiour.local/dms-api-server/internal/wire"
)

func main() {
	engine, err := wire.BuildServer()
	if err != nil {
		log.Fatalf("failed to wire server: %v", err)
	}

	log.Println("Starting Gin server on :8080")
	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
