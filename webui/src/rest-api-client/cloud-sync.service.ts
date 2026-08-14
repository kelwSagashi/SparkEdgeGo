import { axios_api_instance } from "@/lib/api-client";
import type { ReturningQueries } from "@/types/db";

export interface CloudSyncItem {
  id: string;
  event_type: string;
  status: string;
  attempts: number;
  last_error?: string | null;
  payload?: Record<string, unknown>;
  next_retry_at?: string | null;
  sent_at?: string | null;
  created_at?: string;
  updated_at?: string;
}

export interface CloudSyncStats {
  configured: boolean;
  base_url?: string;
  edge_id?: string;
  pending?: number;
  failed?: number;
  sent?: number;
  total?: number;
  ready?: number;
  oldest_pending_age_seconds?: number;
  retention?: {
    sent_retention_hours?: number;
    failed_retention_hours?: number;
    keep_sent_items?: number;
    keep_failed_items?: number;
  };
  usage?: {
    pending_total_pct_of_failed_window?: number;
    sent_pct_of_sent_window?: number;
  };
  mqtt_queue?: {
    total?: number;
    oldest_pending_age_seconds?: number;
    usage_pct?: number;
    retention?: {
      max_items?: number;
      max_age_hours?: number;
      max_age_seconds?: number;
    };
  };
  connectivity?: {
    mode?: string;
    status?: string;
    reasons?: string[];
    heartbeat_interval_seconds?: number;
    stats_interval_seconds?: number;
    cloud_sync_configured?: boolean;
    mqtt_connected?: boolean;
    policy?: Record<string, number>;
  };
}

export const cloudSyncService = {
  list: () =>
    axios_api_instance.get<ReturningQueries<CloudSyncItem[]>>(`/cloud-sync`).then((res) => res.data),

  stats: () =>
    axios_api_instance.get<ReturningQueries<CloudSyncStats>>(`/cloud-sync/stats`).then((res) => res.data),

  flush: () =>
    axios_api_instance
      .post<ReturningQueries<{ sent: number; failed: number; items: number; skipped?: boolean }>>(`/cloud-sync/flush`)
      .then((res) => res.data),

  retry: (id: string) =>
    axios_api_instance
      .post<ReturningQueries<{ id: string; status: string; sent: boolean; skipped?: boolean; message?: string; last_error?: string }>>(
        `/cloud-sync/${id}/retry`,
      )
      .then((res) => res.data),

  remove: (id: string) =>
    axios_api_instance
      .delete<ReturningQueries<{ success: boolean; id: string }>>(`/cloud-sync/${id}`)
      .then((res) => res.data),
};
