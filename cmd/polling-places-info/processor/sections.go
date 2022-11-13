package processor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const urlTemplate = "https://resultados.tse.jus.br/oficial/ele2022/arquivo-urna/407/config/%s/%s-p000407-cs.json"

var ufList = []string{
	"ac", "al", "am", "ap", "ba", "ce", "df", "es", "go", "ma", "mg", "ms", "mt", "pa",
	"pb", "pe", "pi", "pr", "rj", "rn", "ro", "rr", "rs", "sc", "se", "sp", "to", "zz",
}

func NewSectionsProcessor(db *sql.DB) *SectionsProcessor {
	return &SectionsProcessor{
		db: db,
	}
}

type SectionsProcessor struct {
	db *sql.DB
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

func (p *SectionsProcessor) Run() {

	for _, uf := range ufList {
		url := fmt.Sprintf(urlTemplate, uf, uf)
		p.processUF(uf, url, 1)
	}
}

func (p *SectionsProcessor) processUF(uf, url string, retry int) {
	resp, err := http.Get(url)
	if err != nil {
		if retry > 3 {
			log.Fatal(err)
		}
		retry++
		log.Printf("Falha ao obter dados de %s. %da tentativa em 5s.", strings.ToUpper(uf), retry)
		time.Sleep(5 * time.Second)

		p.processUF(uf, url, retry)
		return
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var sectionsResponse SectionsResponse
	err = json.Unmarshal(b, &sectionsResponse)
	if err != nil {
		log.Fatal(err)
	}

	municipioCount := 0
	zoneCount := 0
	sectionCount := 0
	// for... for... for... for... for... argh! :(
	for _, uf := range sectionsResponse.ABR {
		log.Printf("Processando estado %s.", uf.CD)
		for _, m := range uf.MU {
			p.saveMunicipio(&m, uf.CD)
			municipioCount++
			for _, z := range m.ZON {
				p.saveZona(z.CD, m.CD)
				zoneCount++
				for _, s := range z.SEC {
					p.saveSecao(z.CD, m.CD, &s)
					sectionCount++
				}
			}
		}
	}
	log.Printf("Salvos %d seções de %d zonas de %d municípios.", sectionCount, zoneCount, municipioCount)
}

func (p *SectionsProcessor) saveMunicipio(m *MUSection, uf string) {
	query := `INSERT INTO municipio_tse(cd, nm, uf_cd) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, err := p.db.Exec(query, m.CD, m.NM, uf)
	if err != nil {
		fmt.Println(m, uf)
		log.Fatal("Erro salvando município: ", err)
	}
}

func (p *SectionsProcessor) saveZona(z, m string) {
	query := `INSERT INTO zona_tse(cd, municipio_cd) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := p.db.Exec(query, z, m)
	if err != nil {
		log.Fatal("Erro salvando zona: ", err)
	}
}

func (p *SectionsProcessor) saveSecao(m, z string, s *Section) {
	query := `INSERT INTO secao_tse(municipio_cd, zona_cd, ns, nsp) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := p.db.Exec(query, z, m, s.NS, s.NSP)
	if err != nil {
		log.Fatal("Erro salvando seção: ", err)
	}
}
