import React, { useState, useEffect, useRef, useCallback } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { search as searchApi, downloadNarrativeExport, type SearchResult } from '../services/api';
import './AppLayout.css';

export default function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [exportOpen, setExportOpen] = useState(false);
  const [exporting, setExporting] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);
  const exportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const close = (e: MouseEvent) => {
      if (exportRef.current && !exportRef.current.contains(e.target as Node)) setExportOpen(false);
    };
    document.addEventListener('click', close);
    return () => document.removeEventListener('click', close);
  }, []);

  const handleExportFormat = async (format: 'json' | 'csv') => {
    setExporting(true);
    try {
      await downloadNarrativeExport(format);
      setExportOpen(false);
    } catch {
      // Could show a toast; for now no-op
    } finally {
      setExporting(false);
    }
  };

  const runSearch = useCallback(async (q: string) => {
    if (q.length < 2) {
      setSearchResults(null);
      return;
    }
    setSearchLoading(true);
    try {
      const res = await searchApi(q, 8);
      setSearchResults(res);
      setSearchOpen(true);
    } catch {
      setSearchResults(null);
    } finally {
      setSearchLoading(false);
    }
  }, []);

  useEffect(() => {
    if (searchQuery.length < 2) {
      setSearchResults(null);
      setSearchOpen(false);
      return;
    }
    const t = setTimeout(() => runSearch(searchQuery), 300);
    return () => clearTimeout(t);
  }, [searchQuery, runSearch]);

  useEffect(() => {
    const onBlur = () => {
      setTimeout(() => setSearchOpen(false), 150);
    };
    const el = searchRef.current;
    el?.addEventListener('focusout', onBlur);
    return () => el?.removeEventListener('focusout', onBlur);
  }, []);

  const totalResults = searchResults
    ? (searchResults.hosts?.length ?? 0) + (searchResults.urls?.length ?? 0) + (searchResults.findings?.length ?? 0)
    : 0;
  const showDropdown = searchOpen && searchQuery.length >= 2 && !searchLoading && (totalResults > 0 || (searchResults !== null && totalResults === 0));

  return (
    <div className="app-layout">
      <header className="app-nav">
        <div className="app-nav-inner">
          <NavLink to="/dashboard" className="app-nav-logo">
            Edda
          </NavLink>
          <div className="app-nav-search" ref={searchRef}>
            <input
              type="search"
              className="app-nav-search-input"
              placeholder="Search hosts, URLs, findings..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onFocus={() => searchQuery.length >= 2 && setSearchOpen(true)}
              aria-label="Search"
            />
            {searchLoading && <span className="app-nav-search-spinner" aria-hidden />}
            {showDropdown && (
              <div className="app-nav-search-dropdown">
                {searchResults?.hosts?.length ? (
                  <div className="search-group">
                    <div className="search-group-label">Hosts</div>
                    {searchResults.hosts.map((h) => (
                      <button
                        key={h.id}
                        type="button"
                        className="search-hit"
                        onClick={() => { navigate(`/hosts/${h.id}`); setSearchOpen(false); setSearchQuery(''); }}
                      >
                        {h.label}
                      </button>
                    ))}
                  </div>
                ) : null}
                {searchResults?.urls?.length ? (
                  <div className="search-group">
                    <div className="search-group-label">URLs</div>
                    {searchResults.urls.map((u) => (
                      <button
                        key={u.id}
                        type="button"
                        className="search-hit"
                        onClick={() => { navigate('/urls'); setSearchOpen(false); setSearchQuery(''); }}
                      >
                        {u.label}
                      </button>
                    ))}
                  </div>
                ) : null}
                {searchResults?.findings?.length ? (
                  <div className="search-group">
                    <div className="search-group-label">Findings</div>
                    {searchResults.findings.map((f) => (
                      <button
                        key={f.id}
                        type="button"
                        className="search-hit"
                        onClick={() => { navigate('/findings'); setSearchOpen(false); setSearchQuery(''); }}
                      >
                        <span className={`search-hit-severity severity-${f.severity}`}>{f.severity}</span>
                        {f.title}
                      </button>
                    ))}
                  </div>
                ) : null}
                {totalResults === 0 && searchResults !== null && (
                  <div className="search-group">
                    <div className="search-empty">No results</div>
                  </div>
                )}
              </div>
            )}
          </div>
          <nav className="app-nav-links">
            <NavLink to="/dashboard" className={({ isActive }) => `app-nav-link ${isActive ? 'active' : ''}`} end>
              Dashboard
            </NavLink>
            <NavLink to="/hosts" className={({ isActive }) => `app-nav-link ${isActive ? 'active' : ''}`}>
              Hosts
            </NavLink>
            <NavLink to="/ports" className={({ isActive }) => `app-nav-link ${isActive ? 'active' : ''}`}>
              Ports
            </NavLink>
            <NavLink to="/urls" className={({ isActive }) => `app-nav-link ${isActive ? 'active' : ''}`}>
              URLs
            </NavLink>
            <NavLink to="/findings" className={({ isActive }) => `app-nav-link ${isActive ? 'active' : ''}`}>
              Findings
            </NavLink>
          </nav>
          <div className="app-nav-right">
            <div className="app-nav-export-dropdown" ref={exportRef}>
              <button
                type="button"
                className="app-nav-export app-nav-export-trigger"
                onClick={() => setExportOpen((o) => !o)}
                disabled={exporting}
                title="Export all hosts, ports, URLs, findings, and notes"
              >
                {exporting ? 'Exporting…' : 'Export All'}
                <span className="export-caret" aria-hidden>▾</span>
              </button>
              {exportOpen && (
                <div className="app-nav-export-menu">
                  <button type="button" className="app-nav-export-item" onClick={() => handleExportFormat('json')}>JSON</button>
                  <button type="button" className="app-nav-export-item" onClick={() => handleExportFormat('csv')}>CSV</button>
                </div>
              )}
            </div>
            <span className="app-nav-user">{user?.email}</span>
            {user?.is_admin && (
              <NavLink to="/admin" className={({ isActive }) => `app-nav-link app-nav-admin ${isActive ? 'active' : ''}`}>
                Admin
              </NavLink>
            )}
            <button type="button" className="app-nav-logout" onClick={logout}>
              Logout
            </button>
          </div>
        </div>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
