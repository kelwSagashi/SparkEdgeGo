import { useState, useEffect, useCallback } from 'react';
import { api } from '@/server/server.service';
import type { ServerReturningValues, ServerTypeReturningValues } from "@/types/db";
import type { ServerWithRelations } from '@/interfaces/servers';

/**
 * Custom hook for managing servers data and operations
 */
export function useServers() {
  const [servers, setServers] = useState<ServerWithRelations[]>([]);
  const [serverTypes, setServerTypes] = useState<ServerTypeReturningValues[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Fetch all servers from the API
   */
  const fetchServers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.listServers();
      if (response.data.error) {
        throw new Error('Failed to fetch servers');
      }
      setServers(response.data.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
      console.error('Error fetching servers:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * Fetch all server types from the API
   */
  const fetchServerTypes = useCallback(async () => {
    try {
      const response = await api.listServersTypes();
      if (response.data.error) {
        throw new Error('Failed to fetch server types');
      }
      setServerTypes(response.data.data);
    } catch (err) {
      console.error('Error fetching server types:', err);
    }
  }, []);

  /**
   * Delete a server by ID
   */
  const deleteServer = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    try {
      // TODO: Implement delete endpoint in API
      // await api.deleteServer(id);
      setServers(prev => prev.filter(server => server.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete server');
      console.error('Error deleting server:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * Refresh servers list
   */
  const refresh = useCallback(() => {
    fetchServers();
  }, [fetchServers]);

  // Initial load
  useEffect(() => {
    fetchServers();
    fetchServerTypes();
  }, [fetchServers, fetchServerTypes]);

  return {
    servers,
    serverTypes,
    loading,
    error,
    deleteServer,
    refresh,
  };
}

/**
 * Mock hook for development/testing with sample data
 */
export function useServersMock() {
  const [servers, setServers] = useState<ServerWithRelations[]>([
    {
      id: '1',
      name: 'Production API',
      type: 'rest-api',
      server_type_id: 'api',
      driver_key: 'http',
      base_url: 'https://api.production.com',
      credential_id: 'cred-1',
      headers: { 'X-API-Version': 'v1' },
      project_id: 'proj-1',
      created_by: 'user-1',
      created_at: new Date('2024-01-15').toISOString(),
      updated_at: new Date('2024-01-15').toISOString(),
      serverType: {
        id: 'rest-api',
        key: 'rest_api',
        name: 'REST API',
        description: 'RESTful API Server',
      },
      project: {
        id: 'proj-1',
        name: 'Main Project',
        key: 'main',
        description: 'Main project workspace',
        visibility: 'private',
        owner_id: 'user-1',
        created_at: new Date('2024-01-01').toISOString(),
        updated_at: new Date('2024-01-01').toISOString(),
      },
    },
    {
      id: '2',
      name: 'Development Database',
      type: 'database',
      server_type_id: 'database',
      driver_key: 'postgres',
      base_url: 'postgresql://localhost:5432',
      credential_id: 'cred-2',
      headers: null,
      project_id: 'proj-1',
      created_by: 'user-1',
      created_at: new Date('2024-02-10').toISOString(),
      updated_at: new Date('2024-02-10').toISOString(),
      serverType: {
        id: 'database',
        key: 'database',
        name: 'Database',
        description: 'Database Server',
      },
      project: {
        id: 'proj-1',
        name: 'Main Project',
        key: 'main',
        description: 'Main project workspace',
        visibility: 'private',
        owner_id: 'user-1',
        created_at: new Date('2024-01-01').toISOString(),
        updated_at: new Date('2024-01-01').toISOString(),
      },
    },
    {
      id: '3',
      name: 'Google Drive Integration',
      type: 'google-drive',
      server_type_id: 'google-drive',
      driver_key: 'google-drive',
      base_url: 'https://www.googleapis.com/drive/v3',
      credential_id: 'cred-3',
      headers: { 'Content-Type': 'application/json' },
      project_id: 'proj-2',
      created_by: 'user-1',
      created_at: new Date('2024-03-05').toISOString(),
      updated_at: new Date('2024-03-05').toISOString(),
      serverType: {
        id: 'google-drive',
        key: 'google_drive',
        name: 'Google Drive',
        description: 'Google Drive API',
      },
      project: {
        id: 'proj-2',
        name: 'Integration Project',
        key: 'integration',
        description: 'External integrations',
        visibility: 'private',
        owner_id: 'user-1',
        created_at: new Date('2024-02-01').toISOString(),
        updated_at: new Date('2024-02-01').toISOString(),
      },
    },
  ]);

  const [serverTypes] = useState<ServerTypeReturningValues[]>([
    {
      id: 'rest-api',
      key: 'rest_api',
      name: 'REST API',
      description: 'RESTful API Server',
    },
    {
      id: 'database',
      key: 'database',
      name: 'Database',
      description: 'Database Server',
    },
    {
      id: 'google-drive',
      key: 'google_drive',
      name: 'Google Drive',
      description: 'Google Drive API',
    },
  ]);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const deleteServer = useCallback(async (id: string) => {
    setLoading(true);
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    setServers(prev => prev.filter(server => server.id !== id));
    setLoading(false);
  }, []);

  const refresh = useCallback(async () => {
    setLoading(true);
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    setLoading(false);
  }, []);

  return {
    servers,
    serverTypes,
    loading,
    error,
    deleteServer,
    refresh,
  };
}

