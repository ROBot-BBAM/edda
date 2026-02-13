import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import api, { listFindings, createFinding, updateFinding, deleteFinding, getFindingsSummary, type Finding, type CreateFindingParams } from '../services/api';
import './Findings.css';

interface HostForSelect {
  id: string;
  ip_address: string;
  hostname?: string;
}

interface PortForSelect {
  id: string;
  port: number;
  protocol: string;
}

interface URLForSelect {
  id: string;
  path: string;
  method: string;
}

const SEVERITIES = ['low', 'medium', 'high', 'critical'];
const STATUSES = ['open', 'fixed', 'accepted', 'wontfix'];

const Findings: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const prefillHostId = searchParams.get('host_id') || undefined;
  const prefillPortId = searchParams.get('port_id') || undefined;
  const prefillUrlId = searchParams.get('url_id') || undefined;
  const hasPrefill = !!(prefillHostId || prefillPortId || prefillUrlId);

  const [findings, setFindings] = useState<Finding[]>([]);
  const [summary, setSummary] = useState<Record<string, number> | null>(null);
  const [openCount, setOpenCount] = useState<number>(0);
  const [filterSeverity, setFilterSeverity] = useState<string>('');
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [hosts, setHosts] = useState<HostForSelect[]>([]);
  const [urls, setUrls] = useState<URLForSelect[]>([]);
  const [selectedHostId, setSelectedHostId] = useState<string>(prefillHostId || '');
  const [portsForHost, setPortsForHost] = useState<PortForSelect[]>([]);
  const [showForm, setShowForm] = useState(hasPrefill);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState('');
  const [editingFinding, setEditingFinding] = useState<Finding | null>(null);
  const [editForm, setEditForm] = useState<CreateFindingParams & { port_id?: string; url_id?: string }>({
    title: '', severity: 'medium', description: '', status: 'open', host_id: undefined, port_id: undefined, url_id: undefined,
  });
  const [editSubmitting, setEditSubmitting] = useState(false);
  const [editError, setEditError] = useState('');

  const [form, setForm] = useState<CreateFindingParams & { port_id?: string; url_id?: string }>({
    title: '',
    severity: 'medium',
    description: '',
    status: 'open',
    host_id: prefillHostId || undefined,
    port_id: prefillPortId || undefined,
    url_id: prefillUrlId || undefined,
  });

  const fetchFindings = useCallback(async (severity?: string, status?: string) => {
    try {
      setLoading(true);
      const data = await listFindings({
        ...(severity && { severity }),
        ...(status && { status }),
      });
      setFindings(data);
    } catch (err) {
      console.error('Failed to fetch findings', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchSummary = useCallback(async () => {
    try {
      const res = await getFindingsSummary();
      setSummary(res.by_severity || null);
      setOpenCount(res.open_count ?? 0);
    } catch (err) {
      console.error('Failed to fetch findings summary', err);
    }
  }, []);

  useEffect(() => {
    fetchFindings(filterSeverity || undefined, filterStatus || undefined);
  }, [fetchFindings, filterSeverity, filterStatus]);

  useEffect(() => {
    fetchSummary();
  }, [fetchSummary]);

  useEffect(() => {
    const loadHosts = async () => {
      try {
        const res = await api.get<HostForSelect[]>('/hosts');
        setHosts(Array.isArray(res.data) ? res.data : []);
      } catch (e) {
        console.error('Failed to fetch hosts', e);
      }
    };
    loadHosts();
  }, []);

  useEffect(() => {
    const loadUrls = async () => {
      try {
        const res = await api.get<URLForSelect[]>('/urls');
        const list = Array.isArray(res.data) ? res.data : [];
        setUrls(list.map((u: any) => ({ id: u.id, path: u.path || u.url || '', method: u.method || 'GET' })));
      } catch (e) {
        console.error('Failed to fetch URLs', e);
      }
    };
    loadUrls();
  }, []);

  useEffect(() => {
    setSelectedHostId((prev) => prefillHostId || prev);
    if (prefillHostId || prefillPortId || prefillUrlId) {
      setForm((f) => ({
        ...f,
        host_id: prefillHostId || f.host_id,
        port_id: prefillPortId || f.port_id,
        url_id: prefillUrlId || f.url_id,
      }));
    }
  }, [prefillHostId, prefillPortId, prefillUrlId]);

  useEffect(() => {
    if (!selectedHostId) {
      setPortsForHost([]);
      return;
    }
    const loadPorts = async () => {
      try {
        const res = await api.get<{ ports: PortForSelect[] }>(`/hosts/${selectedHostId}`);
        setPortsForHost(res.data?.ports || []);
      } catch (e) {
        setPortsForHost([]);
      }
    };
    loadPorts();
  }, [selectedHostId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    const title = form.title?.trim();
    if (!title) {
      setFormError('Title is required.');
      return;
    }
    const hostId = form.host_id?.trim() || undefined;
    const portId = form.port_id?.trim() || undefined;
    const urlId = form.url_id?.trim() || undefined;
    if (!hostId && !portId && !urlId) {
      setFormError('Select at least one: Host, Port, or URL.');
      return;
    }
    setSubmitting(true);
    try {
      await createFinding({
        title,
        severity: form.severity || 'medium',
        description: form.description?.trim() || undefined,
        status: form.status || 'open',
        host_id: hostId,
        port_id: portId,
        url_id: urlId,
      });
      setForm({ title: '', severity: 'medium', description: '', status: 'open', host_id: prefillHostId || undefined, port_id: prefillPortId || undefined, url_id: prefillUrlId || undefined });
      setShowForm(false);
      await fetchFindings(filterSeverity || undefined, filterStatus || undefined);
      await fetchSummary();
    } catch (err: any) {
      setFormError(err.response?.data?.error || 'Failed to create finding.');
    } finally {
      setSubmitting(false);
    }
  };

  const linkDisplay = (f: Finding) => {
    const parts = [];
    if (f.host_display) parts.push(f.host_display);
    if (f.port_display) parts.push(f.port_display);
    if (f.url_display) parts.push(f.url_display.length > 40 ? f.url_display.slice(0, 37) + '...' : f.url_display);
    return parts.length ? parts.join(' · ') : '—';
  };

  const [exportOpen, setExportOpen] = useState(false);
  const exportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const close = (e: MouseEvent) => {
      if (exportRef.current && !exportRef.current.contains(e.target as Node)) setExportOpen(false);
    };
    document.addEventListener('click', close);
    return () => document.removeEventListener('click', close);
  }, []);

  const downloadBlob = (blob: Blob, filename: string) => {
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename;
    link.click();
    URL.revokeObjectURL(link.href);
    setExportOpen(false);
  };

  const exportCSV = () => {
    const headers = ['Title', 'Severity', 'Status', 'Host', 'Port', 'URL', 'Description', 'Created', 'Updated'];
    const rows = findings.map((f) => [
      `"${(f.title || '').replace(/"/g, '""')}"`,
      f.severity,
      f.status,
      f.host_display || '',
      f.port_display || '',
      `"${(f.url_display || '').replace(/"/g, '""')}"`,
      `"${(f.description || '').replace(/"/g, '""')}"`,
      f.created_at,
      f.updated_at,
    ]);
    const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
    downloadBlob(new Blob([csv], { type: 'text/csv;charset=utf-8;' }), `findings-${new Date().toISOString().slice(0, 10)}.csv`);
  };

  const exportJSON = () => {
    const data = findings.map((f) => ({
      title: f.title,
      severity: f.severity,
      status: f.status,
      host: f.host_display || null,
      port: f.port_display || null,
      url: f.url_display || null,
      description: f.description || null,
      created_at: f.created_at,
      updated_at: f.updated_at,
    }));
    const json = JSON.stringify(data, null, 2);
    downloadBlob(new Blob([json], { type: 'application/json' }), `findings-${new Date().toISOString().slice(0, 10)}.json`);
  };

  const openEdit = (f: Finding) => {
    setEditingFinding(f);
    setEditForm({
      title: f.title,
      severity: f.severity || 'medium',
      description: f.description || '',
      status: f.status || 'open',
      host_id: f.host_id,
      port_id: f.port_id,
      url_id: f.url_id,
    });
    setEditError('');
    setSelectedHostId(f.host_id || '');
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingFinding) return;
    setEditError('');
    const title = editForm.title?.trim();
    if (!title) {
      setEditError('Title is required.');
      return;
    }
    const hostId = editForm.host_id?.trim() || undefined;
    const portId = editForm.port_id?.trim() || undefined;
    const urlId = editForm.url_id?.trim() || undefined;
    if (!hostId && !portId && !urlId) {
      setEditError('Select at least one: Host, Port, or URL.');
      return;
    }
    setEditSubmitting(true);
    try {
      // Send empty string for cleared host/port/url so the backend persists removal (omitting would keep existing)
      await updateFinding(editingFinding.id, {
        title: editForm.title,
        severity: editForm.severity || 'medium',
        description: editForm.description?.trim() || undefined,
        status: editForm.status || 'open',
        host_id: editForm.host_id?.trim() || '',
        port_id: editForm.port_id?.trim() || '',
        url_id: editForm.url_id?.trim() || '',
      });
      setEditingFinding(null);
      await fetchFindings(filterSeverity || undefined, filterStatus || undefined);
      await fetchSummary();
    } catch (err: any) {
      setEditError(err.response?.data?.error || 'Failed to update finding.');
    } finally {
      setEditSubmitting(false);
    }
  };

  return (
    <div className="findings-page">
      <div className="container">
        <div className="findings-content">
          <div className="page-header">
            <h2>Vulnerabilities &amp; Findings</h2>
            <div className="page-header-actions">
              <div className="export-dropdown" ref={exportRef}>
                <button
                  type="button"
                  className="btn btn-secondary export-dropdown-trigger"
                  onClick={() => setExportOpen((o) => !o)}
                  disabled={findings.length === 0}
                  title="Export filtered findings"
                >
                  Export <span className="export-caret" aria-hidden>▾</span>
                </button>
                {exportOpen && (
                  <div className="export-dropdown-menu">
                    <button type="button" className="export-dropdown-item" onClick={exportCSV}>CSV</button>
                    <button type="button" className="export-dropdown-item" onClick={exportJSON}>JSON</button>
                  </div>
                )}
              </div>
              <button type="button" className="btn btn-primary" onClick={() => setShowForm((s) => !s)}>
                {showForm ? 'Cancel' : 'Add finding'}
              </button>
            </div>
          </div>

          {summary !== null && (
            <div className="findings-summary-bar">
              {SEVERITIES.map((sev) => (
                <span key={sev} className={`findings-summary-item severity-${sev}`}>
                  {sev}: {summary?.[sev] ?? 0}
                </span>
              ))}
              <span className="findings-summary-item open-count">
                open: {openCount}
              </span>
            </div>
          )}

          <div className="findings-filters">
            <label htmlFor="filter-severity">Severity</label>
            <select id="filter-severity" value={filterSeverity} onChange={(e) => setFilterSeverity(e.target.value)}>
              <option value="">All</option>
              {SEVERITIES.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
            <label htmlFor="filter-status">Status</label>
            <select id="filter-status" value={filterStatus} onChange={(e) => setFilterStatus(e.target.value)}>
              <option value="">All</option>
              {STATUSES.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
          </div>

          {showForm && (
            <div className="card form-card">
              <h3>New finding</h3>
              <form onSubmit={handleSubmit}>
                {formError && <p className="form-error">{formError}</p>}
                <div className="form-row">
                  <label htmlFor="finding-title">Title *</label>
                  <input
                    id="finding-title"
                    type="text"
                    value={form.title}
                    onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                    placeholder="Short title"
                    required
                  />
                </div>
                <div className="form-row">
                  <label htmlFor="finding-severity">Severity</label>
                  <select
                    id="finding-severity"
                    value={form.severity}
                    onChange={(e) => setForm((f) => ({ ...f, severity: e.target.value }))}
                  >
                    {SEVERITIES.map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </div>
                <div className="form-row">
                  <label htmlFor="finding-status">Status</label>
                  <select
                    id="finding-status"
                    value={form.status}
                    onChange={(e) => setForm((f) => ({ ...f, status: e.target.value }))}
                  >
                    {STATUSES.map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </div>
                <div className="form-row">
                  <label htmlFor="finding-description">Description</label>
                  <textarea
                    id="finding-description"
                    value={form.description || ''}
                    onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                    placeholder="Optional details"
                    rows={3}
                  />
                </div>
                <div className="form-section">
                  <span className="form-section-label">Link to (at least one)</span>
                  <div className="form-row">
                    <label htmlFor="finding-host">Host</label>
                    <select
                      id="finding-host"
                      value={form.host_id || ''}
                      onChange={(e) => {
                        const v = e.target.value || undefined;
                        setForm((f) => ({ ...f, host_id: v, port_id: undefined }));
                        setSelectedHostId(v || '');
                      }}
                    >
                      <option value="">—</option>
                      {hosts.map((h) => (
                        <option key={h.id} value={h.id}>{h.hostname || h.ip_address}</option>
                      ))}
                    </select>
                  </div>
                  <div className="form-row">
                    <label htmlFor="finding-port">Port (on selected host)</label>
                    <select
                      id="finding-port"
                      value={form.port_id || ''}
                      onChange={(e) => setForm((f) => ({ ...f, port_id: e.target.value || undefined }))}
                      disabled={!selectedHostId}
                    >
                      <option value="">—</option>
                      {portsForHost.map((p) => (
                        <option key={p.id} value={p.id}>{p.port}/{p.protocol}</option>
                      ))}
                    </select>
                  </div>
                  <div className="form-row">
                    <label htmlFor="finding-url">URL</label>
                    <select
                      id="finding-url"
                      value={form.url_id || ''}
                      onChange={(e) => setForm((f) => ({ ...f, url_id: e.target.value || undefined }))}
                    >
                      <option value="">—</option>
                      {urls.slice(0, 200).map((u) => (
                        <option key={u.id} value={u.id}>{u.method} {u.path.length > 50 ? u.path.slice(0, 47) + '...' : u.path}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div className="form-actions">
                  <button type="submit" className="btn btn-primary" disabled={submitting}>
                    {submitting ? 'Creating…' : 'Create finding'}
                  </button>
                </div>
              </form>
            </div>
          )}

          {editingFinding && (
            <div className="modal-overlay" onClick={() => setEditingFinding(null)}>
              <div className="modal-content card" onClick={(e) => e.stopPropagation()}>
                <h3>Edit finding</h3>
                <form onSubmit={handleEditSubmit}>
                  {editError && <p className="form-error">{editError}</p>}
                  <div className="form-row">
                    <label htmlFor="edit-title">Title *</label>
                    <input
                      id="edit-title"
                      type="text"
                      value={editForm.title}
                      onChange={(e) => setEditForm((f) => ({ ...f, title: e.target.value }))}
                      placeholder="Short title"
                      required
                    />
                  </div>
                  <div className="form-row">
                    <label htmlFor="edit-severity">Severity</label>
                    <select
                      id="edit-severity"
                      value={editForm.severity}
                      onChange={(e) => setEditForm((f) => ({ ...f, severity: e.target.value }))}
                    >
                      {SEVERITIES.map((s) => (
                        <option key={s} value={s}>{s}</option>
                      ))}
                    </select>
                  </div>
                  <div className="form-row">
                    <label htmlFor="edit-status">Status</label>
                    <select
                      id="edit-status"
                      value={editForm.status}
                      onChange={(e) => setEditForm((f) => ({ ...f, status: e.target.value }))}
                    >
                      {STATUSES.map((s) => (
                        <option key={s} value={s}>{s}</option>
                      ))}
                    </select>
                  </div>
                  <div className="form-row">
                    <label htmlFor="edit-description">Description</label>
                    <textarea
                      id="edit-description"
                      value={editForm.description || ''}
                      onChange={(e) => setEditForm((f) => ({ ...f, description: e.target.value }))}
                      placeholder="Optional details"
                      rows={3}
                    />
                  </div>
                  <div className="form-section">
                    <span className="form-section-label">Link to (at least one)</span>
                    <div className="form-row">
                      <label>Host</label>
                      <select
                        value={editForm.host_id || ''}
                        onChange={(e) => {
                          const v = e.target.value || undefined;
                          setEditForm((f) => ({ ...f, host_id: v, port_id: undefined }));
                          setSelectedHostId(v || '');
                          if (v) {
                            api.get(`/hosts/${v}`).then((res) => setPortsForHost(res.data?.ports || [])).catch(() => setPortsForHost([]));
                          } else setPortsForHost([]);
                        }}
                      >
                        <option value="">—</option>
                        {hosts.map((h) => (
                          <option key={h.id} value={h.id}>{h.hostname || h.ip_address}</option>
                        ))}
                      </select>
                    </div>
                    <div className="form-row">
                      <label>Port (on selected host)</label>
                      <select
                        value={editForm.port_id || ''}
                        onChange={(e) => setEditForm((f) => ({ ...f, port_id: e.target.value || undefined }))}
                        disabled={!selectedHostId}
                      >
                        <option value="">—</option>
                        {portsForHost.map((p) => (
                          <option key={p.id} value={p.id}>{p.port}/{p.protocol}</option>
                        ))}
                      </select>
                    </div>
                    <div className="form-row">
                      <label>URL</label>
                      <select
                        value={editForm.url_id || ''}
                        onChange={(e) => setEditForm((f) => ({ ...f, url_id: e.target.value || undefined }))}
                      >
                        <option value="">—</option>
                        {urls.slice(0, 200).map((u) => (
                          <option key={u.id} value={u.id}>{u.method} {u.path.length > 50 ? u.path.slice(0, 47) + '...' : u.path}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                  <div className="form-actions">
                    <button type="button" className="btn btn-secondary" onClick={() => setEditingFinding(null)}>
                      Cancel
                    </button>
                    <button type="submit" className="btn btn-primary" disabled={editSubmitting}>
                      {editSubmitting ? 'Saving…' : 'Save'}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}

          {loading ? (
            <div className="card">Loading...</div>
          ) : findings.length === 0 ? (
            <div className="card">
              <p>No findings yet. Add one with the button above.</p>
            </div>
          ) : (
            <div className="card">
              <table className="findings-table">
                <thead>
                  <tr>
                    <th>Title</th>
                    <th>Severity</th>
                    <th>Status</th>
                    <th>Linked to</th>
                    <th>Created</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {findings.map((f) => (
                    <tr key={f.id}>
                      <td className="finding-title">{f.title}</td>
                      <td>
                        <span className={`severity-badge severity-${f.severity}`}>{f.severity}</span>
                      </td>
                      <td>
                        <span className={`status-badge status-${f.status}`}>{f.status}</span>
                      </td>
                      <td className="finding-link">{linkDisplay(f)}</td>
                      <td>{new Date(f.created_at).toLocaleString()}</td>
                      <td>
                        <button type="button" className="btn btn-sm btn-secondary" onClick={() => openEdit(f)}>
                          Edit
                        </button>
                        <button
                          type="button"
                          className="btn btn-sm btn-secondary"
                          onClick={async () => {
                            if (window.confirm('Delete this finding?')) {
                              await deleteFinding(f.id);
                              await fetchFindings(filterSeverity || undefined, filterStatus || undefined);
                              await fetchSummary();
                            }
                          }}
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Findings;
