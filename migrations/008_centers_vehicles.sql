CREATE TABLE IF NOT EXISTS distribution_centers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    warehouse_id UUID NOT NULL,
    zip VARCHAR(10) NOT NULL DEFAULT '',
    street VARCHAR(150) NOT NULL DEFAULT '',
    number VARCHAR(20) NOT NULL DEFAULT '',
    complement VARCHAR(80) NOT NULL DEFAULT '',
    district VARCHAR(80) NOT NULL DEFAULT '',
    city VARCHAR(80) NOT NULL DEFAULT '',
    state VARCHAR(2) NOT NULL DEFAULT '',
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS delivery_vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    capacity_m3 NUMERIC(12,4) NOT NULL,
    capacity_kg NUMERIC(12,3) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO code_counters (name, n) VALUES ('distribution_center', 0), ('delivery_vehicle', 0)
ON CONFLICT (name) DO NOTHING;
