import type {
  CredentialReturningValues,
  CredentialUpsertValues,
  DeviceReturningValues,
  DeviceUpsertValues,
  ProjectReturningValues,
  ReturningQueries,
  ResourceOperationReturningValues,
  ResourceOperationUpsertValues,
  ServerResourceReturningValues,
  ServerResourceUpsertValues,
  ServerReturningValues,
  ServerTypeReturningValues,
  ServerUpsertValues,
  UserReturningValues,
  AuthorizationsTypeUpsertValues,
  AuthorizationsTypeReturningValues,
  InstanceReturningValues,
  DownloadedScriptReturningValues,
  InstanceDestinationReturningValues,
} from "@/types/db";
import { axios_api_instance } from "@/lib/api-client";

export interface ConfigField {
  key: string;
  label: string;
  type: "text" | "password" | "textarea" | "number" | "select" | "boolean";
  placeholder?: string;
  options?: { label: string; value: any }[];
  grid?: string;
}

export interface AdapterMetadata extends AuthorizationsTypeReturningValues {
  resourceFields?: ConfigField[];
  operationFields?: ConfigField[];
}

interface AdapterMetadataPayload {
  server_types?: ServerTypeReturningValues[];
  auth_types?: AdapterMetadata[];
}

const adapterMetadataFallbacks: Record<
  string,
  Pick<AdapterMetadata, "resourceFields" | "operationFields">
> = {
  supabase: {
    resourceFields: [
      {
        key: "table",
        label: "Tabela",
        type: "text",
        placeholder: "vehicle_model_catalog",
      },
    ],
    operationFields: [
      {
        key: "method",
        label: "Ação",
        type: "select",
        options: [
          { label: "Select", value: "select" },
          { label: "Insert", value: "insert" },
          { label: "Update", value: "update" },
          { label: "Upsert", value: "upsert" },
        ],
      },
    ],
  },
  mongo: {
    resourceFields: [
      {
        key: "collection",
        label: "Coleção",
        type: "text",
        placeholder: "my_collection",
      },
    ],
    operationFields: [
      {
        key: "operation",
        label: "Ação",
        type: "select",
        options: [
          { label: "Find", value: "find" },
          { label: "Insert One", value: "insertOne" },
          { label: "Update One", value: "updateOne" },
          { label: "Delete One", value: "deleteOne" },
        ],
      },
    ],
  },
};

function normalizeAdapterMetadata(item: AdapterMetadata): AdapterMetadata {
  const fallback =
    adapterMetadataFallbacks[item.id] ??
    (item.server_type_id
      ? adapterMetadataFallbacks[item.server_type_id]
      : undefined);

  return {
    ...item,
    resourceFields:
      item.resourceFields && item.resourceFields.length > 0
        ? item.resourceFields
        : fallback?.resourceFields,
    operationFields:
      item.operationFields && item.operationFields.length > 0
        ? item.operationFields
        : fallback?.operationFields,
  };
}

export class API {
  async listServersTypes() {
    const response =
      await axios_api_instance.get<
        ReturningQueries<ServerTypeReturningValues[]>
      >("/server-types");
    return response;
  }

  async listAdaptersMetadata() {
    const response =
      await axios_api_instance.get<ReturningQueries<AdapterMetadataPayload>>(
        "/adapters/metadata",
      );
    const authTypes = response.data?.data?.auth_types;
    if (Array.isArray(authTypes)) {
      response.data.data.auth_types = authTypes.map((item) =>
        normalizeAdapterMetadata(item),
      );
    }
    return response;
  }

  async listServers() {
    return axios_api_instance.get<ReturningQueries<ServerReturningValues[]>>(
      "/servers",
    );
  }

  async registerServer(payload: {
    server: any;
    resources: {
      resource: any;
      operations: any[];
    }[];
  }) {
    const response = await axios_api_instance.post<
      ReturningQueries<{
        server: ServerReturningValues;
        resources: {
          resource: ServerResourceReturningValues;
          operations: ResourceOperationReturningValues[];
        }[];
      }>
    >("/servers/register", payload);

    return response;
  }

  async listResources(serverId: string) {
    return axios_api_instance.get<ReturningQueries<any[]>>(
      `/servers/${serverId}/resources`,
    );
  }

  async createDevice(device: DeviceUpsertValues) {
    const response = await axios_api_instance.post<
      ReturningQueries<DeviceReturningValues | null>
    >("/devices", device);
    return response;
  }

  async getProject(userId: string, projectName: string = "PERSONAL") {
    const response = await axios_api_instance.get<
      ReturningQueries<{
        user: UserReturningValues;
        project: ProjectReturningValues;
      } | null>
    >(`/users/project/${userId}/${projectName}`);
    return response;
  }

