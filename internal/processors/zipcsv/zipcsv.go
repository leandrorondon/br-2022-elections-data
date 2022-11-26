package zipcsv

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/leandrorondon/br-2022-elections-data/internal/httpwithretry"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/encoding/charmap"
)

const insertBatch = 500

type StepsService interface {
	Execute(ctx context.Context, step string, fn func(context.Context) error) error
}

type Config struct {
	Name  string
	Step  string
	Table string
	URL   string
}

type Option func(*Processor)

func WithColumns(columns []string) Option {
	return func(p *Processor) {
		p.overrideColumns = columns
	}
}

func WithKeepDownload(keep bool) Option {
	return func(p *Processor) {
		p.keepDownload = keep
	}
}

func New(db *sqlx.DB, stepsService StepsService, config Config, opts ...Option) *Processor {
	p := &Processor{
		name:         config.Name,
		step:         config.Step,
		table:        config.Table,
		url:          config.URL,
		db:           db,
		stepsService: stepsService,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type Processor struct {
	name            string
	step            string
	table           string
	url             string
	db              *sqlx.DB
	stepsService    StepsService
	overrideColumns []string
	keepDownload    bool
}

func (p *Processor) Run(ctx context.Context) error {
	return p.stepsService.Execute(ctx, p.step, p.process)
}

func (p *Processor) process(ctx context.Context) error {
	log.Printf("[%s] Buscando dados de %s.", p.step, p.name)

	fileName := path.Base(p.url)
	dir := fmt.Sprintf("%s/eleicoes", os.TempDir())
	os.Mkdir(dir, 0644)
	filePath := fmt.Sprintf("%s/%s", dir, fileName)

	r, err := zip.OpenReader(filePath)
	if err == nil {
		log.Printf("[%s] Arquivo zip já existe. Analisando conteúdo.", p.step)
		defer r.Close()
		return p.processZip(ctx, r)
	}

	log.Printf("[%s] Baixando %s.", p.step, fileName)

	resp, err := httpwithretry.Get(p.url, 2)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer out.Close()

	if !p.keepDownload {
		defer os.Remove(filePath)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("[%s] Arquivo zip salvo. Analisando conteúdo.", p.step)

	r, err = zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	return p.processZip(ctx, r)
}

func (p *Processor) processZip(ctx context.Context, r *zip.ReadCloser) error {
	count := 0
	for _, f := range r.File {
		if !strings.Contains(f.Name, ".csv") {
			continue
		}

		count++

		s := fmt.Sprintf("%s-%s", p.step, f.Name)
		err := p.stepsService.Execute(ctx, s, func(ctx context.Context) error {
			return p.processFile(ctx, f, s)
		})
		if err != nil {
			return err
		}
	}

	log.Printf("[%s] Arquivos processados: %d.", p.step, count)
	return nil
}

func (p *Processor) processFile(ctx context.Context, f *zip.File, s string) error {
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
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
		if len(p.overrideColumns) > 0 {
			queryColumns = columnListToQuery(p.overrideColumns)
		}

		placeholders := buildPlaceholders(parser.FieldsPerRecord, insertBatch)
		return p.saveCSVToDB(gctx, parser, queryColumns, placeholders, s)
	})

	return g.Wait()
}

func (p *Processor) saveCSVToDB(ctx context.Context, parser *csv.Reader, columns, placeholders, s string) error {
	count := 0
	insertCount := 0
	values := make([]any, 0, insertBatch)
	for {
		record, err := parser.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		count++
		insertCount++

		v := recordToValues(record)
		values = append(values, v...)

		if insertCount%insertBatch == 0 {
			insertCount = 0

			query := fmt.Sprintf(
				`INSERT INTO %s(%s) VALUES %s`,
				p.table, columns, placeholders,
			)

			_, err = p.db.ExecContext(ctx, query, values...)
			if err != nil {
				return err
			}

			values = nil
		}
	}

	// Save remaining records
	if insertCount > 0 {
		placeholders = buildPlaceholders(parser.FieldsPerRecord, insertCount)
		query := fmt.Sprintf(
			`INSERT INTO %s(%s) VALUES %s`,
			p.table, columns, placeholders,
		)

		_, err := p.db.ExecContext(ctx, query, values...)
		if err != nil {
			return err
		}
	}

	log.Printf("[%s] Registros salvos: %d.", s, count)

	return nil
}

func columnListToQuery(columns []string) string {
	var s string
	for _, c := range columns {
		s = s + "," + strings.ToLower(c)
	}
	return s[1:]
}

func buildPlaceholders(f, batch int) string {
	var s string
	p := 1
	for i := 0; i < batch; i++ {
		var v string
		for j := 0; j < f; j, p = j+1, p+1 {
			v = v + fmt.Sprintf(",$%d", p)
		}
		s = s + fmt.Sprintf(",(%s)", v[1:])
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
