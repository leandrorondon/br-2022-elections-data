package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	indicadores := fmt.Sprintf("%d|%d|%d|%d|%d|%d|%d|%d", IndPopulacao, IndPopulacaoEstimada, IndDensidadeDemografica,
		IndSalarioMedio, IndTaxaEscolarizacao, IndPIBPerCapita, IndIDHM, IndMortalidadeInfantil)

	for i := 1; i < 10; i++ {
		getIndicatorsRange(db, indicadores, i)
	}

	log.Println("Indicadores salvos.")
}

func getIndicatorsRange(db *sql.DB, indicadores string, i int) {
	log.Printf("Obtendo indicadores de municípios com ID iniciando em %d.", i)
	url := fmt.Sprintf("https://servicodados.ibge.gov.br/api/v1/pesquisas/indicadores/%s/resultados/%dxxxxxx", indicadores, i)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var b []byte
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var indResp []IndicatorsResponse
	err = json.Unmarshal(b, &indResp)
	if err != nil {
		log.Fatal(err)
	}

	for _, ind := range indResp {
		indDB, ok := indicatorDB[ind.ID]
		if !ok {
			continue
		}

		table := fmt.Sprintf("indicador_%s", indDB)
		column := indDB
		log.Printf("Salvando dados de %s de %d municípios.", indicatorNames[ind.ID], len(ind.Result))

		for _, r := range ind.Result {
			result := latestResult(r.Result)
			if result == "" {
				continue
			}

			query := fmt.Sprintf(`INSERT INTO %s(localidade, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING`, table, column)
			_, err := db.Exec(query, r.Localidade, result)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func latestResult(m map[string]string) string {
	for year := 2022; year >= 2005; year-- {
		if v, ok := m[strconv.Itoa(year)]; ok && v != "-" {
			return v
		}
	}

	return ""
}
