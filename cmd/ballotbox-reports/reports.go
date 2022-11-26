package main

import (
	"context"
	"fmt"
	"github.com/leandrorondon/br-2022-elections-data/internal/processors/ballotboxreports"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/leandrorondon/br-2022-elections-data/internal/steps"
	_ "github.com/lib/pq"
)

const (
	dbName = "bronze"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOSTNAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbName,
	)
	db, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stepsService := steps.NewService(db)

	reports := ballotboxreports.New(db, stepsService)
	err = reports.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
