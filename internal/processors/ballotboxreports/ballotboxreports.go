package ballotboxreports

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/leandrorondon/br-2022-elections-data/internal/processors/zipcsv"
	"golang.org/x/sync/errgroup"
)

const (
	relatorioUrnaTable       = "relatorio_urna"
	relatorioUrnaURLTemplate = "https://cdn.tse.jus.br/estatistica/sead/eleicoes/eleicoes2022/buweb/bweb_2t_%s_311020221535.zip"
)

var ufList = []string{
	"ac", "al", "am", "ap", "ba", "ce", "df", "es", "go", "ma", "mg", "ms", "mt", "pa",
	"pb", "pe", "pi", "pr", "rj", "rn", "ro", "rr", "rs", "sc", "se", "sp", "to", "zz",
}

func New(db *sqlx.DB, stepsService zipcsv.StepsService) *Processor {
	return &Processor{
		db:           db,
		stepsService: stepsService,
	}
}

type Processor struct {
	db           *sqlx.DB
	stepsService zipcsv.StepsService
}

func (p *Processor) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)
	for _, uf := range ufList {
		p.processUF(gctx, g, uf)
	}

	return g.Wait()
}

func (p *Processor) processUF(ctx context.Context, g *errgroup.Group, uf string) {
	g.Go(func() error {
		url := fmt.Sprintf(relatorioUrnaURLTemplate, strings.ToUpper(uf))
		s := fmt.Sprintf("relatorio-urna-%s", uf)
		config := zipcsv.Config{
			Name:  "Relatórios de Urna",
			Step:  s,
			Table: relatorioUrnaTable,
			URL:   url,
		}
		modelosUrna := zipcsv.New(p.db, p.stepsService, config, zipcsv.WithKeepDownload(true))
		if err := modelosUrna.Run(ctx); err != nil {
			return err
		}
		return nil
	})
}
