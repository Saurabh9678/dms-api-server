CREATE TYPE vehicle_type AS ENUM ('bike', 'car', 'scooty');
CREATE TYPE fuel_type AS ENUM ('petrol', 'diesel', 'ev');
CREATE TYPE transmission_type AS ENUM ('manual', 'automatic');

CREATE TABLE IF NOT EXISTS vehicles (
    id BIGSERIAL PRIMARY KEY,
    type vehicle_type NOT NULL,
    manufacturer VARCHAR NOT NULL,
    model VARCHAR NOT NULL,
    variant VARCHAR NOT NULL,
    color VARCHAR NOT NULL,
    year_of_manufacture INT NOT NULL,
    rto_code VARCHAR NOT NULL,
    registration_number VARCHAR NOT NULL,
    registration_state VARCHAR NOT NULL,
    usage_km INT NOT NULL,
    fuel_type fuel_type NOT NULL,
    transmission_type transmission_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
