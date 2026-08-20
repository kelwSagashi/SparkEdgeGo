import { useEffect, useState } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useShallow } from "zustand/react/shallow";
import { AlertCircle, CheckCircle, Loader2, Save } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useAuthStore } from "@/stores/auth-store";
import { api } from "@/server/server.service";
import {
  InstanceFormSchema,
  type InstanceFormValues,
} from "./instance-form.schemas";
import { InstanceBasicForm } from "./instance-basic-form";
import { InstanceDestinationsForm } from "./instance-destinations-form";
import { InstanceFallbackForm } from "./instance-fallback-form";
import { InstanceMappingForm } from "./instance-mapping-form";
import { InstanceScriptForm } from "./instance-script-form";
import { InstanceTriggerForm } from "./instance-trigger-form";
import type {
  DeviceReturningValues,
  DownloadedScriptReturningValues,
  ProjectReturningValues,
  ResourceOperationReturningValues,
  ServerReturningValues,
} from "@/types/db";

type Props = {
  instanceId?: string;
  onClose?: () => void;
};

const defaultValues: InstanceFormValues = {
  name: "",
  description: "",
  project_id: "",
  device_id: null,
  tags: [],
  includeDeviceData: false,
  script_id: "",
  scriptParameters: [],
  scriptInputs: {},
  triggerType: "interval",
  triggerConfig: {
    interval_seconds: 300,
    webhook_path: undefined,
    webhook_secret: undefined,
    event_name: undefined,
    mqtt_topic: undefined,
    state_field: undefined,
    state_equals: undefined,
    save_execution_on_server: true,
  },
  dependsOn: [],
  executionMode: "sequential",
  orchestrationConfig: {
    workflow_enabled: false,
    allow_partial_success: true,
    debounce_seconds: undefined,
  },
  destinations: [],
  fallbackConfig: {
    enabled: true,
    strategy: "background_job",
    retry_interval_seconds: 300,
    max_retries: undefined,
  },
  errorConfig: {
    action: "log_only",
    notify_url: undefined,
    max_retries: undefined,
    retry_interval_seconds: 5,
  },
  active: true,
};

type ServerWithResources = ServerReturningValues & {
  resources?: Array<{
    resource: any;
    operations: ResourceOperationReturningValues[];
  }>;
};

function flattenOperations(resources: any[]): ResourceOperationReturningValues[] {
  const operations: ResourceOperationReturningValues[] = [];
  resources.forEach((item: any) => {
    (item.operations || []).forEach((operation: any) => {
      operations.push({ ...(operation as any), id: String(operation.id) });
    });
  });
  return operations;
}

