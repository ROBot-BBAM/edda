import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Admin from './pages/Admin';
import Hosts from './pages/Hosts';
import HostDetail from './pages/HostDetail';
import Ports from './pages/Ports';
import PortDetail from './pages/PortDetail';
import URLs from './pages/URLs';
import Findings from './pages/Findings';
import PrivateRoute from './components/PrivateRoute';
import AppLayout from './components/AppLayout';
import './App.css';

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="App">
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/" element={<PrivateRoute><AppLayout /></PrivateRoute>}>
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="admin" element={<Admin />} />
              <Route path="hosts" element={<Hosts />} />
              <Route path="hosts/:id" element={<HostDetail />} />
              <Route path="ports" element={<Ports />} />
              <Route path="ports/by-number/:port/:protocol" element={<PortDetail />} />
              <Route path="urls" element={<URLs />} />
              <Route path="findings" element={<Findings />} />
            </Route>
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
