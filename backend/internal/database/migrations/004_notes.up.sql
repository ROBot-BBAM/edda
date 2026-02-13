-- Notes: one of host_id, port_id, or url_id must be set
CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID REFERENCES hosts(id) ON DELETE CASCADE,
    port_id UUID REFERENCES ports(id) ON DELETE CASCADE,
    url_id UUID REFERENCES urls(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT notes_target_check CHECK (
        (host_id IS NOT NULL AND port_id IS NULL AND url_id IS NULL)
        OR (host_id IS NULL AND port_id IS NOT NULL AND url_id IS NULL)
        OR (host_id IS NULL AND port_id IS NULL AND url_id IS NOT NULL)
    )
);

CREATE INDEX idx_notes_host_id ON notes(host_id);
CREATE INDEX idx_notes_port_id ON notes(port_id);
CREATE INDEX idx_notes_url_id ON notes(url_id);

CREATE TRIGGER update_notes_updated_at BEFORE UPDATE ON notes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
