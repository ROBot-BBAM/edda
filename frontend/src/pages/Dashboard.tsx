import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api, { listFindings, getFindingsSummary } from '../services/api';
import './Dashboard.css';

interface ScanFile {
  id: string;
  type: string;
  filename: string;
  uploaded_by: string;
  uploaded_at: string;
  status: string;
  error_message?: string;
  created_at: string;
}

interface Host {
  id: string;
  ip_address: string;
  hostname?: string;
  os?: string;
  reviewed: boolean;
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
  created_at: string;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [scanFiles, setScanFiles] = useState<ScanFile[]>([]);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [ports, setPorts] = useState<Port[]>([]);
  const [urls, setUrls] = useState<URL[]>([]);
  const [findingsCount, setFindingsCount] = useState(0);
  const [findingsSummary, setFindingsSummary] = useState<Record<string, number> | null>(null);
  const [openFindingsCount, setOpenFindingsCount] = useState(0);
  const [needsAttention, setNeedsAttention] = useState<Array<{ id: string; title: string; severity: string }>>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState('');

  useEffect(() => {
    fetchAllData();
  }, []);

  const fetchFindings = async () => {
    try {
      const data = await listFindings();
      setFindingsCount(Array.isArray(data) ? data.length : 0);
    } catch (err) {
      console.error('Failed to fetch findings', err);
    }
  };

  const fetchFindingsSummary = async () => {
    try {
      const res = await getFindingsSummary();
      setFindingsSummary(res.by_severity || null);
      setOpenFindingsCount(res.open_count ?? 0);
    } catch (err) {
      console.error('Failed to fetch findings summary', err);
    }
  };

  const fetchNeedsAttention = async () => {
    try {
      const data = await listFindings({ status: 'open' });
      const list = Array.isArray(data) ? data : [];
      const criticalHigh = list
        .filter((f: { severity: string }) => f.severity === 'critical' || f.severity === 'high')
        .slice(0, 10)
        .map((f: { id: string; title: string; severity: string }) => ({ id: f.id, title: f.title, severity: f.severity }));
      setNeedsAttention(criticalHigh);
    } catch (err) {
      console.error('Failed to fetch needs attention', err);
    }
  };

