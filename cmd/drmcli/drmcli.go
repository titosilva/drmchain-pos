package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("Welcome to drmchain-pos CLI tool")

	log.Println("Testing connection to drmclient http server")
	if err := testConnection(); err != nil {
		log.Printf("Error testing connection: %v\n", err)
		return
	}
	log.Println("Server is up and running")

	log.Println("Exiting drmchain-pos CLI tool")
}

func testConnection() error {
	resp, err := http.Get("http://localhost:2502/check") // todo: move to a config file
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}
