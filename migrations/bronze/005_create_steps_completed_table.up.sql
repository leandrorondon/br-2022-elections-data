BEGIN;

CREATE TABLE steps_completed
(
    step VARCHAR(100) NOT NULL,
    PRIMARY KEY (step)
);

COMMIT;
