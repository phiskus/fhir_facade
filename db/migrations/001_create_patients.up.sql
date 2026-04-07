CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE patients (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    version_id      INTEGER NOT NULL DEFAULT 1,
    last_updated    TIMESTAMPTZ NOT NULL DEFAULT now(),

    family_name     TEXT,
    given_name      TEXT,
    birth_date      DATE,
    gender          TEXT,
    active          BOOLEAN DEFAULT true,
    mrn             TEXT,
    phone           TEXT,
    email           TEXT,

    resource        JSONB NOT NULL,

    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_patients_family_name ON patients (lower(family_name));
CREATE INDEX idx_patients_given_name  ON patients (lower(given_name));
CREATE INDEX idx_patients_birth_date  ON patients (birth_date);
CREATE INDEX idx_patients_gender      ON patients (gender);
CREATE INDEX idx_patients_mrn         ON patients (mrn);
CREATE INDEX idx_patients_active      ON patients (active) WHERE deleted_at IS NULL;
CREATE INDEX idx_patients_resource_gin ON patients USING gin (resource);
