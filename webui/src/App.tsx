import React, { Suspense, lazy, useEffect } from 'react'
import './App.css'
import { SidebarInset, SidebarProvider } from './components/ui/sidebar';
import SideBar from './components/sidebar';
import { Routes, Route, Navigate, } from 'react-router-dom'
import Login from './pages/Login';
import Register from './pages/Register';
import ProtectedRoute from './components/ProtectedRoute';
import { useAuthStore } from './stores/auth-store';
import { Toaster } from 'sonner';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';

const InstancesPage = lazy(() => import('./pages/Instances'));
const InstanceCreatePage = lazy(() => import('./pages/InstanceCreatePage'));
const ExecutionHistoryPage = lazy(() => import('./pages/ExecutionHistory'));
const LocalDataPage = lazy(() => import('./pages/LocalData'));
const ScriptHubPage = lazy(() => import('./pages/ScriptHub'));
const ScriptDetailsPage = lazy(() => import('./pages/ScriptDetailsPage'));
const ScriptEditPage = lazy(() => import('./pages/ScriptEditPage'));
const ScriptComposerPage = lazy(() => import('./pages/ScriptComposerPage'));
const ScriptPlaygroundPage = lazy(() => import('./pages/ScriptPlaygroundPage'));
const ServersPage = lazy(() => import('./pages/Servers'));
const ServerCreatePage = lazy(() => import('./pages/ServerCreatePage'));
const ServerEditPage = lazy(() => import('./pages/ServerEditPage'));
const DevicesPage = lazy(() => import('./pages/Devices'));
const SettingsPage = lazy(() => import('./pages/Settings'));
const ProjectsPage = lazy(() => import('./pages/Projects'));
const CredentialsPage = lazy(() => import('./pages/Credentials'));
const DeviceCreationPage = lazy(() => import('./pages/DeviceCreatePage'));
const DeviceEditPage = lazy(() => import('./pages/DeviceEditPage'));
const InstanceEditPage = lazy(() => import('./pages/InstanceEditPage'));
const InstanceLogsPage = lazy(() => import('./pages/InstanceLogsPage'));
const CloudSettingsPage = lazy(() => import('./pages/CloudSettings'));
const AdvancedSettingsPage = lazy(() => import('./pages/AdvancedSettings'));
const UpdateSettingsPage = lazy(() => import('./pages/UpdateSettings'));

function RouteFallback() {
  return (
    <div className="flex min-h-[320px] items-center justify-center text-zinc-500">
      <div className="text-sm">Carregando pagina...</div>
    </div>
  );
}

function ProtectedPage({ children }: { children: React.ReactNode }) {
  return (
    <ProtectedRoute>
      <Suspense fallback={<RouteFallback />}>{children}</Suspense>
    </ProtectedRoute>
  );
}

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
          <div className="grow relative min-w-0 overflow-auto">
            <Routes>
              {/* Instances */}
              <Route path="/" element={<Navigate to="/instances" replace />} />
              <Route path="/instances" element={<ProtectedPage><InstancesPage /></ProtectedPage>} />
              <Route path="/instances/new" element={<ProtectedPage><InstanceCreatePage /></ProtectedPage>} />
              <Route path="/instances/:id" element={<ProtectedPage><InstanceLogsPage /></ProtectedPage>} />
              <Route path="/instances/:id/edit" element={<ProtectedPage><InstanceEditPage /></ProtectedPage>} />

              {/* Script Hub */}
              <Route path="/script-hub" element={<ProtectedPage><ScriptHubPage /></ProtectedPage>} />
              <Route path="/script-hub/new" element={<ProtectedPage><ScriptComposerPage /></ProtectedPage>} />
              <Route path="/script-hub/playground/sample/:sampleName" element={<ProtectedPage><ScriptPlaygroundPage /></ProtectedPage>} />
              <Route path="/script-hub/:id/playground" element={<ProtectedPage><ScriptPlaygroundPage /></ProtectedPage>} />
              <Route path="/script-hub/:id/files/edit" element={<ProtectedPage><ScriptComposerPage /></ProtectedPage>} />
              <Route path="/script-hub/:id" element={<ProtectedPage><ScriptDetailsPage /></ProtectedPage>} />
              <Route path="/script-hub/:id/edit" element={<ProtectedPage><ScriptEditPage /></ProtectedPage>} />

              {/* Local Data */}
              <Route path="/data" element={<ProtectedPage><LocalDataPage /></ProtectedPage>} />

              {/* Execution History */}
              <Route path="/history" element={<ProtectedPage><ExecutionHistoryPage /></ProtectedPage>} />

              {/* Management */}
              <Route path="/servers" element={<ProtectedPage><ServersPage /></ProtectedPage>} />
              <Route path="/servers/new" element={<ProtectedPage><ServerCreatePage /></ProtectedPage>} />
              <Route path="/servers/:id" element={<ProtectedPage><ServerEditPage /></ProtectedPage>} />
              <Route path="/credentials" element={<ProtectedPage><CredentialsPage /></ProtectedPage>} />
              <Route path="/devices" element={<ProtectedPage><DevicesPage /></ProtectedPage>} />
              <Route path="/devices/new" element={<ProtectedPage><DeviceCreationPage /></ProtectedPage>} />
              <Route path="/devices/:id" element={<ProtectedPage><DeviceEditPage /></ProtectedPage>} />
              <Route path="/projects" element={<ProtectedPage><ProjectsPage /></ProtectedPage>} />

              {/* Settings */}
              <Route path="/settings" element={<ProtectedPage><SettingsPage /></ProtectedPage>} />
              <Route path="/settings/cloud" element={<ProtectedPage><CloudSettingsPage /></ProtectedPage>} />
              <Route path="/settings/advanced" element={<ProtectedPage><AdvancedSettingsPage /></ProtectedPage>} />
              <Route path="/settings/update" element={<ProtectedPage><UpdateSettingsPage /></ProtectedPage>} />

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


