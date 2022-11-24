BEGIN;

CREATE TABLE votacao_secao
(
    dt_geracao VARCHAR(10) NOT NULL,
    hh_geracao TIME NOT NULL,
    ano_eleicao INT NOT NULL,
    cd_tipo_eleicao INT NOT NULL,
    nm_tipo_eleicao VARCHAR NOT NULL,
    nr_turno INT NOT NULL,
    cd_eleicao INT NOT NULL,
    ds_eleicao VARCHAR(50) NOT NULL,
    dt_eleicao VARCHAR(10) NOT NULL,
    tp_abrangencia VARCHAR(10),
    sg_uf VARCHAR(2) NOT NULL,
    sg_ue VARCHAR(3) NOT NULL,
    nm_ue VARCHAR(50) NOT NULL,
    cd_municipio INT NOT NULL,
    nm_municipio VARCHAR(100) NOT NULL,
    nr_zona INT NOT NULL,
    nr_secao INT NOT NULL,
    cd_cargo INT NOT NULL,
    ds_cargo VARCHAR(20) NOT NULL,
    nr_votavel INT NOT NULL,
    nm_votavel VARCHAR(20) NOT NULL,
    qt_votos INT NOT NULL,
    nr_local_votacao INT NOT NULL,
    sq_candidato BIGINT NOT NULL,
    nm_local_votacao VARCHAR,
    ds_local_votacao_endereco VARCHAR,

    PRIMARY KEY (cd_municipio, nr_zona, nr_secao, nr_votavel)
);

COMMIT;