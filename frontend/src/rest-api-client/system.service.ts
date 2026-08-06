import { axios_api_instance } from "@/lib/api-client";

export interface SystemIdentity {
  edge_id: string | null;
  edge_name: string | null;
  provisioned: boolean;
}

export interface MqttConfig {
  broker_url: string;
  username: string;
  has_password: boolean;
  password?: string;
}

export const systemService = {
  getMqttConfig: async () => {
    const response = await axios_api_instance.get<MqttConfig>('/cli/mqtt-config');
    return response.data;
  }
};