  async listAllServers() {
    return axios_api_instance.get<ReturningQueries<ServerReturningValues[]>>(
      "/servers",
    );
  }

  async getServerById(serverId: string) {
    return axios_api_instance.get<ReturningQueries<ServerReturningValues>>(
      `/servers/${serverId}`,
    );
  }

  async deleteServerById(serverId: string) {
    return axios_api_instance.delete<ReturningQueries<any>>(
      `/servers/${serverId}`,
    );
  }

  // deprecated
  async listEndpoints(serverId: string) {
    return axios_api_instance.get<ReturningQueries<any[]>>(
      `/servers/${serverId}/endpoints`,
    );
  }

  async listAllDevices() {
    return axios_api_instance.get<ReturningQueries<DeviceReturningValues[]>>(
      "/devices",
    );
  }

  async getDeviceById(id: string) {
    return axios_api_instance.get<ReturningQueries<DeviceReturningValues>>(
      `/devices/${id}`,
    );
  }

  async updateDevice(id: string, device: DeviceUpsertValues) {
    return axios_api_instance.put<ReturningQueries<DeviceReturningValues>>(
      `/devices/${id}`,
      device,
    );
  }

  async deleteDevice(id: string) {
    return axios_api_instance.delete<ReturningQueries<any>>(`/devices/${id}`);
  }

  async testCredential(payload: { auth_type_id: string; data: any }) {
    return axios_api_instance.post<{ success: boolean; error?: string }>(
      "/credentials/test",
      payload,
    );
  }

  async discoverResources(adapterId: string, credentials: any) {
    return axios_api_instance.post<{ success?: boolean; resources?: any[] }>(
      `/adapters/${adapterId}/discover`,
      { credentials },
    );
  }

  async testServer(resource_operation_id: string) {
    return axios_api_instance.post<{ success: boolean; error?: string }>(
      "/servers/execute",
      { resource_operation_id },
    );
  }

  async executeOperation(resource_operation_id: string, payload?: any) {
    return axios_api_instance.post<{
      success: boolean;
      data?: any;
      error?: string;
    }>("/servers/execute", { resource_operation_id, payload });
  }

  // ─── Instances ────────────────────────────────────────────────────────

  async listAllInstances() {
    return axios_api_instance.get<ReturningQueries<InstanceReturningValues[]>>(
      "/instances",
    );
  }

  async listActiveInstances() {
    return axios_api_instance.get<ReturningQueries<InstanceReturningValues[]>>(
      "/instances/active",
    );
  }

  async listInstancesByProject(projectId: string) {
    return axios_api_instance.get<ReturningQueries<InstanceReturningValues[]>>(
      `/instances/project/${projectId}`,
    );
  }

  async getInstanceById(id: string) {
    return axios_api_instance.get<ReturningQueries<{ destinations: InstanceDestinationReturningValues[]; instance: InstanceReturningValues}>>(
      `/instances/${id}`,
    );
  }

  async createInstance(payload: any) {
    return axios_api_instance.post<ReturningQueries<InstanceReturningValues>>(
      "/instances",
      payload,
    );
  }

  async updateInstance(id: string, payload: any) {
    return axios_api_instance.put<ReturningQueries<InstanceReturningValues>>(
      `/instances/${id}`,
      payload,
    );
  }

  async deleteInstance(id: string) {
    return axios_api_instance.delete<ReturningQueries<any>>(`/instances/${id}`);
  }

  async triggerInstance(id: string) {
    return axios_api_instance.post<ReturningQueries<any>>(
      `/instances/${id}/trigger`,
      {},
    );
  }

  // ─── Scripts ──────────────────────────────────────────────────────────

  async listAllScripts() {
    return axios_api_instance.get<
      ReturningQueries<DownloadedScriptReturningValues[]>
    >("/scripts");
  }

  async getScriptById(id: string) {
    return axios_api_instance.get<
      ReturningQueries<DownloadedScriptReturningValues>
    >(`/scripts/${id}`);
  }

  // ─── Projects ─────────────────────────────────────────────────────────

  async listAllProjects() {
    return axios_api_instance.get<ReturningQueries<ProjectReturningValues[]>>(
      "/projects",
    );
  }

  async getProjectById(id: string) {
    return axios_api_instance.get<ReturningQueries<ProjectReturningValues>>(
      `/projects/${id}`,
    );
  }

  async listExecutions() {
    return axios_api_instance.get<ReturningQueries<any[]>>("/executions");
  }

  async listInstanceExecutions(id: string) {
    return axios_api_instance.get<ReturningQueries<any[]>>(`/instances/${id}/executions`);
  }
}

export const api = new API();

