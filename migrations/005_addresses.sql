CREATE TABLE person_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    alias VARCHAR(80) NOT NULL DEFAULT '',
    zip VARCHAR(10) NOT NULL DEFAULT '',
    street VARCHAR(150) NOT NULL DEFAULT '',
    number VARCHAR(20) NOT NULL DEFAULT '',
    complement VARCHAR(80) NOT NULL DEFAULT '',
    district VARCHAR(80) NOT NULL DEFAULT '',
    city VARCHAR(80) NOT NULL DEFAULT '',
    state VARCHAR(2) NOT NULL DEFAULT '',
    lat NUMERIC(10,7),
    lng NUMERIC(10,7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_person_addresses_person ON person_addresses(person_id);

INSERT INTO person_addresses (person_id, alias, zip, street, number, complement, district, city, state)
SELECT id, 'Principal', zip, street, number, complement, district, city, state
FROM people
WHERE COALESCE(street, '') <> '' OR COALESCE(city, '') <> '' OR COALESCE(zip, '') <> '';

ALTER TABLE people
    DROP COLUMN street,
    DROP COLUMN number,
    DROP COLUMN complement,
    DROP COLUMN district,
    DROP COLUMN city,
    DROP COLUMN state,
    DROP COLUMN zip;
