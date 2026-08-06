import { axios_api_instance } from "@/lib/api-client";
import type { ReturningQueries } from "@/types/db";

export const fallbackApi = {
  listAll: () =>
    axios_api_instance.get<ReturningQueries<any[]>>('/fallback').then((res) => res.data),

  getStats: () =>
    axios_api_instance.get<ReturningQueries<any>>('/fallback/stats').then((res) => res.data),

  flush: () =>
    axios_api_instance.post<ReturningQueries<{ sent: number }>>('/fallback/flush').then((res) => res.data),

  retry: (id: string) =>
    axios_api_instance.post<ReturningQueries<{ success: boolean }>>(`/fallback/${id}/retry`).then((res) => res.data),

  delete: (id: string) =>
    axios_api_instance.delete<ReturningQueries<any>>(`/fallback/${id}`).then((res) => res.data),
};

