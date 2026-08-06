import { axios_api_instance } from "@/lib/api-client";
import type { TagReturningValues } from "@/types/db";

export type Tag = TagReturningValues;

export const tagsApi = {
  list: () => axios_api_instance.get('/tags').then((res) => res.data),

  search: (q: string, projectId?: string) =>
    axios_api_instance.get(`/tags/search?q=${encodeURIComponent(q)}${projectId ? `&project_id=${projectId}` : ''}`).then((res) => res.data),

  create: (data: { name: string; color?: string; project_id?: string }) =>
    axios_api_instance.post('/tags', data).then((res) => res.data),

  delete: (id: string) =>
    axios_api_instance.delete(`/tags/${id}`).then((res) => res.data),
};

