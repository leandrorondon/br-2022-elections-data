package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

const urlTemplate = "https://resultados.tse.jus.br/oficial/ele2022/arquivo-urna/407/config/%[1]s/%[1]s-p000407-cs.json"

var ufList = []string{
	"ac", "al", "am", "ap", "ba", "ce", "df", "es", "go", "ma", "mg", "ms", "mt", "pa",
	"pb", "pe", "pi", "pr", "rj", "rn", "ro", "rr", "rs", "sc", "se", "sp", "to", "zz",
}

func NewSectionsProcessor(db *sqlx.DB, stepsService StepsService) *SectionsProcessor {
	return &SectionsProcessor{
		db:           db,
		stepsService: stepsService,
	}
}

type SectionsProcessor struct {
	db           *sqlx.DB
	stepsService StepsService
}

type SectionsResponse struct {
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

func (p *SectionsProcessor) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	for _, uf := range ufList {
		u := uf
		g.Go(func() error {
			s := fmt.Sprintf("tse-secoes-%s", u)
			err := p.stepsService.Execute(gctx, s, func(ct context.Context) error {
				return p.processUF(ct, u, 1)
			})
			return err
		})
	}

	return g.Wait()
}

func (p *SectionsProcessor) processUF(ctx context.Context, uf string, retry int) error {
	url := fmt.Sprintf(urlTemplate, uf)
	resp, err := http.Get(url)
	if err != nil {
		if retry > 3 {
			return err
		}

		retry++
		log.Printf("Falha ao obter dados de %s. %da tentativa em 5s.", strings.ToUpper(uf), retry)
		time.Sleep(5 * time.Second)

		return p.processUF(ctx, uf, retry)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var sectionsResponse SectionsResponse
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

func (p *SectionsProcessor) processMunicipio(ctx context.Context, m MUSection, u ABRSection) error {
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

func (p *SectionsProcessor) processZona(ctx context.Context, z Zone, m MUSection) error {
	if err := p.saveZona(ctx, z.CD, m.CD); err != nil {
		return err
	}

	for _, s := range z.SEC {
		if err := p.saveSecao(ctx, z.CD, m.CD, &s); err != nil {
			return err
		}
	}

	return nil
}

func (p *SectionsProcessor) saveMunicipio(ctx context.Context, m *MUSection, uf string) error {
	query := `INSERT INTO municipio_tse(cd, nm, uf_cd) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, m.CD, m.NM, uf)
	return err
}

func (p *SectionsProcessor) saveZona(ctx context.Context, z, m string) error {
	query := `INSERT INTO zona_tse(cd, municipio_cd) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, z, m)
	return err
}

func (p *SectionsProcessor) saveSecao(ctx context.Context, m, z string, s *Section) error {
	query := `INSERT INTO secao_tse(municipio_cd, zona_cd, ns, nsp) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := p.db.ExecContext(ctx, query, z, m, s.NS, s.NSP)
	return err
}
