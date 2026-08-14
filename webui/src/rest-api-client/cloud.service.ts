import { axios_api_instance } from "@/lib/api-client";

export interface CloudStatus {
  connected: boolean;
  edge_id: string | null;
  edge_name: string | null;
  mqtt: {
    connected: boolean;
  };
  connectivity?: {
    mode?: string;
    status?: string;
    reasons?: string[];
    heartbeat_interval_seconds?: number;
    stats_interval_seconds?: number;
    cloud_sync_configured?: boolean;
    mqtt_connected?: boolean;
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
    sync_token?: string;
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
  update: {
    enabled: boolean;
    provider: string;
    repo: string;
    channel: string;
    allow_prerelease: boolean;
    service_name?: string;
    restart_command?: string;
  };
  connectivity: {
    intermittent_pending_age_seconds: number;
    degraded_pending_age_seconds: number;
    intermittent_cloud_sync_queue_depth: number;
    degraded_cloud_sync_queue_depth: number;
    degraded_mqtt_queue_depth: number;
    heartbeat_healthy_seconds: number;
    heartbeat_degraded_seconds: number;
    stats_healthy_seconds: number;
    stats_degraded_seconds: number;
  };
  retention: {
    mqtt_queue_max_items: number;
    mqtt_queue_max_age_hours: number;
    cloud_sync_sent_retention_hours: number;
    cloud_sync_failed_retention_hours: number;
    cloud_sync_keep_sent_items: number;
    cloud_sync_keep_failed_items: number;
    local_fallback_sent_retention_hours: number;
    local_fallback_failed_retention_hours: number;
    local_fallback_keep_sent_items: number;
    local_fallback_keep_failed_items: number;
  };
  config_file?: string;
}

export interface EdgeConfigUpdate {
  cloud?: { url?: string; mqtt_url?: string; sync_token?: string };
  db?: { file?: string };
  auth?: { jwt_secret?: string };
  server?: { port?: string | number };
  update?: {
    enabled?: boolean;
    provider?: string;
    repo?: string;
    channel?: string;
    allow_prerelease?: boolean;
    service_name?: string;
    restart_command?: string;
  };
  connectivity?: Partial<EdgeConfig['connectivity']>;
  retention?: Partial<EdgeConfig['retention']>;
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
    axios_api_instance.get<EdgeConfig>(`/cli/config`),

  updateConfig: (updates: EdgeConfigUpdate) =>
    axios_api_instance.put<{ success: boolean; message: string; config?: EdgeConfig }>(`/cli/config`, updates),
};
