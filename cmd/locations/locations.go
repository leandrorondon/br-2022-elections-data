package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	ibgelocalidades "github.com/leandrorondon/go-ibge-localidades"
	"github.com/leandrorondon/go-ibge-localidades/api"
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

	api := ibgelocalidades.New()

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
		panic(err)
	}
	defer db.Close()

	getAndSaveRegioes(api, db)
	getAndSaveUFs(api, db)
	getAndSaveMunicipios(api, db)
}

func getAndSaveRegioes(api *api.API, db *sql.DB) {
	regioes, err := api.Regioes.Regioes(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Regioes to insert:", len(regioes))

	for _, r := range regioes {
		query := `INSERT INTO regioes (id, nome, sigla) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.ID, r.Nome, r.Sigla)
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	log.Println("Regioes saved.")
}

func getAndSaveUFs(api *api.API, db *sql.DB) {
	ufs, err := api.UFs.UFs(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("UFs to insert:", len(ufs))

	for _, r := range ufs {
		query := `INSERT INTO ufs (id, nome, sigla, regiao_id) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.ID, r.Nome, r.Sigla, r.Regiao.ID)
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	log.Println("UFs saved.")
}

func getAndSaveMunicipios(api *api.API, db *sql.DB) {
	municipios, err := api.Municipios.Municipios(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Municipios to insert:", len(municipios))

	for _, r := range municipios {
		query := `INSERT INTO municipios (id, nome, uf_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.ID, r.Nome, r.Microrregiao.Mesorregiao.UF.ID)
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	log.Println("Municipios saved.")
}
