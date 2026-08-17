package main

import (
	"log"
	"net/http"
	"os"

	"water-quality/internal/httpapi"
	"water-quality/internal/service"
	"water-quality/internal/store"
)

func main() {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "data.json"
	}

	st, err := store.NewJSONStore(dataFile)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	svc := service.NewWaterService(st)
	handler := httpapi.NewHandler(svc)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
