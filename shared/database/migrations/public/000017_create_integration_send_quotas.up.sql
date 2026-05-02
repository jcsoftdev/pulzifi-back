CREATE TABLE IF NOT EXISTS integration_send_quotas (
    org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    service_type  VARCHAR(32) NOT NULL,
    period_start  DATE NOT NULL,
    count_allowed INT NOT NULL,
    count_used    INT NOT NULL DEFAULT 0,
    PRIMARY KEY (org_id, service_type, period_start)
);
CREATE INDEX IF NOT EXISTS idx_quotas_period
    ON integration_send_quotas(period_start);
