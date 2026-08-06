import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { ScrollArea } from "../ui/scroll-area";
import { api, type AdapterMetadata } from "@/server/server.service";
import { useCallback, useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useShallow } from "zustand/react/shallow";
import { ServerBuilderSchema, type ServerBuilderValues } from "./schemas";
import { CredentialSelecion } from "./credential-selection";
import { ServerConfigForm } from "./server-config-form";
import { ServerBuilderForm } from "./server-builder-form";
import { useCredentialsStore } from "@/stores/credentials-store";
import { Loader2, Database } from "lucide-react";
import { Label } from "../ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Button } from "../ui/button";

type Props = {
  serverId?: string;
  onClose?: () => void;
};

export default function ServerStepForm({ serverId, onClose }: Props) {
  const [user, project] = useAuthStore(useShallow((s) => [s.user, s.project]));
  const [metadata, setMetadata] = useState<AdapterMetadata[]>([]);
  const [loading, setLoading] = useState(true);

  const form = useForm<ServerBuilderValues>({
    resolver: zodResolver(ServerBuilderSchema),
    defaultValues: {
      server: {
        name: "",
        type: "",
        server_type_id: "",
        driver_key: "",
      },
      resources: [],
    },
    mode: "onSubmit",
  });

  useEffect(() => {
    const fetchMetadata = async () => {
      try {
        const [resMeta] = await Promise.all([
          api.listAdaptersMetadata(),
          useCredentialsStore.getState().fetchAll(),
        ]);
        const meta = (resMeta.data?.data?.auth_types ?? []).filter(
          (item): item is AdapterMetadata => !!item && typeof item.id === "string",
        );
        setMetadata(meta);
      } catch (e) {
        console.error("Failed to fetch adapter metadata", e);
      } finally {
        setLoading(false);
      }
    };
    fetchMetadata();
  }, []);

  useEffect(() => {
    if (serverId && metadata.length > 0) {
      const load = async () => {
        try {
          const [resServer, resResources] = await Promise.all([
            api.getServerById(serverId),
            api.listResources(serverId),
          ]);

          if (resServer.data?.data) {
            const svr = resServer.data.data;
            const resourcesData = resResources.data?.data || [];

            // Mapping resources and operations to form structure
            const mappedResources = resourcesData.map((item: any) => ({
              id: item.resource.id,
              name: item.resource.name,
              type: item.resource.type,
              config: item.resource.config || {},
              operations: (item.operations || []).map((op: any) => ({
                id: op.id,
                name: op.name,
                type: op.type,
                config: op.config || {},
                input_schema: op.input_schema,
                output_schema: op.output_schema,
              })),
            }));

            form.reset({
              server: {
                name: svr.name,
                type: svr.type,
                server_type_id: svr.server_type_id || "",
                driver_key: svr.driver_key || svr.type || "",
                credential_id: svr.credential_id || "",
              },
              resources: mappedResources,
            });
          }
        } catch (e) {
          console.error("Failed to load server", e);
        }
      };
      load();
    }
  }, [serverId, metadata, form]);

  const onSubmit = useCallback(
    async (data: ServerBuilderValues) => {
      try {
        const res = await api.registerServer({
          server: {
            id: serverId || undefined,
            name: data.server.name,
            type: data.server.type,
            server_type_id: data.server.server_type_id || undefined,
            credential_id: data.server.credential_id || undefined,
            headers: {},
            project_id: project?.id || undefined,
            created_by: user?.id,
            driver_key: data.server.driver_key || data.server.type,
          },
          resources: data.resources.map((res) => ({
            resource: {
              id: res.id,
              config: res.config,
              name: res.name,
              type: res.type,
              server_id: serverId || undefined,
            },
            operations: res.operations.map((op) => ({
              id: op.id,
              resource_id: res.id,
              name: op.name,
              type: op.type,
              config: op.config,
              input_schema: op.input_schema,
              output_schema: op.output_schema,
            })),
          })),
        });

        console.log("Server saved successfully", res);
        if (onClose) onClose();
      } catch (error) {
        console.error(error);
        alert("Erro ao salvar servidor.");
      }
    },
    [serverId, project?.id, user?.id, onClose],
  );

  const selectedType = form.watch("server.type"); 

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 w-full">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <>
      <div className="flex min-h-0 py-2">
        <FormProvider {...form}>
          <form
            id="server_form"
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col h-full w-full"
          >
            <div className="min-w-full">
              <div className="flex min-h-0 h-full">
                <Tabs
                  className="flex flex-col w-full h-full overflow-hidden"
                  defaultValue="config"
                >
                  <TabsList>
                    <TabsTrigger className="text-primary" value="config">Configurações</TabsTrigger>
                    <TabsTrigger className="text-primary" value="resources" disabled={!selectedType}>
                      Recursos e Operações
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent
                    value="config"
                    className="flex-1 overflow-hidden min-h-0"
                  >
                    <ScrollArea className="h-full w-full p-4">
                      <ServerConfigForm form={form} metadata={metadata} />
                    </ScrollArea>
                  </TabsContent>

                  <TabsContent
                    value="resources"
                    className="flex-1 overflow-hidden min-h-0"
                  >
                    <ScrollArea className="h-full w-full p-4">
                      <ServerBuilderForm metadata={metadata} />
                    </ScrollArea>
                  </TabsContent>
                </Tabs>
              </div>
            </div>
          </form>
        </FormProvider>
      </div>

      <div className="bottom-0 left-0 right-0 py-6 flex justify-center items-center z-50">
        <div className="flex justify-end w-full">
          <Button
            type="submit"
            form="server_form"
            className={`bg-muted text-primary-foreground font-medium`}
          >
            Salvar
          </Button>
        </div>
      </div>
    </>
  );
}

