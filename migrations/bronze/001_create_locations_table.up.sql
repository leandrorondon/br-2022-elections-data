BEGIN;

CREATE TABLE regiao (
    id INT NOT NULL,
    nome VARCHAR(128) NOT NULL,
    sigla VARCHAR(2) NOT NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX regiao_sigla_idx ON regiao(sigla);

CREATE TABLE uf (
    id INT NOT NULL,
    nome VARCHAR(128) NOT NULL,
    sigla VARCHAR(2) NOT NULL,
    regiao_id INT NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_regiao
        FOREIGN KEY(regiao_id)
        REFERENCES regiao(id)
);

CREATE UNIQUE INDEX uf_sigla_idx ON uf(sigla);

CREATE TABLE municipio (
    id INT NOT NULL,
    nome VARCHAR(128) NOT NULL,
    uf_id INT NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_estado
        FOREIGN KEY(uf_id)
        REFERENCES uf(id)
);

CREATE INDEX municipio_nome_idx ON municipio(nome);

COMMIT;