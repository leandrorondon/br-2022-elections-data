package processor

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

func New(name, table, url string, db *sql.DB) *Processor {
	return &Processor{
		name:  name,
		table: table,
		url:   url,
		db:    db,
	}
}

type Processor struct {
	name            string
	table           string
	url             string
	db              *sql.DB
	OverrideColumns []string
}

func (p *Processor) Run() {
	log.Printf("Buscando dados de %s.", p.name)

	resp, err := http.Get(p.url)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("status code:", resp.StatusCode)
	}

	filePath := fmt.Sprintf("%stmp%d.zip", os.TempDir(), rand.Int())
	out, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	defer os.Remove(filePath)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Arquivo zip salvo. Analisando conteúdo.")

	r, err := zip.OpenReader(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		if !strings.Contains(f.Name, ".csv") {
			continue
		}

		count++
		log.Printf("Processando arquivo %s.", f.Name)

		rc, err := f.Open()
		if err != nil {
			log.Println("err:", err)
			continue
		}

		parser := csv.NewReader(rc)
		parser.Comma = ';'

		columns, err := parser.Read()
		if err != nil {
			log.Fatal(err)
		}

		queryColumns := columnListToQuery(columns)
		if len(p.OverrideColumns) > 0 {
			queryColumns = columnListToQuery(p.OverrideColumns)
		}

		queryPlaceholders := buildPlaceholders(parser.FieldsPerRecord)

		saveCSVToDB(parser, p.db, p.table, queryColumns, queryPlaceholders)
	}

	log.Printf("Arquivos processados: %d.", count)
}

func saveCSVToDB(parser *csv.Reader, db *sql.DB, table, columns, placeholders string) {
	count := 0
	for {
		record, err := parser.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		query := fmt.Sprintf(
			`INSERT INTO %s(%s) VALUES (%s) ON CONFLICT DO NOTHING`,
			table, columns, placeholders,
		)
		values := recordToValues(record)

		_, err = db.Exec(query, values...)
		if err != nil {
			log.Fatal(err)
		}

		count++
	}

	log.Printf("Registros salvos: %d.", count)
}

func columnListToQuery(columns []string) string {
	var s string
	for _, c := range columns {
		s = s + "," + strings.ToLower(c)
	}
	return s[1:]
}

func buildPlaceholders(n int) string {
	var s string
	for i := 1; i <= n; i++ {
		s = s + fmt.Sprintf(",$%d", i)
	}
	return s[1:]
}

func recordToValues(record []string) []any {
	var a []any
	for _, r := range record {
		a = append(a, r)
	}
	return a
}
