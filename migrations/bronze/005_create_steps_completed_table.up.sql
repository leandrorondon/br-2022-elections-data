BEGIN;

CREATE TABLE steps_completed
(
    step VARCHAR(32) NOT NULL,
    PRIMARY KEY (step)
);

COMMIT;
