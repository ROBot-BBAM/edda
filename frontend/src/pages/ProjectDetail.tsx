import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import './ProjectDetail.css';

interface Project {
  id: string;
  name: string;
  description: string | null;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

const ProjectDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const [project, setProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (id) {
      fetchProject();
    }
  }, [id]);

  const fetchProject = async () => {
    try {
      const response = await api.get(`/projects/${id}`);
      setProject(response.data);
    } catch (err) {
      console.error('Failed to fetch project', err);
      navigate('/projects');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="container">Loading...</div>;
  }

  if (!project) {
    return <div className="container">Project not found</div>;
  }

  return (
    <div className="project-detail-page">
      <header className="projects-header">
        <div className="container">
          <div className="header-content">
            <h1>Edda</h1>
            <div className="header-actions">
              <button onClick={() => navigate('/projects')} className="btn btn-secondary">
                Back to Projects
              </button>
              <button onClick={logout} className="btn btn-secondary">
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      <div className="container">
        <div className="project-header">
          <h2>{project.name}</h2>
          {project.description && <p>{project.description}</p>}
        </div>

        <div className="card">
          <h3>Project Overview</h3>
          <p>Scan file upload and review features will be implemented here.</p>
          <p>This will include:</p>
          <ul>
            <li>Upload nmap XML files</li>
            <li>Upload ffuf JSON/CSV files</li>
            <li>View parsed hosts, ports, and URLs</li>
            <li>Mark items as reviewed</li>
            <li>Filter by reviewed/unreviewed status</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default ProjectDetail;