export default function InstanceStepForm({ instanceId, onClose }: Props) {
  const [, project] = useAuthStore(useShallow((state) => [state.user, state.project]));
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState("basic");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [projects, setProjects] = useState<ProjectReturningValues[]>([]);
  const [devices, setDevices] = useState<DeviceReturningValues[]>([]);
  const [scripts, setScripts] = useState<DownloadedScriptReturningValues[]>([]);
  const [servers, setServers] = useState<ServerWithResources[]>([]);
  const [instances, setInstances] = useState<any[]>([]);
  const [allOperations, setAllOperations] = useState<ResourceOperationReturningValues[]>([]);
  const [operationsCache, setOperationsCache] = useState<
    Record<string, ResourceOperationReturningValues[]>
  >({});

  const form = useForm<InstanceFormValues>({
    resolver: zodResolver(InstanceFormSchema),
    defaultValues,
    mode: "onSubmit",
  });

  const updateOperationsCache = (
    serverId: string,
    operations: ResourceOperationReturningValues[],
  ) => {
    setOperationsCache((previous) => ({ ...previous, [serverId]: operations }));
    setAllOperations((previous) => {
      const merged = new Map(previous.map((item) => [String(item.id), item]));
      operations.forEach((item) => merged.set(String(item.id), item));
      return Array.from(merged.values());
    });
  };

  useEffect(() => {
    const loadData = async () => {
      try {
        setLoading(true);

        const [projectsResponse, devicesResponse, scriptsResponse, serversResponse, instancesResponse] =
          await Promise.all([
            api.listAllProjects(),
            api.listAllDevices(),
            api.listAllScripts(),
            api.listAllServers(),
            api.listAllInstances(),
          ]);

        const projectsData = projectsResponse.data?.data || [];
        const devicesData = devicesResponse.data?.data || [];
        const scriptsData = scriptsResponse.data?.data || [];
        const serversData = (serversResponse.data?.data || []) as ServerWithResources[];
        const instancesData = (instancesResponse.data?.data || []).filter(
          (item: any) => item.id !== instanceId,
        );

        setProjects(projectsData);
        setDevices(devicesData);
        setScripts(scriptsData);
        setServers(serversData);
        setInstances(instancesData);

        let initialOperationsCache: Record<string, ResourceOperationReturningValues[]> = {};
        let initialOperations: ResourceOperationReturningValues[] = [];

        if (instanceId && serversData.length > 0) {
          const resourceResponses = await Promise.all(
            serversData.map(async (server) => {
              try {
                const response = await api.listResources(String(server.id));
                const operations = flattenOperations(response.data?.data || []);
                return [String(server.id), operations] as const;
              } catch (loadError) {
                console.error("Failed to load operations for server", server.id, loadError);
                return [String(server.id), []] as const;
              }
            }),
          );

          initialOperationsCache = Object.fromEntries(resourceResponses);
          initialOperations = resourceResponses.flatMap(([, operations]) => operations);
          setOperationsCache(initialOperationsCache);
          setAllOperations(initialOperations);
        }

        if (!instanceId) {
          if (project?.id) {
            form.setValue("project_id", project.id);
          }
          return;
        }

        const instanceResponse = await api.getInstanceById(instanceId);
        const instanceData = instanceResponse.data?.data as any;
        if (!instanceData) {
          return;
        }

        const operationToServer = new Map<string, string>();
        Object.entries(initialOperationsCache).forEach(([serverId, operations]) => {
          operations.forEach((operation) => {
            operationToServer.set(String(operation.id), serverId);
          });
        });

        const { instance, destinations } = instanceData;
        const scriptInputs = instance.script_parameters || {};
        const scriptParameters = Object.entries(scriptInputs).map(([key, value]) => ({
          key,
          value,
          sourceType: "manual" as const,
        }));

        const formValues: InstanceFormValues = {
          name: instance.name || "",
          description: instance.description || "",
          project_id: instance.project_id || "",
          device_id: instance.device_id || null,
          tags: instance.tags || [],
          includeDeviceData: !!instance.include_device_data,
          script_id: instance.script_id || "",
          scriptParameters,
          scriptInputs,
          triggerType: instance.trigger_type || "interval",
          triggerConfig: instance.trigger_config || defaultValues.triggerConfig,
          dependsOn: instance.depends_on || [],
          executionMode: instance.execution_mode || "sequential",
          orchestrationConfig: instance.orchestration_config || defaultValues.orchestrationConfig,
          active: instance.active ?? true,
          destinations: (destinations || []).map((item: any) => {
            const mapping =
              item.mapping || item.data_mapping || item.dataMapping || {};
            const operationId = String(item.destination?.resource_operation_id || "");
            return {
              resourceOperationId: operationId,
              serverId: operationToServer.get(operationId) || "",
              enabled: item.destination?.enabled ?? true,
              priority: item.destination?.priority || 0,
              retryPolicy: {
                maxRetries: item.destination?.retry_policy?.max_retries,
                retryInterval: item.destination?.retry_policy?.retry_interval,
                timeoutSeconds: item.destination?.retry_policy?.timeout_seconds,
                continueOnError: item.destination?.retry_policy?.continue_on_error ?? false,
                isolationMode: item.destination?.retry_policy?.isolation_mode ?? "isolate",
                circuitBreakerThreshold:
                  item.destination?.retry_policy?.circuit_breaker_threshold,
                circuitBreakerCooldownSeconds:
                  item.destination?.retry_policy?.circuit_breaker_cooldown_seconds,
              },
              dataMapping: {
                instanceDestinationId: item.destination?.id || "",
                mapping: mapping.mapping || {},
                payloadTemplate: mapping.payload_template || {},
                customFields: mapping.custom_fields || [],
                transformScript: mapping.transform_script || "",
              },
            };
          }),
          fallbackConfig: {
            enabled: instance.fallback_enabled ?? true,
            strategy: instance.fallback_strategy || "background_job",
            retry_interval_seconds:
              instance.fallback_retry_interval_seconds || 300,
            max_retries: instance.fallback_config?.max_retries ?? null,
          },
          errorConfig: {
            action: instance.on_error_action || "log_only",
            notify_url: instance.on_error_config?.notify_url,
            max_retries: instance.on_error_config?.max_retries,
            retry_interval_seconds: instance.on_error_config?.retry_interval_seconds ?? 5,
          },
        };

        form.reset(formValues);
      } catch (loadError) {
        console.error("Failed to load data", loadError);
        setError("Nao foi possivel carregar os dados da instancia.");
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [form, instanceId, project?.id]);

  useEffect(() => {
    const handleSchemaUpdated = (event: Event) => {
      const customEvent = event as CustomEvent<{
        script?: DownloadedScriptReturningValues;
      }>;
      const updatedScript = customEvent.detail?.script;
      if (!updatedScript?.id) {
        return;
      }

      setScripts((current) => {
        const index = current.findIndex((item) => item.id === updatedScript.id);
        if (index === -1) {
          return current;
        }
        const next = [...current];
        next[index] = updatedScript;
        return next;
      });
    };

    window.addEventListener("sparkedge-script-schema-updated", handleSchemaUpdated as EventListener);
    return () => {
      window.removeEventListener("sparkedge-script-schema-updated", handleSchemaUpdated as EventListener);
    };
  }, []);

  const onSubmit = async (data: InstanceFormValues) => {
    try {
      setSubmitting(true);
      setError(null);
      setSuccess(null);

      const resolvedScriptInputs = data.scriptInputs || {};
      const payload = {
        name: data.name,
        description: data.description,
        project_id: data.project_id || project?.id,
        device_id: data.device_id,
        tags: data.tags,
        script_id: data.script_id,
        script_parameters: resolvedScriptInputs,
        script_inputs: resolvedScriptInputs,
        trigger_type: data.triggerType,
        trigger_config: data.triggerConfig,
        depends_on: data.dependsOn,
        execution_mode: data.executionMode,
        orchestration_config: data.orchestrationConfig,
        include_device_data: data.includeDeviceData,
        fallback_config: data.fallbackConfig,
        error_config: {
          action: data.errorConfig.action,
          notify_url: data.errorConfig.notify_url,
          max_retries: data.errorConfig.max_retries,
          retry_interval_seconds: data.errorConfig.retry_interval_seconds,
        },
        active: data.active,
        destinations: data.destinations.map((destination) => ({
          resource_operation_id: destination.resourceOperationId,
          enabled: destination.enabled,
          priority: destination.priority,
          retry_policy: destination.retryPolicy
            ? {
                max_retries: destination.retryPolicy.maxRetries,
                retry_interval: destination.retryPolicy.retryInterval,
                timeout_seconds: destination.retryPolicy.timeoutSeconds,
                continue_on_error: destination.retryPolicy.continueOnError,
                isolation_mode: destination.retryPolicy.isolationMode,
                circuit_breaker_threshold:
                  destination.retryPolicy.circuitBreakerThreshold,
                circuit_breaker_cooldown_seconds:
                  destination.retryPolicy.circuitBreakerCooldownSeconds,
              }
            : {},
          data_mapping: destination.dataMapping
            ? {
                instance_destination_id: destination.dataMapping.instanceDestinationId,
                mapping: destination.dataMapping.mapping || {},
                payload_template: destination.dataMapping.payloadTemplate || {},
                custom_fields: destination.dataMapping.customFields || [],
                transform_script: destination.dataMapping.transformScript || "",
              }
            : undefined,
        })),
      };

      const response = instanceId
        ? await api.updateInstance(instanceId, payload)
        : await api.createInstance(payload);

      if (response.data.error) {
        throw new Error(
          typeof response.data.error === "string"
            ? response.data.error
            : JSON.stringify(response.data.error),
        );
      }

      setSuccess(
        `Instancia ${instanceId ? "atualizada" : "criada"} com sucesso!`,
      );
      setTimeout(() => onClose?.(), 1500);
    } catch (submitError: any) {
      const message =
        submitError?.response?.data?.error ||
        submitError?.message ||
        "Erro ao salvar instancia";
      setError(message);
      console.error("Failed to save instance", submitError);
    } finally {
      setSubmitting(false);
    }
  };

  const selectedScript = scripts.find((item) => item.id === form.watch("script_id"));
  const selectedDevice = devices.find((item) => item.id === form.watch("device_id"));

  if (loading) {
    return (
      <Card className="p-8 flex items-center justify-center min-h-96">
        <Loader2 className="animate-spin mr-2" />
        <span>Carregando dados...</span>
      </Card>
    );
  }

  return (
    <FormProvider {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit, (formErrors) => console.log(formErrors))}
        className="w-full min-w-0 space-y-6"
      >
        {error && (
          <Card className="p-4 bg-destructive/10 border-destructive/20 text-destructive flex gap-3">
            <AlertCircle className="text-destructive shrink-0" size={20} />
            <div>
              <p className="font-medium">Erro ao salvar</p>
              <p className="text-sm opacity-90">{error}</p>
            </div>
          </Card>
        )}

        {success && (
          <Card className="p-4 bg-primary/10 border-primary/20 text-primary flex gap-3">
            <CheckCircle className="text-primary shrink-0" size={20} />
            <div>
              <p className="font-medium">Sucesso!</p>
              <p className="text-sm opacity-90">{success}</p>
            </div>
          </Card>
        )}

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full min-w-0">
          <TabsList className="grid w-full grid-cols-6 bg-transparent gap-1">
            <TabsTrigger
              value="basic"
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Basico
            </TabsTrigger>
            <TabsTrigger
              value="script"
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Script
            </TabsTrigger>
            <TabsTrigger
              value="trigger"
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Trigger
            </TabsTrigger>
            <TabsTrigger
              value="destinations"
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Destinos
            </TabsTrigger>
            <TabsTrigger
              value="mapping"
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Mapping
            </TabsTrigger>
            <TabsTrigger
              value="fallback"
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Fallback
            </TabsTrigger>
          </TabsList>

          <div className="mt-6 min-h-[500px] w-full min-w-0">
            <TabsContent value="basic" className="w-full min-w-0">
              <InstanceBasicForm projects={projects} devices={devices} />
            </TabsContent>

            <TabsContent value="script" className="w-full min-w-0">
              <InstanceScriptForm
                scripts={scripts}
                selectedDevice={selectedDevice}
                includeDeviceData={form.watch("includeDeviceData")}
                instanceId={instanceId}
              />
            </TabsContent>

            <TabsContent value="trigger" className="w-full min-w-0">
              <InstanceTriggerForm instances={instances} />
            </TabsContent>

            <TabsContent value="destinations" className="w-full min-w-0">
              <InstanceDestinationsForm
                servers={servers}
                allOperations={allOperations}
                operationsCache={operationsCache}
                onUpdateOperationsCache={updateOperationsCache}
              />
            </TabsContent>

            <TabsContent value="mapping" className="w-full min-w-0">
              <InstanceMappingForm
                allOperations={allOperations}
                selectedScript={selectedScript}
                selectedDevice={selectedDevice}
                includeDeviceData={form.watch("includeDeviceData")}
                instanceId={instanceId}
              />
            </TabsContent>

            <TabsContent value="fallback" className="w-full min-w-0">
              <InstanceFallbackForm />
            </TabsContent>
          </div>
        </Tabs>

        <div className="flex justify-between mt-8">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancelar
          </Button>

          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                const tabs = [
                  "basic",
                  "script",
                  "trigger",
                  "destinations",
                  "mapping",
                  "fallback",
                ];
                const currentIndex = tabs.indexOf(activeTab);
                if (currentIndex > 0) {
                  setActiveTab(tabs[currentIndex - 1]);
                }
              }}
              disabled={activeTab === "basic"}
              className="text-secondary"
            >
              Anterior
            </Button>

            <Button
              type="button"
              variant="outline"
              onClick={() => {
                const tabs = [
                  "basic",
                  "script",
                  "trigger",
                  "destinations",
                  "mapping",
                  "fallback",
                ];
                const currentIndex = tabs.indexOf(activeTab);
                if (currentIndex < tabs.length - 1) {
                  setActiveTab(tabs[currentIndex + 1]);
                }
              }}
              disabled={activeTab === "fallback"}
              className="text-secondary"
            >
              Proximo
            </Button>

            <Button type="submit" disabled={submitting} className="gap-2 text-secondary">
              {submitting && <Loader2 className="animate-spin" size={16} />}
              <Save size={16} />
              {instanceId ? "Atualizar" : "Criar"} Instancia
            </Button>
          </div>
        </div>
      </form>
    </FormProvider>
  );
}
