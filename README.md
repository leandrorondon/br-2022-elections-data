# br-2022-elections-data
Fetch data from different sources and build a single database to analyse the results of the second round of 2022
Presidential Elections in Brazil.

## Data sources

### Locations

Locations data is obtained from the [IBGE Localidades API](https://servicodados.ibge.gov.br/api/docs/localidades).

The data of interest are regions, states and cities.

### Indicators

Indicator data is obtained from the [IBGE Pesquisas API](https://servicodados.ibge.gov.br/api/docs/pesquisas).

The following indicators are fetched and saved:
- Total population (2010)
- Total population estimation (2021)
- Demographic density
- Average monthly salary of formal workers
- School enrollment rate from 6 to 14 years old
- GNP per capita (PIB)
- Municipal Human Development Index (IDHM)
- Child mortality rate

### Polling Places and Ballot Box info

Get the ballot box models and the electoral locations, zones and sections list.

- Ballot Box model list: [Dados Abertos TSE](https://dadosabertos.tse.jus.br/dataset/correspondencia-entre-numero-interno-e-modelo-da-urna-1)
- List of Electoral districts and Zones: https://resultados.tse.jus.br/oficial/ele2022/545/config/mun-e000545-cm.json
- List of Electoral Sections: https://resultados.tse.jus.br/oficial/ele2022/arquivo-urna/407/config/XX/XX-p000407-cs.json (XX is the UF).

### Ballot Box Reports

- Source: Tribunal Superior Eleitoral -
  [Boletim de Urna](https://dadosabertos.tse.jus.br/dataset/resultados-2022-boletim-de-urna)

### Section Results

Total vote count per candidate per Section.

- Source: Tribunal Superior Eleitoral - [Resultados 2022](https://dadosabertos.tse.jus.br/dataset/resultados-2022/resource/f509562b-3b7f-487d-ad61-145a7ae6b96f)

## Data layers

The data is organised in three layers: Bronze, Silver and Gold.

### Bronze

This layer stores raw data, in the same structure as in its source.

### Silver

This layer contains transformed, aggregated and cleaned data.
It normalises column names and data, and groups related data from different sources.

For example, the same city has a different ID in IBGE and TSE, and city indicators
come from a different IBGE source. In the Silver layer, cities and their indicators
are saved in a single table.

### Gold

The gold layer stores aggregated data about the election results.

## Data structure

### Bronze

From Location:
- `regioes` - id, nome, sigla
- `ufs` - id, nome, sigla, regiao_id
- `municipio` - id, nome, sigla, uf_id

From Indicators:
- `populacao` - localidade, populacao