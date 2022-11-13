package processor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const electoralZonesURL = "https://resultados.tse.jus.br/oficial/ele2022/545/config/mun-e000545-cm.json"

func NewZonesProcessor(db *sql.DB) *ZonesProcessor {
	return &ZonesProcessor{
		db:  db,
		url: electoralZonesURL,
	}
}

type ZonesProcessor struct {
	db  *sql.DB
	url string
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

func (p *ZonesProcessor) Run() {
	resp, err := http.Get(p.url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var zonesResponse ZonesResponse
	err = json.Unmarshal(b, &zonesResponse)
	if err != nil {
		log.Fatal(err)
	}

	municipioCount := 0
	zoneCount := 0
	for _, uf := range zonesResponse.ABR {
		p.saveUF(&uf)

		log.Printf("Processando estado %s.", uf.CD)
		for _, m := range uf.MU {
			p.saveMunicipio(&m, uf.CD)
			municipioCount++
			for _, z := range m.Z {
				p.saveZona(z, m.CD)
				zoneCount++
			}
		}
	}
	log.Printf("Salvos %d zonas de %d municípios.", zoneCount, municipioCount)
}

func (p *ZonesProcessor) saveUF(uf *ABRZone) {
	query := `INSERT INTO uf_tse(cd, ds) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.Exec(query, uf.CD, uf.DS)
	if err != nil {
		log.Fatal("Erro salvando UF: ", err)
	}
}

func (p *ZonesProcessor) saveMunicipio(m *MUZone, uf string) {
	query := `INSERT INTO municipio_tse(cd, cdi, nm, c, uf_cd) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`
	var cdi *string
	if m.CDI != "" {
		cdi = &m.CDI
	}
	_, err := p.db.Exec(query, m.CD, cdi, m.NM, m.C, uf)
	if err != nil {
		fmt.Println(m, uf)
		log.Fatal("Erro salvando município: ", err)
	}
}

func (p *ZonesProcessor) saveZona(z, m string) {
	query := `INSERT INTO zona_tse(cd, municipio_cd) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.Exec(query, z, m)
	if err != nil {
		log.Fatal("Erro salvando zona: ", err)
	}
}
