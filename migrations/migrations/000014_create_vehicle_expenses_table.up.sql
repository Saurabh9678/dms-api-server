CREATE TABLE IF NOT EXISTS vehicle_expenses (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
    status_id BIGINT REFERENCES vehicle_statuses(id),
    type VARCHAR NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    paid_to VARCHAR,
    description TEXT,
    date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vehicle_expenses_vehicle_id ON vehicle_expenses(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_expenses_status_id ON vehicle_expenses(status_id);
