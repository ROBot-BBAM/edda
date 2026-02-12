import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api, { setHostReviewed } from '../services/api';
import './Hosts.css';

interface Host {
  id: string;
  ip_address: string;
  hostname?: string;
  os?: string;
  reviewed: boolean;
  created_at: string;
}

const Hosts: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [hosts, setHosts] = useState<Host[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'reviewed' | 'unreviewed'>('all');
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const fetchHosts = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (searchQuery.trim()) params.set('search', searchQuery.trim());
      if (filter === 'reviewed') params.set('reviewed', 'true');
      if (filter === 'unreviewed') params.set('reviewed', 'false');
      const url = params.toString() ? `/hosts?${params.toString()}` : '/hosts';
      const response = await api.get(url);
      setHosts(response.data);
    } catch (err) {
      console.error('Failed to fetch hosts', err);
    } finally {
      setLoading(false);
    }
  }, [searchQuery, filter]);

  useEffect(() => {
    fetchHosts();
  }, [fetchHosts]);

  useEffect(() => {
    const t = setTimeout(() => setSearchQuery(searchInput), 350);
    return () => clearTimeout(t);
  }, [searchInput]);

  const filteredHosts = hosts;

  const handleToggleReviewed = async (host: Host) => {
    if (togglingId === host.id) return;
    setTogglingId(host.id);
    try {
      const updated = await setHostReviewed(host.id, !host.reviewed);
      setHosts(prev => prev.map(h => (h.id === host.id ? { ...h, reviewed: updated.reviewed } : h)));
    } catch (err) {
      console.error('Failed to update host', err);
    } finally {
      setTogglingId(null);
    }
  };

  return (
    <div className="hosts-page">
      <header className="hosts-header">
        <div className="container">
          <div className="header-content">
            <h1>Edda - Hosts</h1>
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
        <div className="hosts-content">
          <div className="page-header">
            <h2>Discovered Hosts</h2>
            <div className="list-toolbar">
              <input
                type="search"
                className="search-input"
                placeholder="Search IP or hostname..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                aria-label="Search hosts"
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
          ) : filteredHosts.length === 0 ? (
            <div className="card">
              <p>{searchQuery || filter !== 'all' ? 'No hosts match your search or filter.' : 'No hosts found.'}</p>
            </div>
          ) : (
            <div className="card">
              <table className="hosts-table">
                <thead>
                  <tr>
                    <th>IP Address</th>
                    <th>Hostname</th>
                    <th>OS</th>
                    <th>Status</th>
                    <th>Discovered</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredHosts.map((host) => (
                    <tr key={host.id}>
                      <td className="ip-address">{host.ip_address}</td>
                      <td>{host.hostname || '-'}</td>
                      <td>{host.os || '-'}</td>
                      <td>
                        <button
                          type="button"
                          className={`status-badge status-badge-btn ${host.reviewed ? 'reviewed' : 'unreviewed'}`}
                          onClick={() => handleToggleReviewed(host)}
                          disabled={togglingId === host.id}
                          title={host.reviewed ? 'Mark unreviewed' : 'Mark reviewed'}
                        >
                          {togglingId === host.id ? '…' : host.reviewed ? 'Reviewed' : 'Unreviewed'}
                        </button>
                      </td>
                      <td>{new Date(host.created_at).toLocaleString()}</td>
                      <td>
                        <button
                          className="btn btn-sm btn-primary"
                          onClick={() => navigate(`/hosts/${host.id}`)}
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

export default Hosts;
