import React, { useState, useEffect } from 'react';
import { createNote, listNotesByHost, listNotesByPort, listNotesByURL, type Note } from '../services/api';
import './NotesModal.css';

type EntityType = 'host' | 'port' | 'url';

interface NotesModalProps {
  entityType: EntityType;
  entityId: string;
  onClose: () => void;
  onAdded?: () => void;
}

const NotesModal: React.FC<NotesModalProps> = ({ entityType, entityId, onClose, onAdded }) => {
  const [notes, setNotes] = useState<Note[]>([]);
  const [loading, setLoading] = useState(true);
  const [content, setContent] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const fetchNotes = async () => {
      try {
        let list: Note[] = [];
        if (entityType === 'host') list = await listNotesByHost(entityId);
        else if (entityType === 'port') list = await listNotesByPort(entityId);
        else list = await listNotesByURL(entityId);
        setNotes(Array.isArray(list) ? list : []);
      } catch (e) {
        console.error('Failed to fetch notes', e);
      } finally {
        setLoading(false);
      }
    };
    fetchNotes();
  }, [entityType, entityId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = content.trim();
    if (!trimmed) return;
    setSubmitting(true);
    try {
      const body: { host_id?: string; port_id?: string; url_id?: string; content: string } = { content: trimmed };
      if (entityType === 'host') body.host_id = entityId;
      else if (entityType === 'port') body.port_id = entityId;
      else body.url_id = entityId;
      await createNote(body);
      setContent('');
      const list =
        entityType === 'host'
          ? await listNotesByHost(entityId)
          : entityType === 'port'
            ? await listNotesByPort(entityId)
            : await listNotesByURL(entityId);
      setNotes(Array.isArray(list) ? list : []);
      onAdded?.();
    } catch (err) {
      console.error('Failed to add note', err);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="notes-modal-overlay" onClick={onClose}>
      <div className="notes-modal-content card" onClick={(e) => e.stopPropagation()}>
        <div className="notes-modal-header">
          <h3>Notes</h3>
          <button type="button" className="btn btn-sm btn-secondary" onClick={onClose}>
            Close
          </button>
        </div>
        {loading ? (
          <p>Loading...</p>
        ) : (
          <>
            <ul className="notes-list">
              {notes.length === 0 ? (
                <li className="notes-empty">No notes yet.</li>
              ) : (
                notes.map((n) => (
                  <li key={n.id} className="note-item">
                    <p className="note-content">{n.content}</p>
                    <span className="note-meta">{new Date(n.created_at).toLocaleString()}</span>
                  </li>
                ))
              )}
            </ul>
            <form onSubmit={handleSubmit} className="notes-add-form">
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Add a note..."
                rows={3}
                className="notes-textarea"
              />
              <button type="submit" className="btn btn-primary btn-sm" disabled={submitting || !content.trim()}>
                {submitting ? 'Adding…' : 'Add note'}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
};

export default NotesModal;
