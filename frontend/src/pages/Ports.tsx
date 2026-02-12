import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import './Ports.css';

interface PortAggregate {
  port: number;
  protocol: string;
  state?: string;
  service_name?: string;
  service_product?: string;
  service_version?: string;
  reviewed: boolean;
  host_count: number;
}

const Ports: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [ports, setPorts] = useState<PortAggregate[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'reviewed' | 'unreviewed'>('all');
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const fetchPorts = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (searchQuery.trim()) params.set('search', searchQuery.trim());
      if (filter === 'reviewed') params.set('reviewed', 'true');
      if (filter === 'unreviewed') params.set('reviewed', 'false');
      const url = params.toString() ? `/ports?${params.toString()}` : '/ports';
      const response = await api.get(url);
      setPorts(response.data);
    } catch (err) {
      console.error('Failed to fetch ports', err);
    } finally {
      setLoading(false);
    }
  }, [searchQuery, filter]);

  useEffect(() => {
    fetchPorts();
  }, [fetchPorts]);

  useEffect(() => {
    const t = setTimeout(() => setSearchQuery(searchInput), 350);
    return () => clearTimeout(t);
  }, [searchInput]);

  const filteredPorts = ports;

  const viewDetailsUrl = (p: PortAggregate) => `/ports/by-number/${p.port}/${p.protocol}`;

  return (
    <div className="ports-page">
      <header className="ports-header">
        <div className="container">
          <div className="header-content">
            <h1>Edda - Ports</h1>
            <div className="header-actions">
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
        <div className="ports-content">
          <div className="page-header">
            <h2>Open Ports</h2>
            <div className="list-toolbar">
              <input
                type="search"
                className="search-input"
                placeholder="Search port or service..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                aria-label="Search ports"
              />
              <div className="filter-buttons">
                <button
                  className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
                  onClick={() => setFilter('all')}
                >
                  All
                </button>
                <button
                  className={`filter-btn ${filter === 'reviewed' ? 'active' : ''}`}
                  onClick={() => setFilter('reviewed')}
                >
                  Reviewed
                </button>
                <button
                  className={`filter-btn ${filter === 'unreviewed' ? 'active' : ''}`}
                  onClick={() => setFilter('unreviewed')}
                >
                  Unreviewed
                </button>
              </div>
            </div>
          </div>

          {loading ? (
            <div className="card">Loading...</div>
          ) : filteredPorts.length === 0 ? (
            <div className="card">
              <p>{searchQuery || filter !== 'all' ? 'No ports match your search or filter.' : 'No ports found.'}</p>
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
                    <th>Status</th>
                    <th>Hosts</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredPorts.map((p) => (
                    <tr key={`${p.port}-${p.protocol}`}>
                      <td className="port-number">{p.port}</td>
                      <td>{p.protocol.toUpperCase()}</td>
                      <td>{p.state || '-'}</td>
                      <td>{p.service_name || '-'}</td>
                      <td>{p.service_product || '-'}</td>
                      <td>{p.service_version || '-'}</td>
                      <td>
                        <span className={`status-badge ${p.reviewed ? 'reviewed' : 'unreviewed'}`}>
                          {p.reviewed ? 'Reviewed' : 'Unreviewed'}
                        </span>
                      </td>
                      <td>{p.host_count}</td>
                      <td>
                        <button
                          className="btn btn-sm btn-primary"
                          onClick={() => navigate(viewDetailsUrl(p))}
                        >
                          View Details
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

export default Ports;
