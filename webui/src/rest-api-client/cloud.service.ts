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
  config_file?: string;
}

export interface EdgeConfigUpdate {
  cloud?: { url?: string; mqtt_url?: string };
  db?: { file?: string };
  auth?: { jwt_secret?: string };
  server?: { port?: string | number };
}

export const cloudService = {
  getStatus: () => axios_api_instance.get<ReturningQueries<CloudStatus>>(`/cli/status`),

  getOnboarding: () =>
    axios_api_instance.get<ReturningQueries<{ complete: boolean; data: any }>>(`/cli/onboarding`),

  saveOnboarding: (data: {
    name: string;
    description?: string;
    lat: string;
    lng: string;
    tags: string[];
  }) =>
    axios_api_instance.post<ReturningQueries<{ success: boolean }>>(`/cli/onboarding`, data),

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

  getConfig: () =>
    axios_api_instance.get<ReturningQueries<EdgeConfig>>(`/cli/config`),

  updateConfig: (updates: EdgeConfigUpdate) =>
    axios_api_instance.put<ReturningQueries<{ success: boolean; message: string; config?: EdgeConfig }>>(`/cli/config`, updates),
};
