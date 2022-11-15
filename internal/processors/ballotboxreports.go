package processors

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
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
	for _, uf := range ufList {
		url := fmt.Sprintf(relatorioUrnaURLTemplate, strings.ToUpper(uf))
		s := fmt.Sprintf("relatorio-urna-%s", uf)
		modelosUrna := NewZipCsvProcessor(
			"Relatórios de Urna", s, relatorioUrnaTable, url, p.db, p.stepsService)
		if err := modelosUrna.Run(ctx); err != nil {
			return err
		}
	}

	return nil
}
