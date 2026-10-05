-- +goose Up
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_customers ON customers (LOWER(email));

CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    unit TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('quoted','pending','active','past_due','cancelled','lapsed','reinstated')),
    premium_cents BIGINT NOT NULL CHECK (premium_cents >= 0),
    effective_date DATE NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_policies ON policies (customer_id, unit, effective_date);

CREATE TABLE policy_events (
    id BIGSERIAL PRIMARY KEY,
    policy_id UUID NOT NULL REFERENCES policies(id),
    event TEXT NOT NULL CHECK (event IN ('submit','payment_succeeded','payment_failed','grace_period_expired','cancel','reinstate')),
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    actor TEXT DEFAULT NULL,
    occurred_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_events ON policy_events (policy_id, occurred_at);

CREATE TABLE roster_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_sha256 TEXT NOT NULL,
    filename TEXT NOT NULL,
    row_count INT NOT NULL,
    valid_count INT NOT NULL,
    error_count INT NOT NULL,
    errors JSONB DEFAULT NULL,
    uploaded_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_id_roster ON roster_uploads (file_sha256);

CREATE TABLE webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    received_at timestamptz NOT NULL DEfAULT NOW(),
    processed_at timestamptz DEFAULT NULL
);

CREATE UNIQUE INDEX idx_event_id_webhooks ON webhook_events (event_id);

-- +goose Down
DROP TABLE policy_events;
DROP TABLE policies;
DROP TABLE customers;
DROP TABLE webhook_events;
DROP TABLE roster_uploads;
