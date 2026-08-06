import { axios_api_instance } from "@/lib/api-client";
import type { InstanceExecutionReturningValues, ReturningQueries } from "@/types/db";

export const executionsApi = {
  list: () =>
    axios_api_instance.get<ReturningQueries<InstanceExecutionReturningValues[]>>('/executions').then((res) => res.data),

  get: (id: string) =>
    axios_api_instance.get<ReturningQueries<InstanceExecutionReturningValues>>(`/executions/${id}`).then((res) => res.data),

  listByInstance: (instanceId: string) =>
    axios_api_instance.get<ReturningQueries<InstanceExecutionReturningValues[]>>(`/executions/instance/${instanceId}`).then((res) => res.data),
};

