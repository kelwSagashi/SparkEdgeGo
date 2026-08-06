import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Loader2, Save, AlertCircle, CheckCircle } from "lucide-react";
import { useState, useEffect } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useShallow } from "zustand/react/shallow";
// instance-form-styles removed — use design tokens directly
import {
  InstanceFormSchema,
  type InstanceFormValues,
} from "./instance-form.schemas";
import { InstanceBasicForm } from "./instance-basic-form";
import { InstanceScriptForm } from "./instance-script-form";
import { InstanceTriggerForm } from "./instance-trigger-form";
import { InstanceDestinationsForm } from "./instance-destinations-form";
import { InstanceMappingForm } from "./instance-mapping-form";
import { InstanceFallbackForm } from "./instance-fallback-form";
import { api } from "@/server/server.service";
import type {
  DeviceReturningValues,
  ServerReturningValues,
  ProjectReturningValues,
  DownloadedScriptReturningValues,
  ResourceOperationReturningValues,
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
    save_execution_on_server: true,
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
  },
  active: true,
};

export default function InstanceStepForm({ instanceId, onClose }: Props) {
  const [user, project] = useAuthStore(useShallow((s) => [s.user, s.project]));
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState("basic");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Data
  const [projects, setProjects] = useState<ProjectReturningValues[]>([]);
  const [devices, setDevices] = useState<DeviceReturningValues[]>([]);
  const [scripts, setScripts] = useState<DownloadedScriptReturningValues[]>([]);
  const [servers, setServers] = useState<
    (ServerReturningValues & {
      resources?: Array<{
        resource: any;
        operations: ResourceOperationReturningValues[];
      }>;
    })[]
  >([]);
  const [allOperations, setAllOperations] = useState<
    ResourceOperationReturningValues[]
  >([]);
  const [operationsCache, setOperationsCache] = useState<
    Record<string, ResourceOperationReturningValues[]>
  >({});

  const updateOperationsCache = (
    serverId: string,
    ops: ResourceOperationReturningValues[],
  ) => {
    setOperationsCache((prev) => {
      const next = { ...prev, [serverId]: ops };
      // Also update allOperations to include these new ones (avoiding duplicates)
      setAllOperations((prevAll) => {
        const existingIds = new Set(prevAll.map((o) => String(o.id)));
        const newOps = ops.filter((o) => !existingIds.has(String(o.id)));
        return [...prevAll, ...newOps];
      });
      return next;
    });
  };

  const form = useForm<InstanceFormValues>({
    resolver: zodResolver(InstanceFormSchema),
    defaultValues,
    mode: "onSubmit",
  });

  // Load data on mount
  useEffect(() => {
    const loadData = async () => {
      try {
        setLoading(true);
        const [resProjects, resDevices, resScripts, resServers] =
          await Promise.all([
            api.listAllProjects(),
            api.listAllDevices(),
            api.listAllScripts(),
            api.listAllServers(),
          ]);

        const projectsData = resProjects.data?.data || [];
        const devicesData = resDevices.data?.data || [];
        const scriptsData = resScripts.data?.data || [];
        const serversData = resServers.data?.data || [];

        setProjects(projectsData);
        setDevices(devicesData);
        setScripts(scriptsData);
        setServers(serversData);

        // Flatten all operations and initialize cache
        const ops: ResourceOperationReturningValues[] = [];
        const initialCache: Record<string, ResourceOperationReturningValues[]> =
          {};
        
        // Load instance data if editing
        if (instanceId) {
          const resInstance = await api.getInstanceById(instanceId);
          const instanceData = resInstance.data?.data;
          
          if (instanceData) {
            const { instance, destinations } = instanceData as any;
            
            // Convert dictionary params to array for form
            const scriptParams = Object.entries(instance.script_parameters || {}).map(([key, value]) => ({
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
              includeDeviceData: instance.include_device_data || false,
              script_id: instance.script_id || "",
              scriptParameters: scriptParams,
              scriptInputs: instance.script_parameters || {},
              triggerType: instance.trigger_type || "interval",
              triggerConfig: instance.trigger_config || defaultValues.triggerConfig,
              active: instance.active ?? true,
              destinations: (destinations || []).map((d: any) => ({
                resourceOperationId: d.destination.resource_operation_id,
                serverId: "", // We'll find this later or just leave it, individual dest forms load it
                enabled: d.destination.enabled ?? true,
                priority: d.destination.priority || 0,
                retryPolicy: {
                  maxRetries: d.destination.retry_policy?.max_retries,
                  retryInterval: d.destination.retry_policy?.retry_interval,
                },
                dataMapping: {
                  instanceDestinationId: d.destination.id,
                  mapping: d.mapping?.mapping || {},
                  payloadTemplate: d.mapping?.payload_template || {},
                  customFields: d.mapping?.custom_fields || [],
                  transformScript: d.mapping?.transform_script,
                }
              })),
              fallbackConfig: {
                enabled: instance.fallback_enabled ?? true,
                strategy: instance.fallback_strategy || "background_job",
                retry_interval_seconds: instance.fallback_retry_interval_seconds || 300,
                max_retries: null,
              },
              errorConfig: {
                action: instance.on_error_action || "log_only",
                notify_url: instance.on_error_config?.notify_url,
                max_retries: instance.on_error_config?.max_retries,
              }
            };

            form.reset(formValues);
          }
        } else if (project?.id) {
          form.setValue("project_id", project.id);
        }
      } catch (e) {
        console.error("Failed to load data", e);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [project?.id, form]);

  const onSubmit = async (data: InstanceFormValues) => {
    try {
      setSubmitting(true);
      setError(null);
      setSuccess(null);

      // Tipagem correta do payload baseada no servidor
      const payload = {
        name: data.name,
        description: data.description,
        project_id: data.project_id || project?.id,
        device_id: data.device_id,
        tags: data.tags,
        script_id: data.script_id,
        script_parameters: data.scriptParameters || [],
        script_inputs: data.scriptInputs || {},
        trigger_type: data.triggerType,
        trigger_config: data.triggerConfig,
        include_device_data: data.includeDeviceData,
        fallback_config: data.fallbackConfig,
        error_config: {
          action: data.errorConfig.action,
          notify_url: data.errorConfig.notify_url,
          max_retries: data.errorConfig.max_retries,
        },
        active: data.active,
        destinations: data.destinations.map((dest) => ({
          resource_operation_id: dest.resourceOperationId,
          enabled: dest.enabled,
          priority: dest.priority,
          retry_policy: dest.retryPolicy
            ? {
                max_retries: dest.retryPolicy.maxRetries,
                retry_interval: dest.retryPolicy.retryInterval,
              }
            : {},
          data_mapping: dest.dataMapping
            ? {
                instance_destination_id: dest.dataMapping.instanceDestinationId,
                mapping: dest.dataMapping.mapping,
                payload_template: dest.dataMapping.payloadTemplate,
                custom_fields: dest.dataMapping.customFields,
                transform_script: dest.dataMapping.transformScript,
              }
            : undefined,
        })),
      };

      const action = instanceId ? "Atualizar" : "Criar";
      let res;
      if (instanceId) {
        res = await api.updateInstance(instanceId, payload);
      } else {
        res = await api.createInstance(payload);
      }

      if (res.data.error) {
        throw new Error(
          typeof res.data.error === "string"
            ? res.data.error
            : JSON.stringify(res.data.error),
        );
      }

      setSuccess(
        `Instância ${action === "Atualizar" ? "atualizada" : "criada"} com sucesso!`,
      );
      setTimeout(() => onClose?.(), 1500);
    } catch (e: any) {
      const errorMsg =
        e?.response?.data?.error || e?.message || "Erro ao salvar instância";
      setError(errorMsg);
      console.error("Failed to save instance", e);
    } finally {
      setSubmitting(false);
    }
  };
  const selectedScript = scripts.find((s) => s.id === form.watch("script_id"));
  const selectedDevice = devices.find((d) => d.id === form.watch("device_id"));

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
      <form onSubmit={form.handleSubmit(onSubmit, (e) => console.log(e))} className="space-y-6">
        {/* Mensagens de Erro e Sucesso */}
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

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-6 bg-transparent gap-1">
            <TabsTrigger 
              value="basic" 
              className="text-primary hover:bg-primary/10 data-[state=active]:bg-primary/10 data-[state=active]:text-primary transition-all"
            >
              Básico
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

          <div className="mt-6 min-h-[500px]">
            <TabsContent value="basic">
              <InstanceBasicForm projects={projects} devices={devices} />
            </TabsContent>

            <TabsContent value="script">
              <InstanceScriptForm
                scripts={scripts}
                selectedDevice={selectedDevice}
                includeDeviceData={form.watch("includeDeviceData")}
              />
            </TabsContent>

            <TabsContent value="trigger">
              <InstanceTriggerForm />
            </TabsContent>

            <TabsContent value="destinations">
              <InstanceDestinationsForm
                servers={servers}
                allOperations={allOperations}
                operationsCache={operationsCache}
                onUpdateOperationsCache={updateOperationsCache}
              />
            </TabsContent>

            <TabsContent value="mapping">
              <InstanceMappingForm
                allOperations={allOperations}
                selectedScript={selectedScript}
                selectedDevice={selectedDevice}
                includeDeviceData={form.watch("includeDeviceData")}
              />
            </TabsContent>

            <TabsContent value="fallback">
              <InstanceFallbackForm />
            </TabsContent>
          </div>
        </Tabs>

        {/* Navigation Buttons */}
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
              ← Anterior
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
              Próximo →
            </Button>

            <Button
              type="submit"
              disabled={submitting}
              className="gap-2 text-secondary"
            >
              {submitting && <Loader2 className="animate-spin" size={16} />}
              <Save size={16} />
              {instanceId ? "Atualizar" : "Criar"} Instância
            </Button>
          </div>
        </div>
      </form>
    </FormProvider>
  );
}

