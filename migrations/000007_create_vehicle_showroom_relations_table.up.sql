CREATE TABLE IF NOT EXISTS vehicle_showroom_relations (
    id BIGSERIAL PRIMARY KEY,
    showroom_id BIGINT NOT NULL REFERENCES showrooms(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE (showroom_id, vehicle_id)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_showroom_relations_vehicle_id ON vehicle_showroom_relations(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_showroom_relations_showroom_id ON vehicle_showroom_relations(showroom_id);
