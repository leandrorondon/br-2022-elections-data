package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	ibgelocalidades "github.com/leandrorondon/go-ibge-localidades"
	"github.com/leandrorondon/go-ibge-localidades/api"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "eleicoes"
)

func main() {
	api := ibgelocalidades.New()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
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

	fmt.Printf("Regioes to insert: %d ... ", len(regioes))

	for _, r := range regioes {
		query := `INSERT INTO regioes (id, nome, sigla) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.ID, r.Nome, r.Sigla)
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	fmt.Printf("OK!\n")
}

func getAndSaveUFs(api *api.API, db *sql.DB) {
	ufs, err := api.UFs.UFs(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("UFs to insert: %d ... ", len(ufs))

	for _, r := range ufs {
		query := `INSERT INTO ufs (id, nome, sigla, regiao_id) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.ID, r.Nome, r.Sigla, r.Regiao.ID)
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	fmt.Printf("OK!\n")
}

func getAndSaveMunicipios(api *api.API, db *sql.DB) {
	municipios, err := api.Municipios.Municipios(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Municipios to insert: %d ... ", len(municipios))

	for _, r := range municipios {
		query := `INSERT INTO municipios (id, nome, uf_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := db.Exec(query, r.ID, r.Nome, r.Microrregiao.Mesorregiao.UF.ID)
		if err != nil {
			fmt.Printf("Failed!\n")
			log.Fatal(err)
		}
	}

	fmt.Printf("OK!\n")
}
