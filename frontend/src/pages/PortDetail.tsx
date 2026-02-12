import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api, { setPortReviewed } from '../services/api';
import './PortDetail.css';

interface PortInfo {
  port: number;
  protocol: string;
  state?: string;
  service_name?: string;
  service_product?: string;
  service_version?: string;
  reviewed: boolean;
}

interface HostPortRow {
  host: {
    id: string;
    ip_address: string;
    hostname?: string;
    os?: string;
    reviewed: boolean;
    created_at: string;
  };
  port_id: string;
  port_reviewed: boolean;
}

interface PortDetailData {
  port: PortInfo;
  hosts: HostPortRow[];
}

const PortDetail: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const { port: portParam, protocol: protocolParam } = useParams<{ port: string; protocol: string }>();
  const [data, setData] = useState<PortDetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  useEffect(() => {
    if (portParam && protocolParam) {
      fetchPortDetail(portParam, protocolParam);
    }
  }, [portParam, protocolParam]);

  const fetchPortDetail = async (port: string, protocol: string) => {
    try {
      const response = await api.get(`/ports/by-number/${port}/${protocol}`);
      setData(response.data);
    } catch (err) {
      console.error('Failed to fetch port detail', err);
    } finally {
      setLoading(false);
    }
  };

  const handleTogglePortReviewed = async (row: HostPortRow) => {
    if (togglingId === row.port_id) return;
    setTogglingId(row.port_id);
    try {
      const updated = await setPortReviewed(row.port_id, !row.port_reviewed);
      setData(d => {
        if (!d) return null;
        return {
          ...d,
          hosts: d.hosts.map(h => (h.port_id === row.port_id ? { ...h, port_reviewed: updated.reviewed } : h)),
          port: {
            ...d.port,
            reviewed: d.hosts.every(h => (h.port_id === row.port_id ? updated.reviewed : h.port_reviewed)),
          },
        };
      });
    } catch (err) {
      console.error('Failed to update port', err);
    } finally {
      setTogglingId(null);
    }
  };

  if (loading) {
    return (
      <div className="port-detail-page">
        <div className="container">
          <div className="card">Loading...</div>
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="port-detail-page">
        <div className="container">
          <div className="card">Port not found</div>
        </div>
      </div>
    );
  }

  return (
    <div className="port-detail-page">
      <header className="port-detail-header">
        <div className="container">
          <div className="header-content">
            <h1>Edda - Port Detail</h1>
            <div className="header-actions">
              <button onClick={() => navigate('/ports')} className="btn btn-secondary">
                Back to Ports
              </button>
              <button onClick={() => navigate('/dashboard')} className="btn btn-secondary">
                Dashboard
              </button>
              {user?.is_admin && (
                <button onClick={() => navigate('/admin')} className="btn btn-secondary">
                  Admin
                </button>
              )}
              <span>Welcome, {user?.email}</span>
              <button onClick={logout} className="btn btn-secondary">
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      <div className="container">
        <div className="port-detail-content">
          <div className="port-info-card card">
            <h2>Port Information</h2>
            <div className="info-grid">
              <div className="info-item">
                <label>Port:</label>
                <span className="port-number">{data.port.port}</span>
              </div>
              <div className="info-item">
                <label>Protocol:</label>
                <span>{data.port.protocol.toUpperCase()}</span>
              </div>
              {data.port.state && (
                <div className="info-item">
                  <label>State:</label>
                  <span>{data.port.state}</span>
                </div>
              )}
              {data.port.service_name && (
                <div className="info-item">
                  <label>Service:</label>
                  <span>{data.port.service_name}</span>
                </div>
              )}
              {data.port.service_product && (
                <div className="info-item">
                  <label>Product:</label>
                  <span>{data.port.service_product}</span>
                </div>
              )}
              {data.port.service_version && (
                <div className="info-item">
                  <label>Version:</label>
                  <span>{data.port.service_version}</span>
                </div>
              )}
              <div className="info-item">
                <label>Status:</label>
                <span className={`status-badge ${data.port.reviewed ? 'reviewed' : 'unreviewed'}`}>
                  {data.port.reviewed ? 'Reviewed' : 'Unreviewed'}
                </span>
                <span className="info-hint"> (all hosts must be marked reviewed)</span>
              </div>
            </div>
          </div>

          <div className="section">
            <h3>Hosts with Port {data.port.port}/{data.port.protocol.toUpperCase()} ({data.hosts.length})</h3>
            {data.hosts.length === 0 ? (
              <div className="card">
                <p>No hosts found with this port.</p>
              </div>
            ) : (
              <div className="card">
                <table className="hosts-table">
                  <thead>
                    <tr>
                      <th>IP Address</th>
                      <th>Hostname</th>
                      <th>OS</th>
                      <th>Port reviewed</th>
                      <th>Discovered</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.hosts.map((row) => (
                      <tr key={row.port_id}>
                        <td className="ip-address">{row.host.ip_address}</td>
                        <td>{row.host.hostname || '-'}</td>
                        <td>{row.host.os || '-'}</td>
                        <td>
                          <button
                            type="button"
                            className={`status-badge status-badge-btn ${row.port_reviewed ? 'reviewed' : 'unreviewed'}`}
                            onClick={() => handleTogglePortReviewed(row)}
                            disabled={togglingId === row.port_id}
                            title={row.port_reviewed ? 'Mark unreviewed' : 'Mark reviewed'}
                          >
                            {togglingId === row.port_id ? '…' : row.port_reviewed ? 'Reviewed' : 'Unreviewed'}
                          </button>
                        </td>
                        <td>{new Date(row.host.created_at).toLocaleString()}</td>
                        <td>
                          <button
                            className="btn btn-sm btn-primary"
                            onClick={() => navigate(`/hosts/${row.host.id}`)}
                          >
                            View Host
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
    </div>
  );
};

export default PortDetail;
