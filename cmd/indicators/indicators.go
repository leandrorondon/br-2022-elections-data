package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/leandrorondon/br-2022-elections-data/internal/steps"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

const (
	dbName = "bronze"
)

const (
	IndPopulacao            = 25207
	IndPopulacaoEstimada    = 29171
	IndDensidadeDemografica = 29168
	IndSalarioMedio         = 29765
	IndTaxaEscolarizacao    = 60045
	IndPIBPerCapita         = 47001
	IndIDHM                 = 30255
	IndMortalidadeInfantil  = 30279
)

var latestResultPerIndicator = map[int]string{
	IndPopulacao:            "2010",
	IndPopulacaoEstimada:    "2021",
	IndDensidadeDemografica: "2010",
	IndSalarioMedio:         "2020",
	IndTaxaEscolarizacao:    "2010",
	IndPIBPerCapita:         "2019",
	IndIDHM:                 "2010",
	IndMortalidadeInfantil:  "2020",
}

var indicatorNames = map[int]string{
	IndPopulacao:            "População (2010)",
	IndPopulacaoEstimada:    "População Estimada (2021)",
	IndDensidadeDemografica: "Densidade Demográfica",
	IndSalarioMedio:         "Salário Médio Mensal",
	IndTaxaEscolarizacao:    "Taxa de Escolarização",
	IndPIBPerCapita:         "PIB Per Capita",
	IndIDHM:                 "IDHM",
	IndMortalidadeInfantil:  "Taxa de Mortalidade Infantil",
}

var indicatorDB = map[int]string{
	IndPopulacao:            "populacao",
	IndPopulacaoEstimada:    "populacao_estimada",
	IndDensidadeDemografica: "densidade_demografica",
	IndSalarioMedio:         "salario_medio",
	IndTaxaEscolarizacao:    "taxa_escolarizacao",
	IndPIBPerCapita:         "pib_per_capita",
	IndIDHM:                 "idhm",
	IndMortalidadeInfantil:  "taxa_mortalidade_infantil",
}

type IndicatorResult struct {
	Localidade string            `json:"localidade"`
	Result     map[string]string `json:"res"`
}

type IndicatorsResponse struct {
	ID     int               `json:"id"`
	Result []IndicatorResult `json:"res"`
}

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

	indicadores := fmt.Sprintf("%d|%d|%d|%d|%d|%d|%d|%d", IndPopulacao, IndPopulacaoEstimada, IndDensidadeDemografica,
		IndSalarioMedio, IndTaxaEscolarizacao, IndPIBPerCapita, IndIDHM, IndMortalidadeInfantil)

	var steps StepsService = steps.NewService(db)
	g, gctx := errgroup.WithContext(context.Background())

	for i := 1; i < 10; i++ {
		n := i

		s := fmt.Sprintf("indicadores-%d", n)
		g.Go(func() error {
			return steps.Execute(gctx, s, func(ct context.Context) error {
				return getIndicatorsRange(ct, db, indicadores, s, n)
			})
		})
	}

	if g.Wait() != nil {
		log.Fatal(err)
	}
}

func getIndicatorsRange(ctx context.Context, db *sqlx.DB, indicadores, s string, i int) error {
	log.Printf("[%s] Obtendo indicadores de municípios com ID iniciando em %d.", s, i)
	url := fmt.Sprintf("https://servicodados.ibge.gov.br/api/v1/pesquisas/indicadores/%s/resultados/%dxxxxxx", indicadores, i)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var indResp []IndicatorsResponse
	err = json.Unmarshal(b, &indResp)
	if err != nil {
		return err
	}

	for _, ind := range indResp {
		indDB, ok := indicatorDB[ind.ID]
		if !ok {
			continue
		}

		table := fmt.Sprintf("indicador_%s", indDB)
		column := indDB
		log.Printf("[%s] Salvando dados de %s de %d municípios.", s, indicatorNames[ind.ID], len(ind.Result))

		for _, r := range ind.Result {
			result := latestResult(r.Result)
			if result == "" {
				continue
			}

			query := fmt.Sprintf(`INSERT INTO %s(localidade, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING`, table, column)
			_, err := db.ExecContext(ctx, query, r.Localidade, result)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func latestResult(m map[string]string) string {
	for year := 2022; year >= 2005; year-- {
		if v, ok := m[strconv.Itoa(year)]; ok && v != "-" {
			return v
		}
	}

	return ""
}
