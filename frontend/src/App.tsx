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
import PrivateRoute from './components/PrivateRoute';
import './App.css';

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="App">
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route
              path="/dashboard"
              element={
                <PrivateRoute>
                  <Dashboard />
                </PrivateRoute>
              }
            />
            <Route
              path="/admin"
              element={
                <PrivateRoute>
                  <Admin />
                </PrivateRoute>
              }
            />
            <Route
              path="/hosts"
              element={
                <PrivateRoute>
                  <Hosts />
                </PrivateRoute>
              }
            />
            <Route
              path="/hosts/:id"
              element={
                <PrivateRoute>
                  <HostDetail />
                </PrivateRoute>
              }
            />
            <Route
              path="/ports"
              element={
                <PrivateRoute>
                  <Ports />
                </PrivateRoute>
              }
            />
            <Route
              path="/ports/by-number/:port/:protocol"
              element={
                <PrivateRoute>
                  <PortDetail />
                </PrivateRoute>
              }
            />
            <Route
              path="/urls"
              element={
                <PrivateRoute>
                  <URLs />
                </PrivateRoute>
              }
            />
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
