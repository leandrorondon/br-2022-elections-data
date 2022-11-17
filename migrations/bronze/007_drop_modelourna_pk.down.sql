BEGIN;

ALTER TABLE modelourna_numerointerno
    ADD CONSTRAINT modelourna_numerointerno_pkey
        PRIMARY KEY (ds_modelo_urna);

COMMIT;
