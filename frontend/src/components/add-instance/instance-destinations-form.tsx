import { useFormContext, Controller, useFieldArray } from "react-hook-form";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Plus, Trash2, Server } from "lucide-react";
import { useState, useEffect } from "react";
import { api } from "@/server/server.service";
// instance-form-styles.ts removed — use design system tokens directly
import type {
  ServerReturningValues,
  ResourceOperationReturningValues,
} from "@/types/db";
import type { InstanceFormValues } from "./instance-form.schemas";

interface InstanceDestinationsFormProps {
  servers: (ServerReturningValues & {
    resources?: Array<{
      resource: any;
      operations: ResourceOperationReturningValues[];
    }>;
  })[];
  allOperations: ResourceOperationReturningValues[];
  operationsCache: Record<string, ResourceOperationReturningValues[]>;
  onUpdateOperationsCache: (
    serverId: string,
    ops: ResourceOperationReturningValues[],
  ) => void;
  onServerChange?: (serverId: string) => void;
}

export function InstanceDestinationsForm({
  servers,
  allOperations,
  operationsCache,
  onUpdateOperationsCache,
  onServerChange,
}: InstanceDestinationsFormProps) {
  const {
    control,
    register,
    watch,
    setValue,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();
  const {
    fields: destFields,
    append: appendDest,
    remove: removeDest,
  } = useFieldArray({
    control,
    name: "destinations",
  });

  const [selectedOpToAdd, setSelectedOpToAdd] = useState<string>("");
  const [selectedServerToAdd, setSelectedServerToAdd] = useState<string>("");

  const handleAddDestination = () => {
    if (!selectedServerToAdd) return;
    // Prefer explicit selected operation, otherwise pick first op from server
    const ops = operationsCache[selectedServerToAdd] || [];

    const opId = selectedOpToAdd || (ops.length > 0 ? String(ops[0].id) : "");
    if (!opId) {
      alert("O servidor selecionado não possui operações configuradas.");
      return;
    }

    appendDest({
      resourceOperationId: String(opId),
      serverId: String(selectedServerToAdd),
      enabled: true,
      priority: destFields.length,
      retryPolicy: { maxRetries: 3, retryInterval: 60 },
      dataMapping: {
        instanceDestinationId: "",
        mapping: {},
        customFields: [],
      },
    });

    // reset selection after adding
    setSelectedOpToAdd("");
    setSelectedServerToAdd("");
  };

  // Group operations by server (for initial load fallback)
  const operationsByServer = servers.reduce(
    (acc, server) => {
      server.resources?.forEach((res) => {
        res.operations.forEach((op) => {
          if (!acc[server.id]) {
            acc[server.id] = { serverName: server.name, operations: [] };
          }
          // normalize operation id to string for Select value matching
          acc[server.id].operations.push({ ...(op as any), id: String(op.id) });
        });
      });
      return acc;
    },
    {} as Record<
      string,
      { serverName: string; operations: ResourceOperationReturningValues[] }
    >,
  );

  // When the user selects a server to add, fetch its resources/operations if not cached
  useEffect(() => {
    const fetchOps = async (serverId: string) => {
      if (!serverId) return;
      if (operationsCache[serverId]) return; // already cached
      try {
        const res = await api.listResources(serverId);
        const list = res.data?.data || [];
        // flatten operations from resources
        const ops: ResourceOperationReturningValues[] = [];
        list.forEach((item: any) => {
          const operations = item.operations || [];
          operations.forEach((op: any) =>
            ops.push({ ...(op as any), id: String(op.id) }),
          );
        });
        onUpdateOperationsCache(serverId, ops);
      } catch (e) {
        console.error("Failed to fetch resources for server", serverId, e);
        onUpdateOperationsCache(serverId, []);
      }
    };

    if (selectedServerToAdd) fetchOps(selectedServerToAdd);
  }, [selectedServerToAdd, operationsCache, onUpdateOperationsCache]);

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <h3 className="font-medium text-primary uppercase text-xs tracking-wider">Destinos de Envio *</h3>
          <p className="text-sm text-primary/80">
            Adicione um ou mais servidores/endpoints onde os dados serão
            enviados.
          </p>
        </div>

        {/* Lista de Destinos */}
        {destFields.length > 0 ? (
          <div className="space-y-3">
            {destFields.map((field, index) => {
              const resourceOpId = watch(
                `destinations.${index}.resourceOperationId`,
              );
              const selectedOp = allOperations.find(
                (op) => String(op.id) === String(resourceOpId),
              );

              return (
                <Card
                  key={field.id}
                  className="p-4 bg-foreground border-none"
                >
                  <div className="flex justify-between items-start gap-2">
                    <div className="flex-1">
                      <p className="font-medium text-primary flex items-center">
                        <Server size={16}/> Destino {index + 1}
                      </p>
                    </div>
                    {destFields.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removeDest(index)}
                      >
                        <Trash2 size={16} className="text-destructive" />
                      </Button>
                    )}
                  </div>

                  {/* Seleção de Operação */}
                  <div className="space-y-2">
                    <Label className="text-sm font-medium text-primary">Operação do Servidor *</Label>
                    <Controller
                      name={`destinations.${index}.resourceOperationId`}
                      control={control}
                      render={({ field }) => {
                        // prefer cached API operations for the specific server if available
                        const serverIdValue = String(
                          watch(`destinations.${index}.serverId`) || "",
                        );
                        const opsList =
                          operationsCache[serverIdValue] ||
                          operationsByServer[serverIdValue]?.operations ||
                          [];
                        const serverName =
                          operationsByServer[serverIdValue]?.serverName ||
                          servers.find((s) => String(s.id) === serverIdValue)
                            ?.name ||
                          "Operações";

                        return (
                          <Select
                            value={field.value}
                            onValueChange={field.onChange}
                          >
                            <SelectTrigger className="text-primary">
                              <SelectValue placeholder="Selecione uma operação" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectGroup>
                                <SelectLabel>{serverName}</SelectLabel>
                                {opsList.map((op) => (
                                  <SelectItem className="text-primary" key={op.id} value={String(op.id)}>
                                    <div className="flex justify-between items-center">
                                      <span className="text-primary">
                                        {op.name}
                                      </span>
                                      <small className="text-primary/80">
                                        {op.type}
                                      </small>
                                    </div>
                                  </SelectItem>
                                ))}
                              </SelectGroup>
                            </SelectContent>
                          </Select>
                        );
                      }}
                    />
                    {errors.destinations?.[index]?.resourceOperationId && (
                      <p className="text-sm text-destructive">
                        {
                          errors.destinations[index]?.resourceOperationId
                            ?.message
                        }
                      </p>
                    )}
                  </div>

                  {/* Prioridade */}
                  <div className="space-y-2">
                    <Label
                      htmlFor={`dest-priority-${index}`}
                      className="text-sm font-medium text-primary"
                    >
                      Prioridade (0 = mais alta)
                    </Label>
                    <Input
                      id={`dest-priority-${index}`}
                      type="number"
                      min="0"
                      {...register(`destinations.${index}.priority`, {
                        valueAsNumber: true,
                      })}
                      className="text-primary"
                    />
                  </div>

                  {/* Habilitado */}
                  <div className="flex items-center space-x-2">
                    <Controller
                      name={`destinations.${index}.enabled`}
                      control={control}
                      render={({ field }) => (
                        <Checkbox
                          id={`dest-enabled-${index}`}
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      )}
                    />
                    <Label
                      htmlFor={`dest-enabled-${index}`}
                      className="font-medium text-sm cursor-pointer text-primary"
                    >
                      Habilitado
                    </Label>
                  </div>

                  {/* Retry Policy */}
                  <div className="border-t border-border pt-4 space-y-3">
                    <p className="font-bold text-xs uppercase text-primary tracking-widest bg-primary/10 w-fit px-2 py-0.5 rounded">
                      Configuração de Retry
                    </p>

                    <div className="space-y-2">
                      <Label
                        htmlFor={`dest-retries-${index}`}
                        className="text-sm font-medium text-primary"
                      >
                        Máximo de Tentativas
                      </Label>
                      <Input
                        id={`dest-retries-${index}`}
                        type="number"
                        min="0"
                        placeholder="3"
                        {...register(
                          `destinations.${index}.retryPolicy.maxRetries`,
                          {
                            valueAsNumber: true,
                          },
                        )}
                        className="text-primary"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label
                        htmlFor={`dest-interval-${index}`}
                        className="text-sm font-medium text-primary"
                      >
                        Intervalo de Retry (segundos)
                      </Label>
                      <Input
                        id={`dest-interval-${index}`}
                        type="number"
                        min="0"
                        placeholder="60"
                        {...register(
                          `destinations.${index}.retryPolicy.retryInterval`,
                          {
                            valueAsNumber: true,
                          },
                        )}
                        className="text-primary"
                      />
                    </div>
                  </div>

                  {/* Informações da Operação */}
                  {selectedOp && (
                    <div className="border-t border-border pt-4 space-y-2">
                      <p className="text-sm font-bold text-primary">
                        Detalhes da Operação
                      </p>
                      <div className="rounded p-3 space-y-2 text-sm bg-muted/60 border border-border">
                        <div>
                          <span className="text-primary font-medium">Tipo:</span>{" "}
                          <span className="text-primary">
                            {selectedOp.type}
                          </span>
                        </div>
                        {selectedOp.input_schema && (
                          <div>
                            <span className="text-primary font-medium">
                              Schema de Entrada:
                            </span>{" "}
                            <code className="text-xs block mt-1 text-primary bg-black/40 p-2 rounded border border-border">
                              {JSON.stringify(selectedOp.input_schema, null, 2)}
                            </code>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </Card>
              );
            })}
          </div>
        ) : (
          <Card className="p-6 bg-muted/40 border-border text-center">
            <p className="text-sm text-secondary mb-4">
              Nenhum destino adicionado ainda
            </p>
          </Card>
        )}

        {/* Select para escolher operação antes de adicionar destino */}
        <div className="flex gap-2">
          <Select
            value={selectedServerToAdd}
            onValueChange={(v) => setSelectedServerToAdd(v)}
          >
            <SelectTrigger className="w-1/3 text-primary">
              <SelectValue placeholder="Servidor" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                {servers.map((srv) => (
                  <SelectItem className="text-primary" key={srv.id} value={String(srv.id)}>
                    {srv.name}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>

          <Select
            value={selectedOpToAdd}
            onValueChange={(v) => setSelectedOpToAdd(v)}
          >
            <SelectTrigger className="flex-1 text-primary">
              <SelectValue placeholder="Operação (opcional)" />
            </SelectTrigger>
            <SelectContent>
              {selectedServerToAdd && (
                <SelectGroup>
                  {(
                    operationsCache[selectedServerToAdd] ||
                    operationsByServer[selectedServerToAdd]?.operations ||
                    []
                  ).map((op) => (
                    <SelectItem className="text-primary" key={op.id} value={String(op.id)}>
                      {op.name} ({op.type})
                    </SelectItem>
                  ))}
                </SelectGroup>
              )}
            </SelectContent>
          </Select>

          <Button
            type="button"
            variant="outline"
            onClick={handleAddDestination}
            className="w-44"
            disabled={!selectedServerToAdd}
          >
            <Plus size={16} className="mr-2" /> Adicionar
          </Button>
        </div>

        {errors.destinations && !Array.isArray(errors.destinations) && (
          <p className="text-sm text-destructive">{errors.destinations.message}</p>
        )}
      </div>
    </ScrollArea>
  );
}