  const fetchAllData = async () => {
    try {
      await Promise.all([
        fetchScanFiles(),
        fetchHosts(),
        fetchPorts(),
        fetchURLs(),
        fetchFindings(),
        fetchFindingsSummary(),
        fetchNeedsAttention(),
      ]);
    } catch (err) {
      console.error('Failed to fetch data', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchScanFiles = async () => {
    try {
      const response = await api.get('/scan-files');
      setScanFiles(response.data);
    } catch (err) {
      console.error('Failed to fetch scan files', err);
    }
  };

  const fetchHosts = async () => {
    try {
      const response = await api.get('/hosts');
      setHosts(response.data);
    } catch (err) {
      console.error('Failed to fetch hosts', err);
    }
  };

  const fetchPorts = async () => {
    try {
      const response = await api.get('/ports');
      setPorts(response.data);
    } catch (err) {
      console.error('Failed to fetch ports', err);
    }
  };

  const fetchURLs = async () => {
    try {
      const response = await api.get('/urls');
      setUrls(response.data);
    } catch (err) {
      console.error('Failed to fetch URLs', err);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploading(true);
    setUploadError('');

    const formData = new FormData();
    formData.append('file', file);

    try {
      // Don't set Content-Type header - axios will set it automatically with boundary for FormData
      await api.post('/scan-files', formData);
      // Refresh all data after upload
      await fetchAllData();
      e.target.value = ''; // Reset input
      setUploadError(''); // Clear any previous errors
    } catch (err: any) {
      console.error('Upload error:', err);
      console.error('Error response:', err.response);
      const errorMsg = err.response?.data?.error || 
                      (typeof err.response?.data === 'string' ? err.response.data : '') ||
                      err.message || 
                      'Failed to upload file';
      setUploadError(typeof errorMsg === 'string' ? errorMsg : JSON.stringify(errorMsg));
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="dashboard-page">
      <div className="container">
        <div className="dashboard-content">
          <h2>Pentest Engagement Dashboard</h2>
          <p className="dashboard-subtitle">
            All scan data is shared across all team members in this engagement.
          </p>

          {hosts.length > 0 && (
            <div className="dashboard-review-progress">
              <div className="progress-label">
                <span>Hosts reviewed</span>
                <span>{hosts.filter((h) => h.reviewed).length} / {hosts.length}</span>
              </div>
              <div className="progress-bar">
                <div
                  className="progress-fill"
                  style={{ width: `${hosts.length ? (100 * hosts.filter((h) => h.reviewed).length) / hosts.length : 0}%` }}
                />
              </div>
            </div>
          )}

          {needsAttention.length > 0 && (
            <div className="dashboard-needs-attention card">
              <h3>Needs attention</h3>
              <p className="needs-attention-desc">Critical or high severity findings still open</p>
              <ul>
                {needsAttention.map((f) => (
                  <li key={f.id}>
                    <span className={`severity-badge severity-${f.severity}`}>{f.severity}</span>
                    <button type="button" className="needs-attention-link" onClick={() => navigate('/findings')}>{f.title}</button>
                  </li>
                ))}
              </ul>
              <button type="button" className="btn btn-primary btn-sm" onClick={() => navigate('/findings')}>
                View all findings
              </button>
            </div>
          )}

          <div className="dashboard-grid">
            <div className="dashboard-card clickable-card" onClick={() => navigate('/hosts')}>
              <h3>Hosts</h3>
              <p className="stat-number">{hosts.length}</p>
              <p className="stat-label">Total hosts discovered</p>
              <button className="btn btn-primary" onClick={(e) => { e.stopPropagation(); navigate('/hosts'); }}>
                View Hosts
              </button>
            </div>

            <div className="dashboard-card clickable-card" onClick={() => navigate('/ports')}>
              <h3>Ports</h3>
              <p className="stat-number">{ports.length}</p>
              <p className="stat-label">Open ports found</p>
              <button className="btn btn-primary" onClick={(e) => { e.stopPropagation(); navigate('/ports'); }}>
                View Ports
              </button>
            </div>

            <div className="dashboard-card clickable-card" onClick={() => navigate('/urls')}>
              <h3>URLs</h3>
              <p className="stat-number">{urls.length}</p>
              <p className="stat-label">Web paths discovered</p>
              <button className="btn btn-primary" onClick={(e) => { e.stopPropagation(); navigate('/urls'); }}>
                View URLs
              </button>
            </div>

            <div className="dashboard-card clickable-card" onClick={() => navigate('/findings')}>
              <h3>Findings</h3>
              <p className="stat-number">{findingsCount}</p>
              {findingsSummary && (
                <div className="dashboard-findings-breakdown">
                  {['critical', 'high', 'medium', 'low'].map((sev) => ((findingsSummary[sev] ?? 0) > 0 ? (
                    <span key={sev} className={`breakdown-item severity-${sev}`}>{sev}: {findingsSummary[sev]}</span>
                  ) : null))}
                </div>
              )}
              <p className="stat-label">
                {openFindingsCount > 0 ? `${openFindingsCount} open` : 'Vulnerabilities &amp; findings'}
              </p>
              <button className="btn btn-primary" onClick={(e) => { e.stopPropagation(); navigate('/findings'); }}>
                View Findings
              </button>
            </div>

            <div className="dashboard-card">
              <h3>Scan Files</h3>
              <p className="stat-number">{scanFiles.length}</p>
              <p className="stat-label">Files uploaded</p>
              <label className="btn btn-primary" style={{ cursor: uploading ? 'not-allowed' : 'pointer' }}>
                {uploading ? 'Uploading...' : 'Upload Files'}
                <input
                  type="file"
                  accept=".xml,.json,.csv,.yaml,.yml"
                  onChange={handleFileUpload}
                  disabled={uploading}
                  style={{ display: 'none' }}
                />
              </label>
            </div>
          </div>

          {uploadError && (
            <div className="error-message" style={{ marginBottom: '20px' }}>
              {uploadError}
            </div>
          )}

          <div className="dashboard-section">
            <h3>Uploaded Scan Files</h3>
            {loading ? (
              <div className="card">Loading...</div>
            ) : scanFiles.length === 0 ? (
              <div className="card">
                <p>No scan files uploaded yet. Use the "Upload Files" button above to get started.</p>
              </div>
            ) : (
              <div className="card">
                <table className="scan-files-table">
                  <thead>
                    <tr>
                      <th>Filename</th>
                      <th>Type</th>
                      <th>Status</th>
                      <th>Uploaded</th>
                    </tr>
                  </thead>
                  <tbody>
                    {scanFiles.map((file) => (
                      <tr key={file.id}>
                        <td>{file.filename}</td>
                        <td>
                          <span className={`file-type-badge file-type-${file.type}`}>
                            {file.type.replace('_', ' ')}
                          </span>
                        </td>
                        <td>
                          <span className={`status-badge status-${file.status}`}>
                            {file.status}
                          </span>
                        </td>
                        <td>{new Date(file.uploaded_at).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div className="dashboard-section">
            <h3>Getting Started</h3>
            <div className="info-card">
              <p>To begin analyzing scan data:</p>
              <ul>
                <li>Upload nmap XML files to discover hosts and ports</li>
                <li>Upload ffuf JSON/CSV files to discover web paths</li>
                <li>Upload Postman collections (JSON) or OpenAPI/Swagger specs (JSON or YAML) to import API endpoints</li>
                <li>Click on the cards above to view detailed lists</li>
                <li>Filter by reviewed/unreviewed status</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
