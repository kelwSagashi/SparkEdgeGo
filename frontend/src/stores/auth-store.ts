import { create } from 'zustand';
import type { ProjectReturningValues, UserReturningValues } from "@/types/db";
import { authApi } from '@/rest-api-client/auth.service';
import { api } from '@/server/server.service';

type AuthState = {
  user: UserReturningValues | null;
  project: ProjectReturningValues | null;
  loading: boolean;
  error?: string | null;
  setUser: (u: UserReturningValues | null) => void;
  setProject: (p: ProjectReturningValues | null) => void;
  loadMe: () => Promise<void>;
  loadProject: (projectName: string) => Promise<void>;
  login: (email: string, password: string) => Promise<boolean>;
  register: (email: string, password: string, name?: string) => Promise<boolean>;
  logout: () => Promise<void>;
  generateNewApiKey: (userId: string) => Promise<void>;
};

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  project: null,
  loading: true,
  error: null,
  setUser: (u) => set({ user: u, error: null }),
  setProject: (p) => set({ project: p }),
  loadMe: async () => {
    set({ loading: true, error: null });
    try {
      const res = await authApi.me();
      set({ user: res.data?.data ?? null });

      get().loadProject('PERSONAL');
      
    } catch (err: any) {
      set({ user: null });
    } finally {
      set({ loading: false });
    }
  },
  loadProject: async (projectName: string = 'PERSONAL') => {
    try {
      const user = get().user;
      if (!user) {
        set({ project: null });
        return;
      }
      const project = await api.getProject(user.id, projectName);
      // console.log('Loaded project:', project.data?.data);
      set({ project: project.data?.data?.project ?? null });
      
    } catch (err: any) {
      set({ project: null });
    }
  },
  login: async (email, password) => {
    set({ loading: true, error: null });
    try {
      await authApi.login({ email, password });
      await get().loadMe();
      await get().loadProject('PERSONAL');
      return true;
    } catch (err: any) {
      set({ error: err?.response?.data?.message ?? 'Login failed' });
      return false;
    } finally {
      set({ loading: false });
    }
  },
  register: async (email, password, name) => {
    set({ loading: true, error: null });
    try {
      await authApi.register({ email, password, first_name: name });
      await get().loadMe();
      return true;
    } catch (err: any) {
      set({ error: err?.response?.data?.message ?? 'Registration failed' });
      return false;
    } finally {
      set({ loading: false });
    }
  },
  logout: async () => {
    try {
      await authApi.logout();
    } catch (err) {
      // ignore
    }
    set({ user: null });
  },
  generateNewApiKey: async (userId: string) => {
    try {
      const res = await authApi.generateNewApiKey(userId);
      set({ user: { ...get().user, api_key: res.data?.data?.apiKey } as UserReturningValues });
    } catch (err: any) {
      set({ error: err?.response?.data?.message ?? 'Failed to generate new API key' });
    }
  }
}));

