import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api, { setURLReviewed } from '../services/api';
import './URLs.css';

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
  host_id?: string;
  host?: string;
}

const URLs: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [urls, setUrls] = useState<URL[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'reviewed' | 'unreviewed'>('all');
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusInput, setStatusInput] = useState('');
  const [statusQuery, setStatusQuery] = useState('');
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const fetchURLs = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (searchQuery.trim()) params.set('search', searchQuery.trim());
      if (filter === 'reviewed') params.set('reviewed', 'true');
      if (filter === 'unreviewed') params.set('reviewed', 'false');
      if (statusQuery.trim()) {
        const code = parseInt(statusQuery.trim(), 10);
        if (!Number.isNaN(code)) params.set('status', String(code));
      }
      const url = params.toString() ? `/urls?${params.toString()}` : '/urls';
      const response = await api.get(url);
      setUrls(Array.isArray(response.data) ? response.data : []);
    } catch (err) {
      console.error('Failed to fetch URLs', err);
      setUrls([]);
    } finally {
      setLoading(false);
    }
  }, [searchQuery, filter, statusQuery]);

  useEffect(() => {
    fetchURLs();
  }, [fetchURLs]);

  useEffect(() => {
    const t = setTimeout(() => setSearchQuery(searchInput), 350);
    return () => clearTimeout(t);
  }, [searchInput]);

  useEffect(() => {
    const t = setTimeout(() => setStatusQuery(statusInput), 350);
    return () => clearTimeout(t);
  }, [statusInput]);

  const filteredURLs = urls;

  const handleToggleReviewed = async (urlObj: URL) => {
    if (togglingId === urlObj.id) return;
    setTogglingId(urlObj.id);
    try {
      const updated = await setURLReviewed(urlObj.id, !urlObj.reviewed);
      setUrls(prev => prev.map(u => (u.id === urlObj.id ? { ...u, reviewed: updated.reviewed } : u)));
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

  return (
    <div className="urls-page">
      <header className="urls-header">
        <div className="container">
          <div className="header-content">
            <h1>Edda - URLs</h1>
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
        <div className="urls-content">
          <div className="page-header">
            <h2>Discovered URLs</h2>
            <div className="list-toolbar">
              <input
                type="search"
                className="search-input"
                placeholder="Search URL or path..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                aria-label="Search URLs"
              />
              <input
                type="text"
                className="status-filter-input"
                placeholder="Status (e.g. 200, 404)"
                value={statusInput}
                onChange={(e) => setStatusInput(e.target.value)}
                aria-label="Filter by status code"
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
          ) : filteredURLs.length === 0 ? (
            <div className="card">
              <p>{searchQuery || statusQuery || filter !== 'all' ? 'No URLs match your search or filter.' : 'No URLs found.'}</p>
            </div>
          ) : (
            <div className="card">
              <table className="urls-table">
                <thead>
                  <tr>
                    <th>Host</th>
                    <th>URL</th>
                    <th>Method</th>
                    <th>Status</th>
                    <th>Length</th>
                    <th>Words</th>
                    <th>Lines</th>
                    <th>Review Status</th>
                    <th>Discovered</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredURLs.map((urlObj) => (
                    <tr key={urlObj.id}>
                      <td className="host-cell">
                        {(urlObj.host ?? '') ? (
                          urlObj.host_id ? (
                            <button
                              type="button"
                              className="host-link"
                              onClick={() => navigate(`/hosts/${urlObj.host_id}`)}
                            >
                              {urlObj.host}
                            </button>
                          ) : (
                            urlObj.host
                          )
                        ) : (
                          '-'
                        )}
                      </td>
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
                        <button
                          type="button"
                          className={`status-badge status-badge-btn ${urlObj.reviewed ? 'reviewed' : 'unreviewed'}`}
                          onClick={() => handleToggleReviewed(urlObj)}
                          disabled={togglingId === urlObj.id}
                          title={urlObj.reviewed ? 'Mark unreviewed' : 'Mark reviewed'}
                        >
                          {togglingId === urlObj.id ? '…' : urlObj.reviewed ? 'Reviewed' : 'Unreviewed'}
                        </button>
                      </td>
                      <td>{new Date(urlObj.created_at).toLocaleString()}</td>
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

export default URLs;
