CREATE TABLE payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(80) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(80) NOT NULL,
    installments JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO payment_methods (code, name) VALUES
    ('PIX', 'Pix'),
    ('CASH', 'Dinheiro'),
    ('BOLETO', 'Boleto'),
    ('CARD', 'Cartão');

INSERT INTO payment_terms (code, name, installments) VALUES
    ('AVISTA', 'À vista', '[{"days":0,"percent":100}]'),
    ('30', '30 dias', '[{"days":30,"percent":100}]'),
    ('2X3060', '2x 30/60', '[{"days":30,"percent":50},{"days":60,"percent":50}]'),
    ('3X306090', '3x 30/60/90', '[{"days":30,"percent":33.34},{"days":60,"percent":33.33},{"days":90,"percent":33.33}]');
