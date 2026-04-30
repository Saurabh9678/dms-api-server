CREATE TYPE vehicle_status AS ENUM ('garage', 'inspection', 'ready_for_sale', 'sold');

CREATE TABLE IF NOT EXISTS vehicle_statuses (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
    status vehicle_status NOT NULL,
    description TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    added_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vehicle_statuses_vehicle_id ON vehicle_statuses(vehicle_id);
