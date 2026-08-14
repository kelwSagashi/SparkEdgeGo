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
