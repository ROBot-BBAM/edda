-- Findings (vulnerabilities) table: can be linked to a host, port, and/or URL
CREATE TABLE findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID REFERENCES hosts(id) ON DELETE CASCADE,
    port_id UUID REFERENCES ports(id) ON DELETE CASCADE,
    url_id UUID REFERENCES urls(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'medium',
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT findings_target_check CHECK (
        host_id IS NOT NULL OR port_id IS NOT NULL OR url_id IS NOT NULL
    )
);

CREATE INDEX idx_findings_host_id ON findings(host_id);
CREATE INDEX idx_findings_port_id ON findings(port_id);
CREATE INDEX idx_findings_url_id ON findings(url_id);

CREATE TRIGGER update_findings_updated_at BEFORE UPDATE ON findings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
