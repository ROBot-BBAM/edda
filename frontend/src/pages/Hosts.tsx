import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import NotesModal from '../components/NotesModal';
import RowActionIcons from '../components/RowActionIcons';
import './Hosts.css';

interface Host {
  id: string;
  ip_address: string;
  hostname?: string;
  os?: string;
  reviewed: boolean;
  finding_count: number;
  note_count: number;
  note_preview: string;
  created_at: string;
}

const Hosts: React.FC = () => {
  const navigate = useNavigate();
  const [hosts, setHosts] = useState<Host[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'reviewed' | 'unreviewed'>('all');
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [notesModalHostId, setNotesModalHostId] = useState<string | null>(null);

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

  return (
    <div className="hosts-page">
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
                    <th>Findings</th>
                    <th>Notes</th>
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
                        <span
                          className={`status-badge ${host.reviewed ? 'reviewed' : 'unreviewed'}`}
                          aria-label={`Status: ${host.reviewed ? 'Reviewed' : 'Unreviewed'}`}
                        >
                          {host.reviewed ? 'Reviewed' : 'Unreviewed'}
                        </span>
                      </td>
                      <td>
                        <span className={`finding-badge ${host.finding_count > 0 ? 'has-findings' : ''}`} title={`${host.finding_count} finding(s)`}>
                          {host.finding_count > 0 ? host.finding_count : '—'}
                        </span>
                      </td>
                      <td title={host.note_preview || undefined}>
                        <span className={`note-badge ${(host.note_count || 0) > 0 ? 'has-notes' : ''}`}>
                          {(host.note_count || 0) > 0 ? host.note_count : '—'}
                        </span>
                      </td>
                      <td>{new Date(host.created_at).toLocaleString()}</td>
                      <td className="row-actions-cell">
                        <RowActionIcons
                          onAddNote={() => setNotesModalHostId(host.id)}
                          onAddFinding={() => navigate(`/findings?host_id=${host.id}`)}
                          onView={() => navigate(`/hosts/${host.id}`)}
                          viewLabel="View Details"
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          {notesModalHostId && (
            <NotesModal
              entityType="host"
              entityId={notesModalHostId}
              onClose={() => setNotesModalHostId(null)}
              onAdded={fetchHosts}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default Hosts;
