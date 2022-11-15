package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io"
	"net/http"
)

const electoralZonesURL = "https://resultados.tse.jus.br/oficial/ele2022/545/config/mun-e000545-cm.json"

func NewZonesProcessor(db *sqlx.DB, stepsService StepsService) *ZonesProcessor {
	return &ZonesProcessor{
		db:           db,
		url:          electoralZonesURL,
		stepsService: stepsService,
	}
}

type ZonesProcessor struct {
	db           *sqlx.DB
	url          string
	stepsService StepsService
}

type ZonesResponse struct {
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

func (p *ZonesProcessor) Run(ctx context.Context) error {
	return p.stepsService.Execute(ctx, "tse-zonas", p.process)
}

func (p *ZonesProcessor) process(ctx context.Context) error {
	resp, err := http.Get(p.url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var zonesResponse ZonesResponse
	err = json.Unmarshal(b, &zonesResponse)
	if err != nil {
		return err
	}

	for _, uf := range zonesResponse.ABR {
		s := fmt.Sprintf("tse-zonas-%s", uf.CD)
		err = p.stepsService.Execute(ctx, s, func(ctx context.Context) error {
			return p.processUF(ctx, uf)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *ZonesProcessor) processUF(ctx context.Context, uf ABRZone) error {
	if err := p.saveUF(ctx, &uf); err != nil {
		return err
	}

	for _, m := range uf.MU {
		p.processMunicipio(ctx, m, uf)
	}

	return nil
}

func (p *ZonesProcessor) processMunicipio(ctx context.Context, m MUZone, uf ABRZone) error {
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

func (p *ZonesProcessor) saveUF(ctx context.Context, uf *ABRZone) error {
	query := `INSERT INTO uf_tse(cd, ds) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, uf.CD, uf.DS)
	return err
}

func (p *ZonesProcessor) saveMunicipio(ctx context.Context, m *MUZone, uf string) error {
	query := `INSERT INTO municipio_tse(cd, cdi, nm, c, uf_cd) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`
	var cdi *string
	if m.CDI != "" {
		cdi = &m.CDI
	}
	_, err := p.db.ExecContext(ctx, query, m.CD, cdi, m.NM, m.C, uf)
	return err
}

func (p *ZonesProcessor) saveZona(ctx context.Context, z, m string) error {
	query := `INSERT INTO zona_tse(cd, municipio_cd) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, z, m)
	return err
}
