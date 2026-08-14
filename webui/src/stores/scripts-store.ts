import { create } from 'zustand';
import { scriptsApi } from '@/rest-api-client/scripts.service';
import type { DownloadedScriptReturningValues, DownloadedScriptUpsertValues } from "@/types/db";

type ScriptsState = {
  scripts: DownloadedScriptReturningValues[];
  samples: string[];
  loading: boolean;
  fetchAll: () => Promise<void>;
  fetchSamples: () => Promise<void>;
  createScript: (data: Partial<DownloadedScriptUpsertValues>) => Promise<any>;
  upsertScript: (script: DownloadedScriptReturningValues) => void;
  deleteScript: (id: string) => Promise<void>;
};

export const useScriptsStore = create<ScriptsState>((set) => ({
  scripts: [],
  samples: [],
  loading: false,
  fetchAll: async () => {
    set({ loading: true });
    try {
      const res = await scriptsApi.list();
      set({ scripts: res?.data ?? [] });
    } finally { set({ loading: false }); }
  },
  createScript: async (data) => {
    const res = await scriptsApi.create(data);
    if (res?.data) {
      set(s => ({ scripts: [...s.scripts, res.data] }));
      return res.data;
    }
    return null;
  },
  upsertScript: (script) => {
    set((state) => {
      const index = state.scripts.findIndex((item) => item.id === script.id);
      if (index === -1) {
        return { scripts: [...state.scripts, script] };
      }
      const next = [...state.scripts];
      next[index] = script;
      return { scripts: next };
    });
  },
  deleteScript: async (id) => {
    await scriptsApi.delete(id);
    set(s => ({ scripts: s.scripts.filter(sc => sc.id !== id) }));
  },
  fetchSamples: async () => {
    const res = await scriptsApi.listSamples();
    set({ samples: res?.data ?? [] });
  },
}));

