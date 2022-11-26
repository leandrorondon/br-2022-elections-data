package zones

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

const electoralZonesURL = "https://resultados.tse.jus.br/oficial/ele2022/545/config/mun-e000545-cm.json"

func New(db *sqlx.DB, stepsService zipcsv.StepsService) *Processor {
	return &Processor{
		db:           db,
		url:          electoralZonesURL,
		stepsService: stepsService,
	}
}

type Processor struct {
	db           *sqlx.DB
	url          string
	stepsService zipcsv.StepsService
}

type Response struct {
	DG  string
	HG  string
	F   string
	ABR []ABRZone
}

type ABRZone struct {
	CD string
	DS string
	MU []MUZone
}

type MUZone struct {
	CD  string
	CDI string
	NM  string
	C   string
	Z   []string
}

func (p *Processor) Run(ctx context.Context) error {
	return p.stepsService.Execute(ctx, "tse-zonas", p.process)
}

func (p *Processor) process(ctx context.Context) error {
	resp, err := httpwithretry.Get(p.url, 2)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var zonesResponse Response
	err = json.Unmarshal(b, &zonesResponse)
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, uf := range zonesResponse.ABR {
		u := uf
		g.Go(func() error {
			s := fmt.Sprintf("tse-zonas-%s", u.CD)
			err = p.stepsService.Execute(gctx, s, func(ct context.Context) error {
				return p.processUF(ct, u)
			})
			return err
		})
	}

	return g.Wait()
}

func (p *Processor) processUF(ctx context.Context, uf ABRZone) error {
	if err := p.saveUF(ctx, &uf); err != nil {
		return err
	}

	for _, m := range uf.MU {
		err := p.processMunicipio(ctx, m, uf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) processMunicipio(ctx context.Context, m MUZone, uf ABRZone) error {
	if err := p.saveMunicipio(ctx, &m, uf.CD); err != nil {
		return err
	}

	for _, z := range m.Z {
		if err := p.saveZona(ctx, z, m.CD); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) saveUF(ctx context.Context, uf *ABRZone) error {
	query := `INSERT INTO uf_tse(cd, ds) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, uf.CD, uf.DS)
	return err
}

func (p *Processor) saveMunicipio(ctx context.Context, m *MUZone, uf string) error {
	query := `INSERT INTO municipio_tse(cd, cdi, nm, c, uf_cd) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`
	var cdi *string
	if m.CDI != "" {
		cdi = &m.CDI
	}
	_, err := p.db.ExecContext(ctx, query, m.CD, cdi, m.NM, m.C, uf)
	return err
}

func (p *Processor) saveZona(ctx context.Context, z, m string) error {
	query := `INSERT INTO zona_tse(cd, municipio_cd) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, z, m)
	return err
}
