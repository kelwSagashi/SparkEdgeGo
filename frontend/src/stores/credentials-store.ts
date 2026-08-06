import { create } from 'zustand';
import { credentialsApi } from '../rest-api-client/credentials.service';
import type { AuthorizationsTypeReturningValues, CredentialReturningValues, CredentialUpsertValues } from "@/types/db";

interface CredentialsState {
  credentials: CredentialReturningValues[];
  metadata: Record<string, any> | null;
  authTypes: AuthorizationsTypeReturningValues[];
  loading: boolean;
  error: string | null;
  fetchAll: () => Promise<void>;
  fetchMetadata: () => Promise<void>;
  createCredential: (data: Partial<CredentialUpsertValues>) => Promise<any>;
  updateCredential: (id: string, data: Partial<CredentialUpsertValues>) => Promise<any>;
  deleteCredential: (id: string) => Promise<void>;
}

export const useCredentialsStore = create<CredentialsState>((set, get) => ({
  credentials: [],
  metadata: null,
  authTypes: [],
  loading: false,
  error: null,

  fetchMetadata: async () => {
    try {
      const res = (await credentialsApi.getMeta()).data;

      const metadata = res.reduce((acc: Record<string, any>, item: any) => {
        acc[item.id] = item.fields;
        return acc;
      }, {});
      set({ metadata: metadata, authTypes: res });
    } catch (err: any) {
      console.error('Failed to fetch credentials metadata', err);
    }
  },

  fetchAll: async () => {
    set({ loading: true, error: null });
    try {
      const response: any = await credentialsApi.list();
      console.log('cred', response.data);
      set({ credentials: response.data || [] });
    } catch (err: any) {
      set({ error: err.message });
    } finally {
      set({ loading: false });
    }
  },

  createCredential: async (data) => {
    const res: any = await credentialsApi.create(data);
    await get().fetchAll();
    return res.data;
  },

  updateCredential: async (id, data) => {
    const res: any = await credentialsApi.update(id, data);
    await get().fetchAll();
    return res.data;
  },

  deleteCredential: async (id) => {
    await credentialsApi.delete(id);
    await get().fetchAll();
  },
}));

