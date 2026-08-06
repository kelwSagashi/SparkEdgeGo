import { create } from 'zustand';
import { instancesApi, type InstanceRequest } from '@/rest-api-client/instances.service';
import type { InstanceReturningValues } from "@/types/db";

interface InstancesState {
  instances: InstanceReturningValues[];
  loading: boolean;
  error: string | null;
  selectedId: string | null;

  fetchAll: () => Promise<void>;
  fetchActive: () => Promise<void>;
  createInstance: (data: InstanceRequest.InstancePayload) => Promise<any>;
  updateInstance: (id: string, data: Partial<InstanceRequest.InstancePayload>) => Promise<void>;
  deleteInstance: (id: string) => Promise<void>;
  triggerInstance: (id: string) => Promise<{ executionId: string; status: string } | null>;
  setSelectedId: (id: string | null) => void;
}

export const useInstancesStore = create<InstancesState>((set, get) => ({
  instances: [],
  loading: false,
  error: null,
  selectedId: null,

  fetchAll: async () => {
    set({ loading: true, error: null });
    try {
      const res = await instancesApi.list();
      console.log("instances", res.data.data)
      set({ instances: res.data.data ?? [], loading: false });
    } catch (err: any) {
      set({ error: err.message, loading: false });
    }
  },

  fetchActive: async () => {
    set({ loading: true, error: null });
    try {
      const res = await instancesApi.listActive();
      set({ instances: res.data.data ?? [], loading: false });
    } catch (err: any) {
      set({ error: err.message, loading: false });
    }
  },

  createInstance: async (data) => {
    try {
      const res = await instancesApi.create(data);
      if (res.data) {
        set({ instances: [...get().instances, res.data.data] });
        return res.data;
      }
      return null;
    } catch (err: any) {
      set({ error: err.message });
      return null;
    }
  },

  updateInstance: async (id, data) => {
    try {
      const res = await instancesApi.update(id, data);
      if (res.data.error) {
        set({ error: typeof res.data.error === 'string' ? res.data.error : JSON.stringify(res.data.error) });
        return;
      }
      set({
        instances: get().instances.map(i => i.id === id ? res.data.data : i),
      });
    } catch (err: any) {
      set({ error: err.message });
    }
  },

  deleteInstance: async (id) => {
    try {
      await instancesApi.delete(id);
      set({ instances: get().instances.filter(i => i.id !== id) });
    } catch (err: any) {
      set({ error: err.message });
    }
  },

  triggerInstance: async (id) => {
    try {
      const res = await instancesApi.trigger(id);
      return res.data ?? null;
    } catch (err: any) {
      set({ error: err.message });
      return null;
    }
  },

  setSelectedId: (id) => set({ selectedId: id }),
}));

