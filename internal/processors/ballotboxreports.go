package processors

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

const (
	relatorioUrnaTable       = "relatorio_urna"
	relatorioUrnaURLTemplate = "https://cdn.tse.jus.br/estatistica/sead/eleicoes/eleicoes2022/buweb/bweb_2t_%s_311020221535.zip"
)

func NewBallotBoxReportsProcessor(db *sqlx.DB, stepsService StepsService) *BallotBoxReportsProcessor {
	return &BallotBoxReportsProcessor{
		db:           db,
		stepsService: stepsService,
	}
}

type BallotBoxReportsProcessor struct {
	db           *sqlx.DB
	stepsService StepsService
}

func (p *BallotBoxReportsProcessor) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)
	for _, uf := range ufList {
		p.processUF(gctx, g, uf)
	}

	return g.Wait()
}

func (p *BallotBoxReportsProcessor) processUF(ctx context.Context, g *errgroup.Group, uf string) {
	g.Go(func() error {
		url := fmt.Sprintf(relatorioUrnaURLTemplate, strings.ToUpper(uf))
		s := fmt.Sprintf("relatorio-urna-%s", uf)
		config := ZipCsvConfig{
			Name:  "Relatórios de Urna",
			Step:  s,
			Table: relatorioUrnaTable,
			URL:   url,
		}
		modelosUrna := NewZipCsvProcessor(p.db, p.stepsService, config)
		if err := modelosUrna.Run(ctx); err != nil {
			return err
		}
		return nil
	})
}
