package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	//fmt.Print("%+v\n", store)
	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}

	server := NewAPIServer(":"+port, store)
	server.Run()

	store.db.Close()
}
