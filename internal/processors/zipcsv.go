package processors

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/text/encoding/charmap"
)

type StepsService interface {
	Execute(ctx context.Context, step string, fn func(context.Context) error) error
}

func NewZipCsvProcessor(name, step, table, url string, db *sqlx.DB, stepsService StepsService) *ZipCsvProcessor {
	return &ZipCsvProcessor{
		name:         name,
		step:         step,
		table:        table,
		url:          url,
		db:           db,
		stepsService: stepsService,
	}
}

type ZipCsvProcessor struct {
	name            string
	step            string
	table           string
	url             string
	db              *sqlx.DB
	stepsService    StepsService
	OverrideColumns []string
}

func (p *ZipCsvProcessor) Run(ctx context.Context) error {
	return p.stepsService.Execute(ctx, p.step, p.process)
}

func (p *ZipCsvProcessor) process(ctx context.Context) error {
	log.Printf("Buscando dados de %s.", p.name)

	resp, err := http.Get(p.url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	filePath := fmt.Sprintf("%stmp%d.zip", os.TempDir(), rand.Int())
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.Remove(filePath)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Arquivo zip salvo. Analisando conteúdo.")

	r, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		if !strings.Contains(f.Name, ".csv") {
			continue
		}

		count++

		s := fmt.Sprintf("%s-%s", p.step, f.Name)
		err = p.stepsService.Execute(ctx, s, func(ctx context.Context) error {
			return p.processFile(ctx, f)
		})
		if err != nil {
			return err
		}
	}

	log.Printf("Arquivos processados: %d.", count)
	return nil
}

func (p *ZipCsvProcessor) processFile(ctx context.Context, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}

	r := charmap.ISO8859_1.NewDecoder().Reader(rc)
	parser := csv.NewReader(r)
	parser.Comma = ';'

	columns, err := parser.Read()
	if err != nil {
		return err
	}

	queryColumns := columnListToQuery(columns)
	if len(p.OverrideColumns) > 0 {
		queryColumns = columnListToQuery(p.OverrideColumns)
	}

	queryPlaceholders := buildPlaceholders(parser.FieldsPerRecord)

	return p.saveCSVToDB(ctx, parser, queryColumns, queryPlaceholders)
}

func (p *ZipCsvProcessor) saveCSVToDB(ctx context.Context, parser *csv.Reader, columns, placeholders string) error {
	count := 0
	for {
		record, err := parser.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		query := fmt.Sprintf(
			`INSERT INTO %s(%s) VALUES (%s) ON CONFLICT DO NOTHING`,
			p.table, columns, placeholders,
		)
		values := recordToValues(record)

		_, err = p.db.ExecContext(ctx, query, values...)
		if err != nil {
			return err
		}

		count++
	}

	log.Printf("Registros salvos: %d.", count)

	return nil
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
