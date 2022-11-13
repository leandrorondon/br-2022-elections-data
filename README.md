# br-2022-elections-data
Fetch data from different sources and build a single database to analyse the results of the second round of 2022
Presidential Elections in Brazil.

## Data sources

### Locations

Locations data is obtained from the [IBGE Localidades API](https://servicodados.ibge.gov.br/api/docs/localidades).

The data of interest are regions, states and cities.

### Indicators

Indicators data is obtained from the [IBGE Pesquisas API](https://servicodados.ibge.gov.br/api/docs/pesquisas).

The following indicators are fetched and saved:
- Total population (2010)
- Total population estimation (2021)
- Demographic density
- Average monthly salary of formal workers
- School enrollment rate from 6 to 14 year old
- GNP per capita (PIB)
- Municipal Human Development Index (IDHM)
- Child mortality rate

### Polling Places and Ballot Box info

Get the ballot box models and the list of electoral locations, zones and sections.

- Ballot Box model list: [Dados Abertos TSE](https://dadosabertos.tse.jus.br/dataset/correspondencia-entre-numero-interno-e-modelo-da-urna-1)
- ...

### Ballot Box Reports

Get the number of each ballot box, which is then used to infer their model.

- Source: Tribunal Superior Eleitoral -
  [Boletim de Urna](https://dadosabertos.tse.jus.br/dataset/resultados-2022-boletim-de-urna)

### Election Results

Total vote count per candidate per ballot box.

- Source: Tribunal Superior Eleitoral - [Resultados 2022](https://dadosabertos.tse.jus.br/dataset/resultados-2022)

## Data layers

The data is organised in three layers: Bronze, Silver and Gold.

### Bronze

This layer stores raw data, in the same structure as in its source.

### Silver

This layer contains transformed, aggregated and cleaned data.
It normalises column names and data, and group related data from different sources.

For example, a same city has a different ID in IBGE and TSE, and cities indicators
come from a different IBGE source. In the Silver layer, cities and their indicators
are saved in a single table.

### Gold

Gold layer stores aggregated data about the election results.

## Data structure

### Bronze

From Location:
- `regioes` - id, nome, sigla
- `ufs` - id, nome, sigla, regiao_id
- `municipio` - id, nome, sigla, uf_id

From Indicators:
- `populacao` - localidade, populacao
