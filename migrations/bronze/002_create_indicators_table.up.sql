BEGIN;

CREATE TABLE indicador_populacao
(
    localidade INT NOT NULL,
    populacao  INT NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_populacao_estimada
(
    localidade INT NOT NULL,
    populacao_estimada  INT NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_densidade_demografica
(
    localidade INT NOT NULL,
    densidade_demografica  REAL NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_salario_medio
(
    localidade INT NOT NULL,
    salario_medio REAL NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_taxa_escolarizacao
(
    localidade INT NOT NULL,
    taxa_escolarizacao REAL NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_pib_per_capita
(
    localidade INT NOT NULL,
    pib_per_capita REAL NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_idhm
(
    localidade INT NOT NULL,
    idhm REAL NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE TABLE indicador_taxa_mortalidade_infantil
(
    localidade INT NOT NULL,
    taxa_mortalidade_infantil REAL NOT NULL,
    PRIMARY KEY (localidade)
);

COMMIT;
