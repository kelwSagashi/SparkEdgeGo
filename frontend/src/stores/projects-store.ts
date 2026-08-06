import { create } from 'zustand';
import { projectsApi } from '@/rest-api-client/projects.service';
import type { ProjectReturningValues, ProjectUpsertValues } from "@/types/db";

type ProjectsState = {
  projects: ProjectReturningValues[];
  loading: boolean;
  fetchAll: () => Promise<void>;
  createProject: (data: ProjectUpsertValues) => Promise<any>;
  updateProject: (id: string, data: Partial<ProjectUpsertValues>) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
};

export const useProjectsStore = create<ProjectsState>((set) => ({
  projects: [],
  loading: false,
  fetchAll: async () => {
    set({ loading: true });
    try {
      const res = await projectsApi.list();
      set({ projects: res?.data ?? [] });
    } finally { set({ loading: false }); }
  },
  createProject: async (data) => {
    const res = await projectsApi.create(data);
    if (res?.data) {
      set(s => ({ projects: [...s.projects, res.data] }));
      return res.data;
    }
    return null;
  },
  updateProject: async (id, data) => {
    await projectsApi.update(id, data);
    set(s => ({ projects: s.projects.map(p => p.id === id ? { ...p, ...data } : p) }));
  },
  deleteProject: async (id) => {
    await projectsApi.delete(id);
    set(s => ({ projects: s.projects.filter(p => p.id !== id) }));
  },
}));

