package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/leandrorondon/br-2022-elections-data/internal/steps"
	ibgelocalidades "github.com/leandrorondon/go-ibge-localidades"
	"github.com/leandrorondon/go-ibge-localidades/api"
	_ "github.com/lib/pq"
)

const (
	dbName = "bronze"
)

type StepsService interface {
	Execute(ctx context.Context, step string, fn func(context.Context) error) error
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
	db, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var steps StepsService = steps.NewService(db)

	ctx := context.Background()
	api := ibgelocalidades.New()

	err = steps.Execute(ctx, "localidades-regioes", func(ctx context.Context) error {
		return getAndSaveRegioes(ctx, api, db)
	})
	if err != nil {
		log.Print(err)
		return
	}

	err = steps.Execute(ctx, "localidades-ufs", func(ctx context.Context) error {
		return getAndSaveUFs(ctx, api, db)
	})
	if err != nil {
		log.Print(err)
		return
	}

	err = steps.Execute(ctx, "localidades-municipios", func(ctx context.Context) error {
		return getAndSaveMunicipios(ctx, api, db)
	})
	if err != nil {
		log.Print(err)
	}
}

func getAndSaveRegioes(ctx context.Context, api *api.API, db *sqlx.DB) error {
	regioes, err := api.Regioes.Regioes(ctx)
	if err != nil {
		return err
	}

	log.Println("[localidades-regioes] Regiões obtidas:", len(regioes))

	for _, r := range regioes {
		query := `INSERT INTO regiao (id, nome, sigla) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := db.ExecContext(ctx, query, r.ID, r.Nome, r.Sigla)
		if err != nil {
			return err
		}
	}

	log.Println("[localidades-regioes] Regiões salvas.")

	return nil
}

func getAndSaveUFs(ctx context.Context, api *api.API, db *sqlx.DB) error {
	ufs, err := api.UFs.UFs(ctx)
	if err != nil {
		return err
	}

	log.Println("[localidades-ufs UFs obtidas:", len(ufs))

	for _, r := range ufs {
		query := `INSERT INTO uf (id, nome, sigla, regiao_id) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
		_, err := db.ExecContext(ctx, query, r.ID, r.Nome, r.Sigla, r.Regiao.ID)
		if err != nil {
			return err
		}
	}

	log.Println("[localidades-ufs] UFs salvas.")

	return nil
}

func getAndSaveMunicipios(ctx context.Context, api *api.API, db *sqlx.DB) error {
	municipios, err := api.Municipios.Municipios(ctx)
	if err != nil {
		return err
	}

	log.Println("[localidades-municipios] Municipios obtidos:", len(municipios))

	for _, r := range municipios {
		query := `INSERT INTO municipio (id, nome, uf_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
		_, err := db.ExecContext(ctx, query, r.ID, r.Nome, r.Microrregiao.Mesorregiao.UF.ID)
		if err != nil {
			return err
		}
	}

	log.Println("[localidades-municipios] Municipios salvos.")

	return nil
}
