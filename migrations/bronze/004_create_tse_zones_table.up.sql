BEGIN;

CREATE TABLE uf_tse
(
    cd VARCHAR(2) NOT NULL,
    ds VARCHAR(20) NOT NULL,
    PRIMARY KEY (cd)
);

CREATE TABLE municipio_tse
(
    cd VARCHAR(5) NOT NULL,
    cdi INT,
    nm VARCHAR(40) NOT NULL,
    c VARCHAR(1),
    uf_cd VARCHAR(2) NOT NULL,
    PRIMARY KEY (cd)
);

CREATE UNIQUE INDEX municipio_tse_cdi_idx ON municipio_tse(cdi);

CREATE TABLE zona_tse
(
    municipio_cd VARCHAR(5) NOT NULL,
    cd VARCHAR(4) NOT NULL,
    PRIMARY KEY (municipio_cd, cd),
    CONSTRAINT fk_municipio
        FOREIGN KEY(municipio_cd)
            REFERENCES municipio_tse(cd)
);

CREATE TABLE secao_tse
(
    municipio_cd VARCHAR(5) NOT NULL,
    zona_cd VARCHAR(4) NOT NULL,
    ns VARCHAR(4) NOT NULL,
    nsp VARCHAR(4) NOT NULL,
    PRIMARY KEY (municipio_cd, zona_cd, ns),
    CONSTRAINT fk_zona
        FOREIGN KEY(municipio_cd, zona_cd)
            REFERENCES zona_tse(municipio_cd, cd)
);

COMMIT;
