package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	dbName = "bronze"
)

type result struct {
	Localidade string            `json:"localidade"`
	Result     map[string]string `json:"res"`
}

type population struct {
	ID     int      `json:"id"`
	Result []result `json:"res"`
}

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

	resp, err := http.Get("https://servicodados.ibge.gov.br/api/v1/pesquisas/indicadores/25207/resultados/xxxxxxx")
	if err != nil {
		log.Fatal(err)
	}

	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var p []population
	err = json.Unmarshal(b, &p)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Populations to insert:", len(p[0].Result))

	for _, r := range p[0].Result {
		query := `INSERT INTO populacao(localidade, populacao) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.Localidade, r.Result["2010"])
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	log.Println("Populations saved")
}
