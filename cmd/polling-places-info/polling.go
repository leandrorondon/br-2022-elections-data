package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/leandrorondon/br-2022-elections-data/internal/processors"
	"github.com/leandrorondon/br-2022-elections-data/internal/steps"
	_ "github.com/lib/pq"
)

const (
	dbName = "bronze"

	modelosUrnaTable = "modelourna_numerointerno"
	modelosUrnaURL   = "https://cdn.tse.jus.br/estatistica/sead/odsele/modelo_urna/modelourna_numerointerno.zip"
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

	modelosUrna := processors.NewZipCsvProcessor("Modelos de Urna x Número Interno", "modelosurna", modelosUrnaTable, modelosUrnaURL, db, stepsService)
	modelosUrna.OverrideColumns = []string{"ds_modelo_urna", "nr_faixa_inicial", "nr_faixa_final"}
	err = modelosUrna.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	electoralZones := processors.NewZonesProcessor(db, stepsService)
	err = electoralZones.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	electoralSections := processors.NewSectionsProcessor(db, stepsService)
	err = electoralSections.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
