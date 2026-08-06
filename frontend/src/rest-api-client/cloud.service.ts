import { axios_api_instance } from "@/lib/api-client";
import type { ReturningQueries } from "@/types/db";

export interface CloudStatus {
  connected: boolean;
  edge_id: string | null;
  edge_name: string | null;
  mqtt: {
    connected: boolean;
  };
}

export interface ConnectPayload {
  email: string;
  password: string;
  edge_name?: string;
}

export interface ConnectResult {
  success: boolean;
  edge_id: string;
  edge_name: string;
  mqtt: { connected: boolean };
}

// ─── CloudIntegration Config types ─────────────────────────────────────────

export interface EdgeConfig {
  cloud: {
    url: string;
    mqtt_url: string;
  };
  db: {
    file: string;
  };
  auth: {
    jwt_secret: string;
    is_default: boolean;
  };
  server: {
    port: string;
  };
}

export interface EdgeConfigUpdate {
  cloud?: { url?: string; mqtt_url?: string };
  db?: { file?: string };
  auth?: { jwt_secret?: string };
  server?: { port?: string | number };
}

export const cloudService = {
  getStatus: (): Promise<ReturningQueries<CloudStatus>> => axios_api_instance.get(`/cli/status`),

  getOnboarding: (): Promise<ReturningQueries<{ complete: boolean; data: any }>> => axios_api_instance.get(`/cli/onboarding`),

  saveOnboarding: (data: { name: string; description?: string; lat: string; lng: string; tags: string[] }): Promise<ReturningQueries<{ success: boolean }>> => axios_api_instance.post(`/cli/onboarding`, data),

  connect: (payload: ConnectPayload): Promise<ConnectResult> =>
    axios_api_instance.post(`/cli/connect`, payload),

  disconnect: (): Promise<{ success: boolean }> =>
    axios_api_instance.post(`/cli/disconnect`),

  reconnect: (): Promise<{ success: boolean; mqtt: { connected: boolean } }> =>
    axios_api_instance.post(`/cli/reconnect`),

  remove: (): Promise<{ success: boolean }> =>
    axios_api_instance.post(`/cli/remove`),

  pair: (token: string, name?: string): Promise<ConnectResult> =>
    axios_api_instance.post(`/cli/pair`, { token, name }),

  // ─── CloudIntegration Configuration ────────────────────────────────────────

  /** Get current configuration (jwt_secret is partially masked) */
  getConfig: (): Promise<ReturningQueries<EdgeConfig>> =>
    axios_api_instance.get(`/cli/config`),

  /** Persist config updates to spark-edge.config.yml */
  updateConfig: (updates: EdgeConfigUpdate): Promise<ReturningQueries<{ success: boolean; message: string }>> =>
    axios_api_instance.put(`/cli/config`, updates),
};
