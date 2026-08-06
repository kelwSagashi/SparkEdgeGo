import React from 'react'
import './App.css'
import { SidebarInset, SidebarProvider } from './components/ui/sidebar';
import SideBar from './components/sidebar';
import { Routes, Route, Navigate, } from 'react-router-dom'
import Login from './pages/Login';
import Register from './pages/Register';
import ProtectedRoute from './components/ProtectedRoute';
import { useEffect } from 'react';
import { useAuthStore } from './stores/auth-store';
import InstancesPage from './pages/Instances';
import InstanceCreatePage from './pages/InstanceCreatePage';
import ExecutionHistoryPage from './pages/ExecutionHistory';
import LocalDataPage from './pages/LocalData';
import { Toaster } from 'sonner';
import ScriptHubPage from './pages/ScriptHub';
import ScriptDetailsPage from './pages/ScriptDetailsPage';
import ScriptEditPage from './pages/ScriptEditPage';
import ServersPage from './pages/Servers';
import ServerCreatePage from './pages/ServerCreatePage';
import ServerEditPage from './pages/ServerEditPage';
import DevicesPage from './pages/Devices';
import SettingsPage from './pages/Settings';
import ProjectsPage from './pages/Projects';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import CredentialsPage from './pages/Credentials';
import DeviceCreationPage from './pages/DeviceCreatePage';
import DeviceEditPage from './pages/DeviceEditPage';
import InstanceEditPage from './pages/InstanceEditPage';
import InstanceLogsPage from './pages/InstanceLogsPage';
import CloudSettingsPage from './pages/CloudSettings';
import AdvancedSettingsPage from './pages/AdvancedSettings';

function App() {
  const loadMe = useAuthStore((s) => s.loadMe);
  const user = useAuthStore((s) => s.user);
  const loading = useAuthStore((s) => s.loading);

  useEffect(() => {
    loadMe();
  }, [loadMe]);

  // While checking session, show a minimal loading state
  if (loading) return <div className="h-screen flex items-center justify-center">Loading...</div>;

  // Unauthenticated layout: only show auth pages (no header/sidebar)
  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background text-foreground">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </div>
    );
  }

  // Authenticated layout
  return (
    <DndProvider backend={HTML5Backend}>
    <SidebarProvider
      className="flex flex-col h-screen bg-background text-foreground"
      defaultOpen={false}
      style={
        {
          "position": "relative",
          "--sidebar-width": "calc(var(--spacing) * 72)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as React.CSSProperties
      }
    >
      <div className="flex grow overflow-hidden">
        <SideBar />
        <SidebarInset className='flex flex-col grow relative overflow-hidden'>
          <div className="grow relative overflow-y-auto">
            <Routes>
              {/* Instances */}
              <Route path="/" element={<Navigate to="/instances" replace />} />
              <Route path="/instances" element={<ProtectedRoute><InstancesPage /></ProtectedRoute>} />
              <Route path="/instances/new" element={<ProtectedRoute><InstanceCreatePage /></ProtectedRoute>} />
              <Route path="/instances/:id" element={<ProtectedRoute><InstanceLogsPage /></ProtectedRoute>} />
              <Route path="/instances/:id/edit" element={<ProtectedRoute><InstanceEditPage /></ProtectedRoute>} />

              {/* Script Hub */}
              <Route path="/script-hub" element={<ProtectedRoute><ScriptHubPage /></ProtectedRoute>} />
              <Route path="/script-hub/:id" element={<ProtectedRoute><ScriptDetailsPage /></ProtectedRoute>} />
              <Route path="/script-hub/:id/edit" element={<ProtectedRoute><ScriptEditPage /></ProtectedRoute>} />

              {/* Local Data */}
              <Route path="/data" element={<ProtectedRoute><LocalDataPage /></ProtectedRoute>} />

              {/* Execution History */}
              <Route path="/history" element={<ProtectedRoute><ExecutionHistoryPage /></ProtectedRoute>} />

              {/* Management */}
              <Route path="/servers" element={<ProtectedRoute><ServersPage /></ProtectedRoute>} />
              <Route path="/servers/new" element={<ProtectedRoute><ServerCreatePage /></ProtectedRoute>} />
              <Route path="/servers/:id" element={<ProtectedRoute><ServerEditPage /></ProtectedRoute>} />
              <Route path="/credentials" element={<ProtectedRoute><CredentialsPage /></ProtectedRoute>} />
              <Route path="/devices" element={<ProtectedRoute><DevicesPage /></ProtectedRoute>} />
              <Route path="/devices/new" element={<ProtectedRoute><DeviceCreationPage /></ProtectedRoute>} />
              <Route path="/devices/:id" element={<ProtectedRoute><DeviceEditPage /></ProtectedRoute>} />
              <Route path="/projects" element={<ProtectedRoute><ProjectsPage /></ProtectedRoute>} />

              {/* Settings */}
              <Route path="/settings" element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
              <Route path="/settings/cloud" element={<ProtectedRoute><CloudSettingsPage /></ProtectedRoute>} />
              <Route path="/settings/advanced" element={<ProtectedRoute><AdvancedSettingsPage /></ProtectedRoute>} />

              {/* Legacy redirect */}
              <Route path="/workflow" element={<Navigate to="/instances" replace />} />
              <Route path="/workflow/*" element={<Navigate to="/instances" replace />} />
              <Route path="*" element={<Navigate to="/instances" replace />} />
            </Routes>
          </div>
        </SidebarInset>
      </div>
      </SidebarProvider>
      <Toaster position="top-right" expand={false} richColors theme="dark" />
    </DndProvider>
  );
}

export default App


