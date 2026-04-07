package main

import (
	"fhir_facade/config"
	"fhir_facade/db"
	"fmt"
	"log"
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


}
