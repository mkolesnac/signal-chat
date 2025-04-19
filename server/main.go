package main

import (
	"flag"
	"github.com/dgraph-io/badger/v4"
	"log"
)

func main() {
	// Parse command line flags
	host := flag.String("host", "localhost", "Host to listen on")
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Initialize database
	dbOpts := badger.DefaultOptions("./data")
	db, err := badger.Open(dbOpts)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize server
	server, err := NewServer(db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	if err := server.Start(*host, *port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
