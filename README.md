# br-2022-elections-data
Fetch data from different sources and build a single database to analyse the results of the second round of 2022
Presidential Elections in Brazil.

## Data sources

### Locations

Locations data is obtained from the [IBGE Localidades API](https://servicodados.ibge.gov.br/api/docs/localidades)

### County Population 

Used to get the total population per county.

- Source: IBGE - TBD

### Election Results

Total vote count per candidate per ballot box.

- Source: Tribunal Superior Eleitoral - [Resultados 2022](https://dadosabertos.tse.jus.br/dataset/resultados-2022)

### Ballot Box Reports

Get the number of each ballot box, which is then used to infer their model.

- Source: Tribunal Superior Eleitoral -
[Boletim de Urna](https://dadosabertos.tse.jus.br/dataset/resultados-2022-boletim-de-urna)

## Data layers

The data is organised in three layers: Bronze, Silver and Gold.

### Bronze

This layer stores raw data, in the same structure as in its source.

### Silver

This layers contains transformed, aggregated and cleaned data. 
It normalises column names and data, and group related data from different sources.

For example, a same city has a different ID in IBGE and TSE, and cities population
come from a different IBGE source. In the Silver layer, cities and their population 
are saved in a single table.

### Gold

Gold layer stores aggregated data about the election results.

## Data structure

### Bronze

From Locations:
- `regioes` - id, nome, sigla
- `ufs` - id, nome, sigla, regiao_id
- `municipio` - id, nome, sigla, uf_id

