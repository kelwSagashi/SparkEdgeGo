import type {
  DataMappingUpsertValues,
  InstanceDetailReturningValues,
  InstanceDestinationUpsertValues,
  InstanceReturningValues,
  InstanceUpsertValues,
  ReturningQueries,
} from "@/types/db";
export type Instance = InstanceReturningValues;
import { axios_api_instance } from "@/lib/api-client";

export namespace InstanceRequest {

  export interface InstancePayload {
    instance?: Partial<InstanceUpsertValues>;
    destinations?: any[];
    active?: boolean;
    name?: string;
    description?: string;
    status?: 'idle' | 'running' | 'paused' | 'error';
    [key: string]: any;
  }

  export type Create = InstancePayload;
  export type Update = Partial<InstancePayload>;
}

export const instancesApi = {
  list: () => axios_api_instance.get<ReturningQueries<InstanceReturningValues[]>>('/instances'),

  listActive: () => axios_api_instance.get<ReturningQueries<InstanceReturningValues[]>>('/instances/active'),

  listByProject: (projectId: string) => axios_api_instance.get<ReturningQueries<InstanceReturningValues[]>>(`/instances/project/${projectId}`),

  get: (id: string) => axios_api_instance.get<ReturningQueries<InstanceDetailReturningValues>>(`/instances/${id}`),

  create: (data: InstanceRequest.Create) => axios_api_instance.post<ReturningQueries<InstanceReturningValues>>('/instances', data),

  update: (id: string, data: InstanceRequest.Update) => axios_api_instance.put<ReturningQueries<InstanceReturningValues>>(`/instances/${id}`, data),

  delete: (id: string) => axios_api_instance.delete(`/instances/${id}`),

  trigger: (id: string) => axios_api_instance.post(`/instances/${id}/trigger`),
};

export const executionsApi = {
  list: () => axios_api_instance.get<ReturningQueries<InstanceReturningValues>>('/executions'),
};

