package sections

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/leandrorondon/br-2022-elections-data/internal/httpwithretry"
	"github.com/leandrorondon/br-2022-elections-data/internal/processors/zipcsv"
	"golang.org/x/sync/errgroup"
)

const urlTemplate = "https://resultados.tse.jus.br/oficial/ele2022/arquivo-urna/407/config/%[1]s/%[1]s-p000407-cs.json"

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

type Response struct {
	DG  string
	HG  string
	F   string
	CDP string
	ABR []ABRSection
}

type ABRSection struct {
	CD string
	DS string
	MU []MUSection
}

type MUSection struct {
	CD  string
	NM  string
	ZON []Zone
}

type Zone struct {
	CD  string
	SEC []Section
}

type Section struct {
	NS  string
	NSP string
}

func (p *Processor) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	for _, uf := range ufList {
		u := uf
		s := fmt.Sprintf("tse-secoes-%s", u)
		g.Go(func() error {
			err := p.stepsService.Execute(gctx, s, func(ct context.Context) error {
				return p.processUF(ct, u)
			})
			return err
		})
	}

	return g.Wait()
}

func (p *Processor) processUF(ctx context.Context, uf string) error {
	url := fmt.Sprintf(urlTemplate, uf)
	resp, err := httpwithretry.Get(ctx, url, 2)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var sectionsResponse Response
	err = json.Unmarshal(b, &sectionsResponse)
	if err != nil {
		return err
	}

	u := sectionsResponse.ABR[0]
	for _, m := range u.MU {
		if err := p.processMunicipio(ctx, m, u); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) processMunicipio(ctx context.Context, m MUSection, u ABRSection) error {
	if err := p.saveMunicipio(ctx, &m, u.CD); err != nil {
		return err
	}
	for _, z := range m.ZON {
		if err := p.processZona(ctx, z, m); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) processZona(ctx context.Context, z Zone, m MUSection) error {
	if err := p.saveZona(ctx, z.CD, m.CD); err != nil {
		return err
	}

	for _, s := range z.SEC {
		sec := s
		if err := p.saveSecao(ctx, z.CD, m.CD, &sec); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) saveMunicipio(ctx context.Context, m *MUSection, uf string) error {
	query := `INSERT INTO municipio_tse(cd, nm, uf_cd) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, m.CD, m.NM, uf)
	return err
}

func (p *Processor) saveZona(ctx context.Context, z, m string) error {
	query := `INSERT INTO zona_tse(cd, municipio_cd) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, z, m)
	return err
}

func (p *Processor) saveSecao(ctx context.Context, m, z string, s *Section) error {
	query := `INSERT INTO secao_tse(municipio_cd, zona_cd, ns, nsp) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, z, m, s.NS, s.NSP)
	return err
}
