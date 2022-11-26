package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/leandrorondon/br-2022-elections-data/internal/processors/zipcsv"
	"github.com/leandrorondon/br-2022-elections-data/internal/steps"
	_ "github.com/lib/pq"
)

const (
	dbName = "bronze"

	sectionResultTable = "votacao_secao"
	sectionResultURL   = "https://cdn.tse.jus.br/estatistica/sead/odsele/votacao_secao/votacao_secao_2022_BR.zip"
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

	config := zipcsv.Config{
		Name:  "Votação por Seção",
		Step:  "votacao-secao",
		Table: sectionResultTable,
		URL:   sectionResultURL,
	}
	err = zipcsv.New(db, stepsService, config, zipcsv.WithKeepDownload(true)).Run(context.Background())
	if err != nil {
		log.Print(err)
	}
}
