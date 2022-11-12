BEGIN;

CREATE TABLE populacao (
    localidade INT NOT NULL,
    populacao INT NOT NULL,
    PRIMARY KEY (localidade)
);

CREATE UNIQUE INDEX populacao_localidade_idx ON populacao(localidade);

COMMIT;