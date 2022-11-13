package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/leandrorondon/br-2022-elections-data/cmd/polling-places-info/processor"
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
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	modelosUrna := processor.New("Modelos de Urna x Número Interno", modelosUrnaTable, modelosUrnaURL, db)
	modelosUrna.OverrideColumns = []string{"ds_modelo_urna", "nr_faixa_inicial", "nr_faixa_final"}
	modelosUrna.Run()

}
