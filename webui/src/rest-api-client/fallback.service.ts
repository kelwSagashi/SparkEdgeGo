import { axios_api_instance } from "@/lib/api-client";
import type { ReturningQueries } from "@/types/db";

export interface FallbackItem {
  id: string;
  instance_id: string;
  destination_id: string | null;
  execution_id: string | null;
  status: 'pending' | 'sending' | 'sent' | 'failed';
  payload: string;
  filepath: string | null;
  retry_count: number;
  last_retry_at: string | null;
  last_error: string | null;
  created_at: string;
  updated_at: string;
}

export interface FallbackStats {
  total?: number;
  pending?: number;
  sending?: number;
  sent?: number;
  failed?: number;
  oldest_pending_age_seconds?: number;
  retention?: {
    sent_retention_hours?: number;
    failed_retention_hours?: number;
    keep_sent_items?: number;
    keep_failed_items?: number;
  };
  usage?: {
    sent_pct_of_sent_window?: number;
    failed_pct_of_failed_window?: number;
  };
}

export const fallbackApi = {
  listAll: () =>
    axios_api_instance.get<ReturningQueries<FallbackItem[]>>('/fallback').then((res) => res.data),

  getStats: () =>
    axios_api_instance.get<ReturningQueries<FallbackStats>>('/fallback/stats').then((res) => res.data),

  flush: () =>
    axios_api_instance.post<ReturningQueries<{ sent: number }>>('/fallback/flush').then((res) => res.data),

  retry: (id: string) =>
    axios_api_instance.post<ReturningQueries<{ success: boolean }>>(`/fallback/${id}/retry`).then((res) => res.data),

  delete: (id: string) =>
    axios_api_instance.delete<ReturningQueries<any>>(`/fallback/${id}`).then((res) => res.data),
};

