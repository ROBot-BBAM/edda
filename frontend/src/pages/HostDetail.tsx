import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import api, { setPortReviewed, setURLReviewed } from '../services/api';
import NotesModal from '../components/NotesModal';
import RowActionIcons from '../components/RowActionIcons';
import './HostDetail.css';

interface Host {
  id: string;
  ip_address: string;
  hostname?: string;
  os?: string;
  reviewed: boolean;
  note_count: number;
  note_preview: string;
  created_at: string;
}

interface Port {
  id: string;
  host_id: string;
  port: number;
  protocol: string;
  state?: string;
  service_name?: string;
  service_product?: string;
  service_version?: string;
  reviewed: boolean;
  finding_count: number;
  note_count: number;
  note_preview: string;
  created_at: string;
}

interface URL {
  id: string;
  url: string;
  path: string;
  method: string;
  status_code?: number;
  content_length?: number;
  words?: number;
  lines?: number;
  reviewed: boolean;
  finding_count: number;
  note_count: number;
  note_preview: string;
  created_at: string;
}

interface HostDetailData {
  host: Host;
  ports: Port[];
  urls: URL[];
}

const HostDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [data, setData] = useState<HostDetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [notesModal, setNotesModal] = useState<{ type: 'host' | 'port' | 'url'; id: string } | null>(null);

  useEffect(() => {
    if (id) {
      fetchHostDetail(id);
    }
  }, [id]);

  const fetchHostDetail = async (hostId: string) => {
    try {
      const response = await api.get(`/hosts/${hostId}`);
      setData(response.data);
    } catch (err) {
      console.error('Failed to fetch host detail', err);
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePortReviewed = async (port: Port) => {
    if (togglingId === port.id) return;
    setTogglingId(port.id);
    try {
      const updated = await setPortReviewed(port.id, !port.reviewed);
      setData(d => {
        if (!d) return null;
        const newPorts = d.ports.map(p => (p.id === port.id ? { ...p, reviewed: updated.reviewed } : p));
        const allPortsReviewed = newPorts.every(p => p.reviewed);
        const allUrlsReviewed = d.urls.every(u => u.reviewed);
        return { ...d, ports: newPorts, host: { ...d.host, reviewed: allPortsReviewed && allUrlsReviewed } };
      });
    } catch (err) {
      console.error('Failed to update port', err);
    } finally {
      setTogglingId(null);
    }
  };

  const handleToggleURLReviewed = async (urlObj: URL) => {
    if (togglingId === urlObj.id) return;
    setTogglingId(urlObj.id);
    try {
      const updated = await setURLReviewed(urlObj.id, !urlObj.reviewed);
      setData(d => {
        if (!d) return null;
        const newUrls = d.urls.map(u => (u.id === urlObj.id ? { ...u, reviewed: updated.reviewed } : u));
        const allPortsReviewed = d.ports.every(p => p.reviewed);
        const allUrlsReviewed = newUrls.every(u => u.reviewed);
        return { ...d, urls: newUrls, host: { ...d.host, reviewed: allPortsReviewed && allUrlsReviewed } };
      });
    } catch (err) {
      console.error('Failed to update URL', err);
    } finally {
      setTogglingId(null);
    }
  };

  const getStatusColor = (statusCode?: number) => {
    if (!statusCode) return '';
    if (statusCode >= 200 && statusCode < 300) return 'status-success';
    if (statusCode >= 300 && statusCode < 400) return 'status-redirect';
    if (statusCode >= 400 && statusCode < 500) return 'status-client-error';
    if (statusCode >= 500) return 'status-server-error';
    return '';
  };

  if (loading) {
    return (
      <div className="host-detail-page">
        <div className="container">
          <div className="card">Loading...</div>
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="host-detail-page">
        <div className="container">
          <div className="card">Host not found</div>
        </div>
      </div>
    );
  }

  return (
    <div className="host-detail-page">
      <div className="container">
        <div className="host-detail-content">
          <div className="host-info-card card">
            <h2>Host Information</h2>
            <div className="info-grid">
              <div className="info-item">
                <label>IP Address:</label>
                <span className="ip-address">{data.host.ip_address}</span>
              </div>
              {data.host.hostname && (
                <div className="info-item">
                  <label>Hostname:</label>
                  <span>{data.host.hostname}</span>
                </div>
              )}
              {data.host.os && (
                <div className="info-item">
                  <label>Operating System:</label>
                  <span>{data.host.os}</span>
                </div>
              )}
              <div className="info-item">
                <label>Status:</label>
                <span className={`status-badge ${data.host.reviewed ? 'reviewed' : 'unreviewed'}`}>
                  {data.host.reviewed ? 'Reviewed' : 'Unreviewed'}
                </span>
              </div>
              <div className="info-item">
                <label>Discovered:</label>
                <span>{new Date(data.host.created_at).toLocaleString()}</span>
              </div>
              <div className="info-item host-notes-actions">
                <label>Notes:</label>
                <span>
                  <span className={`note-badge ${(data.host.note_count ?? 0) > 0 ? 'has-notes' : ''}`} title={data.host.note_preview || undefined}>
                    {(data.host.note_count ?? 0) > 0 ? data.host.note_count : '0'}
                  </span>
                  <RowActionIcons onAddNote={() => setNotesModal({ type: 'host', id: data.host.id })} />
                </span>
              </div>
            </div>
          </div>

          {notesModal && (
            <NotesModal
              entityType={notesModal.type}
              entityId={notesModal.id}
              onClose={() => setNotesModal(null)}
              onAdded={() => id && fetchHostDetail(id)}
            />
          )}

          <div className="section">
            <h3>Open Ports ({data.ports.length})</h3>
            {data.ports.length === 0 ? (
              <div className="card">
                <p>No ports found for this host.</p>
              </div>
            ) : (
              <div className="card">
                <table className="ports-table">
                <thead>
                  <tr>
                    <th>Port</th>
                    <th>Protocol</th>
                    <th>State</th>
                    <th>Service</th>
                    <th>Product</th>
                    <th>Version</th>
                    <th>Findings</th>
                    <th>Notes</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                    {data.ports.map((port) => (
                      <tr key={port.id}>
                        <td className="port-number">{port.port}</td>
                        <td>{port.protocol.toUpperCase()}</td>
                        <td>{port.state || '-'}</td>
                        <td>{port.service_name || '-'}</td>
                        <td>{port.service_product || '-'}</td>
                        <td>{port.service_version || '-'}</td>
                        <td>
                          <span className={`finding-badge ${(port.finding_count ?? 0) > 0 ? 'has-findings' : ''}`} title={`${port.finding_count ?? 0} finding(s)`}>
                            {(port.finding_count ?? 0) > 0 ? (port.finding_count ?? 0) : '—'}
                          </span>
                        </td>
                        <td title={port.note_preview || undefined}>
                          <span className={`note-badge ${(port.note_count ?? 0) > 0 ? 'has-notes' : ''}`}>{(port.note_count ?? 0) > 0 ? port.note_count : '—'}</span>
                        </td>
                        <td>
                          <button
                            type="button"
                            className={`status-badge status-badge-btn ${port.reviewed ? 'reviewed' : 'unreviewed'}`}
                            onClick={() => handleTogglePortReviewed(port)}
                            disabled={togglingId === port.id}
                            title={port.reviewed ? 'Mark unreviewed' : 'Mark reviewed'}
                          >
                            {togglingId === port.id ? '…' : port.reviewed ? 'Reviewed' : 'Unreviewed'}
                          </button>
                        </td>
                        <td className="row-actions-cell">
                          <RowActionIcons
                            onAddNote={() => setNotesModal({ type: 'port', id: port.id })}
                            onAddFinding={() => navigate(`/findings?host_id=${data.host.id}&port_id=${port.id}`)}
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div className="section">
            <h3>Discovered URLs ({data.urls.length})</h3>
            {data.urls.length === 0 ? (
              <div className="card">
                <p>No URLs found for this host.</p>
              </div>
            ) : (
              <div className="card">
                <table className="urls-table">
                <thead>
                  <tr>
                    <th>URL</th>
                    <th>Method</th>
                    <th>Status</th>
                    <th>Length</th>
                    <th>Words</th>
                    <th>Lines</th>
                    <th>Findings</th>
                    <th>Notes</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                    {data.urls.map((urlObj) => (
                      <tr key={urlObj.id}>
                        <td className="url-cell">
                          <a href={urlObj.url} target="_blank" rel="noopener noreferrer" className="url-link">
                            {urlObj.path}
                          </a>
                        </td>
                        <td>{urlObj.method}</td>
                        <td>
                          {urlObj.status_code ? (
                            <span className={`status-code ${getStatusColor(urlObj.status_code)}`}>
                              {urlObj.status_code}
                            </span>
                          ) : '-'}
                        </td>
                        <td>{urlObj.content_length?.toLocaleString() || '-'}</td>
                        <td>{urlObj.words?.toLocaleString() || '-'}</td>
                        <td>{urlObj.lines?.toLocaleString() || '-'}</td>
                        <td>
                          <span className={`finding-badge ${(urlObj.finding_count ?? 0) > 0 ? 'has-findings' : ''}`} title={`${urlObj.finding_count ?? 0} finding(s)`}>
                            {(urlObj.finding_count ?? 0) > 0 ? (urlObj.finding_count ?? 0) : '—'}
                          </span>
                        </td>
                        <td title={urlObj.note_preview || undefined}>
                          <span className={`note-badge ${(urlObj.note_count ?? 0) > 0 ? 'has-notes' : ''}`}>{(urlObj.note_count ?? 0) > 0 ? urlObj.note_count : '—'}</span>
                        </td>
                        <td>
                          <button
                            type="button"
                            className={`status-badge status-badge-btn ${urlObj.reviewed ? 'reviewed' : 'unreviewed'}`}
                            onClick={() => handleToggleURLReviewed(urlObj)}
                            disabled={togglingId === urlObj.id}
                            title={urlObj.reviewed ? 'Mark unreviewed' : 'Mark reviewed'}
                          >
                            {togglingId === urlObj.id ? '…' : urlObj.reviewed ? 'Reviewed' : 'Unreviewed'}
                          </button>
                        </td>
                        <td className="row-actions-cell">
                          <RowActionIcons
                            onAddNote={() => setNotesModal({ type: 'url', id: urlObj.id })}
                            onAddFinding={() => navigate(`/findings?host_id=${data.host.id}&url_id=${urlObj.id}`)}
                          />
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
    </div>
  );
};

export default HostDetail;
