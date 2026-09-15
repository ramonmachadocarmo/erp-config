CREATE TABLE units_of_measure (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(80) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO units_of_measure (code, name, symbol) VALUES
    ('UN', 'Unidade', 'un'),
    ('KG', 'Quilograma', 'kg'),
    ('G', 'Grama', 'g'),
    ('SACA', 'Saca', 'sc'),
    ('CX', 'Caixa', 'cx'),
    ('L', 'Litro', 'l'),
    ('ML', 'Mililitro', 'ml')
ON CONFLICT (code) DO NOTHING;
