import { create } from 'zustand';
import { api } from '@/server/server.service';
import type { ServerReturningValues } from "@/types/db";

type ServersState = {
  servers: ServerReturningValues[];
  loading: boolean;
  fetchAll: () => Promise<void>;
  createServer: (data: Partial<ServerReturningValues>) => Promise<ServerReturningValues | null>;
  deleteServer: (id: string) => Promise<void>;
};

export const useServerssStore = create<ServersState>((set) => ({
  servers: [],
  loading: false,
  fetchAll: async () => {
    set({ loading: true });
    try {
      const res = await api.listAllServers();
      set({ servers: res?.data.data ?? [] });
    } finally { set({ loading: false }); }
  },
  createServer: async (data) => {
    // const res = await scriptsApi.create(data);
    // if (res?.data) {
    //   set(s => ({ servers: [...s.servers, res.data] }));
    //   return res.data;
    // }
    return null;
  },
  deleteServer: async (id) => {
    await api.deleteServerById(id);
    set(s => ({ servers: s.servers.filter(sc => sc.id !== id) }));
  }
}));

