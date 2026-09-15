CREATE TABLE code_counters (
    name TEXT PRIMARY KEY,
    n BIGINT NOT NULL DEFAULT 0
);

INSERT INTO code_counters (name, n)
SELECT 'unit', COALESCE(MAX(code::BIGINT), 0)
FROM units_of_measure
WHERE code ~ '^[0-9]+$';

INSERT INTO code_counters (name, n)
SELECT 'payment_method', COALESCE(MAX(code::BIGINT), 0)
FROM payment_methods
WHERE code ~ '^[0-9]+$';

INSERT INTO code_counters (name, n)
SELECT 'payment_term', COALESCE(MAX(code::BIGINT), 0)
FROM payment_terms
WHERE code ~ '^[0-9]+$';
