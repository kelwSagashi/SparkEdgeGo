import { create } from 'zustand';
import { tagsApi, type Tag } from '@/rest-api-client/tags.service';

type TagsState = {
  tags: Tag[];
  loading: boolean;
  fetchAll: () => Promise<void>;
  search: (q: string, projectId?: string) => Promise<Tag[]>;
  createTag: (name: string, projectId?: string, color?: string) => Promise<Tag | null>;
  deleteTag: (id: string) => Promise<void>;
};

export const useTagsStore = create<TagsState>((set, get) => ({
  tags: [],
  loading: false,
  fetchAll: async () => {
    set({ loading: true });
    try {
      const res = await tagsApi.list();
      set({ tags: res?.data ?? [] });
    } finally { set({ loading: false }); }
  },
  search: async (q: string, projectId?: string) => {
    const res = await tagsApi.search(q, projectId);
    return res?.data ?? [];
  },
  createTag: async (name: string, projectId?: string, color?: string) => {
    const res = await tagsApi.create({ name, color, project_id: projectId });
    if (res?.data) {
      set(s => ({ tags: [...s.tags, res.data] }));
      return res.data;
    }
    return null;
  },
  deleteTag: async (id: string) => {
    await tagsApi.delete(id);
    set(s => ({ tags: s.tags.filter(t => t.id !== id) }));
  },
}));

