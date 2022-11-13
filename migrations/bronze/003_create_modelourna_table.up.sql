BEGIN;

CREATE TABLE modelourna_numerointerno
(
    ds_modelo_urna INT NOT NULL,
    nr_faixa_inicial INT NOT NULL,
    nr_faixa_final INT NOT NULL,
    PRIMARY KEY (ds_modelo_urna)
);

COMMIT;
