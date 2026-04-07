package main

import (
	"fhir_facade/config"
	"fhir_facade/db"
	"fhir_facade/handler"
	"fhir_facade/store"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("FHIR Facade starting in [%s] mode on port %s\n", cfg.Env, cfg.Port)

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	fmt.Println("Connected to database successfully")

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("Database migrations applied")

	base := "http://localhost:" + cfg.Port
	patientStore := store.NewPatientStore(database)
	patientHandler := handler.NewPatientHandler(patientStore, base)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /Patient", patientHandler.Create)
	mux.HandleFunc("GET /Patient/{id}", patientHandler.Read)

	log.Printf("Listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
