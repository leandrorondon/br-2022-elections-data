BEGIN;

CREATE TABLE regioes (
    id INT NOT NULL,
    nome VARCHAR(128) NOT NULL,
    sigla VARCHAR(2) NOT NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX regioes_sigla_idx ON regioes(sigla);

CREATE TABLE ufs (
    id INT NOT NULL,
    nome VARCHAR(128) NOT NULL,
    sigla VARCHAR(2) NOT NULL,
    regiao_id INT NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_regiao
        FOREIGN KEY(regiao_id)
        REFERENCES regioes(id)
);

CREATE UNIQUE INDEX uf_sigla_idx ON ufs(sigla);

CREATE TABLE municipios (
    id INT NOT NULL,
    nome VARCHAR(128) NOT NULL,
    uf_id INT NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_estado
        FOREIGN KEY(uf_id)
        REFERENCES ufs(id)
);

CREATE INDEX municipio_nome_idx ON municipios(nome);

COMMIT;