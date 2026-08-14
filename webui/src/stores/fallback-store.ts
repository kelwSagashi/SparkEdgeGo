import { create } from 'zustand';
import type { LocalFallbackItemReturningValues } from "@/types/db";
import { fallbackApi, type FallbackStats } from '@/rest-api-client/fallback.service';

interface FallbackState {
  items: LocalFallbackItemReturningValues[];
  stats: FallbackStats | null;
  loading: boolean;
  refreshing: boolean;
  error: string | null;
  
  fetchItems: () => Promise<void>;
  flushQueue: () => Promise<number>;
  retryItem: (id: string) => Promise<boolean>;
  deleteItem: (id: string) => Promise<void>;
  reset: () => void;
}

export const useFallbackStore = create<FallbackState>((set, get) => ({
  items: [],
  stats: null,
  loading: false,
  refreshing: false,
  error: null,

  fetchItems: async () => {
    set({ loading: true, error: null });
    try {
      const [res, statsRes] = await Promise.all([fallbackApi.listAll(), fallbackApi.getStats()]);
      set({ items: res.data || [], stats: statsRes.data || null });
    } catch (err: any) {
      set({ error: err?.message || 'Erro ao carregar dados locais' });
    } finally {
      set({ loading: false });
    }
  },

  flushQueue: async () => {
    set({ refreshing: true, error: null });
    try {
      const res = await fallbackApi.flush();
      await get().fetchItems();
      return res.data?.sent || 0;
    } catch (err: any) {
      set({ error: err?.message || 'Erro ao reenviar dados' });
      return 0;
    } finally {
      set({ refreshing: false });
    }
  },

  retryItem: async (id: string) => {
    try {
      const res = await fallbackApi.retry(id);
      await get().fetchItems();
      return !!res.data?.success;
    } catch (err: any) {
      set({ error: err?.message || 'Erro ao processar reenvio' });
      return false;
    }
  },

  deleteItem: async (id: string) => {
    try {
      await fallbackApi.delete(id);
      await get().fetchItems();
    } catch (err: any) {
      set({ error: err?.message || 'Erro ao excluir item' });
    }
  },

  reset: () => set({ items: [], stats: null, loading: false, refreshing: false, error: null }),
}));
