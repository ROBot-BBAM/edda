import React from 'react';
import './RowActionIcons.css';

const IconNote = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
  </svg>
);

const IconFinding = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <circle cx="12" cy="12" r="10" />
    <line x1="12" y1="8" x2="12" y2="16" />
    <line x1="8" y1="12" x2="16" y2="12" />
  </svg>
);

const IconView = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
    <circle cx="12" cy="12" r="3" />
  </svg>
);

export interface RowActionIconsProps {
  onAddNote?: () => void;
  onAddFinding?: () => void;
  onView?: () => void;
  viewLabel?: string;
}

export default function RowActionIcons({ onAddNote, onAddFinding, onView, viewLabel = 'View Details' }: RowActionIconsProps) {
  return (
    <div className="row-actions">
      {onAddNote != null && (
        <button type="button" className="row-action-icon" onClick={onAddNote} title="Add Note" aria-label="Add Note">
          <IconNote />
        </button>
      )}
      {onAddFinding != null && (
        <button type="button" className="row-action-icon" onClick={onAddFinding} title="Add Finding" aria-label="Add Finding">
          <IconFinding />
        </button>
      )}
      {onView != null && (
        <button type="button" className="row-action-icon" onClick={onView} title={viewLabel} aria-label={viewLabel}>
          <IconView />
        </button>
      )}
    </div>
  );
}
