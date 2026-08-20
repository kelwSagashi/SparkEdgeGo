import type { DownloadedScriptUpsertValues } from "@/types/db";
import { axios_api_instance } from "@/lib/api-client";

export type ScriptUploadFinalizePayload = {
  tempFolder: string;
  mainFile: string;
  name?: string;
  description?: string;
  tags?: string[];
  author?: string;
  version?: string;
};

export type ScriptDraftFile = {
  path: string;
  content: string;
};

export type ScriptDraftPayload = {
  files: ScriptDraftFile[];
  mainFile: string;
  name?: string;
  description?: string;
  tags?: string[];
  author?: string;
  version?: string;
  inputs?: Record<string, unknown>;
};

export const scriptsApi = {
  list: () => axios_api_instance.get('/scripts').then((res) => res.data),

  get: (id: string) => axios_api_instance.get(`/scripts/${id}`).then((res) => res.data),
  getById: (id: string) => axios_api_instance.get(`/scripts/${id}`).then((res) => res.data),
  getHistory: (id: string) => axios_api_instance.get(`/scripts/${id}/history`).then((res) => res.data),
  restoreHistory: (id: string, historyId: string) =>
    axios_api_instance.post(`/scripts/${id}/history/${historyId}/restore`).then((res) => res.data),
  getFileContent: (id: string, filename: string) => axios_api_instance.get(`/scripts/${id}/contents/${filename}`).then((res) => res.data),
  listFiles: (id: string) => axios_api_instance.get(`/scripts/${id}/files`).then((res) => res.data),

  create: (data: Partial<DownloadedScriptUpsertValues>) => axios_api_instance.post('/scripts', data).then((res) => res.data),

  update: (id: string, data: Partial<DownloadedScriptUpsertValues>) => axios_api_instance.put(`/scripts/${id}`, data).then((res) => res.data),

  delete: (id: string) => axios_api_instance.delete(`/scripts/${id}`).then((res) => res.data),

  uploadInspect: (formData: FormData) =>
    axios_api_instance.post('/scripts/upload/inspect', formData, { headers: { 'Content-Type': 'multipart/form-data' } }).then((res) => res.data),

  uploadFinalize: (data: ScriptUploadFinalizePayload) =>
    axios_api_instance.post('/scripts/upload/finalize', data).then((res) => res.data),

  replaceUploadFinalize: (id: string, data: ScriptUploadFinalizePayload) =>
    axios_api_instance.post(`/scripts/${id}/upload/finalize`, data).then((res) => res.data),

  draftFinalize: (data: ScriptDraftPayload) =>
    axios_api_instance.post('/scripts/draft/finalize', data).then((res) => res.data),

  replaceDraftFinalize: (id: string, data: ScriptDraftPayload) =>
    axios_api_instance.post(`/scripts/${id}/draft/finalize`, data).then((res) => res.data),

  runDraftPlayground: (data: ScriptDraftPayload) =>
    axios_api_instance.post('/scripts/draft/playground/run', data).then((res) => res.data),

  generateDraftReadme: (data: ScriptDraftPayload) =>
    axios_api_instance.post('/scripts/draft/readme', data).then((res) => res.data),

  listSamples: () => axios_api_instance.get('/scripts/samples/list').then((res) => res.data),

  getSampleSchema: (name: string) => axios_api_instance.get(`/scripts/samples/${name}/schema`).then((res) => res.data),

  runPlayground: (data: { script_id?: string; sample_name?: string; inputs: any }) =>
    axios_api_instance.post('/scripts/playground/run', data).then((res) => res.data),
};

