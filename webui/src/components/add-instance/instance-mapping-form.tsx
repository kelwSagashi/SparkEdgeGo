import { useEffect } from "react";
import { useFormContext } from "react-hook-form";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { ArrowRightLeft, Info } from "lucide-react";
import { JsonViewMain } from "../json-view/json-view";
import type {
  DeviceReturningValues,
  DownloadedScriptReturningValues,
  ResourceOperationReturningValues,
  SchemaConfigIO,
} from "@/types/db";
import type { InstanceFormValues } from "./instance-form.schemas";
import { buildInstanceRuntimePreview } from "@/lib/instance-mapping";

interface InstanceMappingFormProps {
  allOperations: ResourceOperationReturningValues[];
  selectedScript?: DownloadedScriptReturningValues;
  selectedDevice?: DeviceReturningValues;
  includeDeviceData: boolean;
  instanceId?: string;
}

function inferTemplateDefault(schema: unknown): unknown {
  if (Array.isArray(schema)) {
    return [];
  }

  if (schema && typeof schema === "object") {
    const schemaObject = schema as Record<string, unknown>;
    const explicitType = schemaObject.type;

    if (explicitType === "number" || explicitType === "integer") {
      return 0;
    }
    if (explicitType === "boolean") {
      return false;
    }
    if (explicitType === "array") {
      return [];
    }
    if (explicitType === "object") {
      return {};
    }
    if ("properties" in schemaObject && schemaObject.properties && typeof schemaObject.properties === "object") {
      return {};
    }
    if (Object.keys(schemaObject).length === 0) {
      return {};
    }
  }

  if (schema === "number" || schema === "integer") {
    return 0;
  }
  if (schema === "boolean") {
    return false;
  }
  if (schema === "array") {
    return [];
  }
  if (schema === "object") {
    return {};
  }

  return "";
}

export function InstanceMappingForm({
  allOperations,
  selectedScript,
  selectedDevice,
  includeDeviceData,
  instanceId,
}: InstanceMappingFormProps) {
  const { register, watch, setValue } = useFormContext<InstanceFormValues>();

  const destinations = watch("destinations") || [];
  const instanceName = watch("name");
  const deviceId = watch("device_id");

  const buildOutputTree = (fields: SchemaConfigIO[], prefix = "$.script") => {
    const tree: any = {};
    fields.forEach((field) => {
      if (!field.name) {
        return;
      }
      if (field.fields && field.fields.length > 0) {
        tree[field.name] = buildOutputTree(field.fields, `${prefix}.${field.name}`);
        return;
      }
      tree[field.name] = `{{${prefix}.${field.name}}}`;
    });
    return tree;
  };

  const outputs = Array.isArray(selectedScript?.schema_config?.outputs)
    ? selectedScript.schema_config.outputs
    : [];

  const sourceData = buildInstanceRuntimePreview({
    instanceName,
    instanceId,
    deviceId,
    selectedDevice,
    includeDeviceData,
    destinations,
    scriptOutput: buildOutputTree(outputs),
  });

  useEffect(() => {
    if (!allOperations.length || !destinations.length) {
      return;
    }

    destinations.forEach((destination, index) => {
      if (!destination.resourceOperationId) {
        return;
      }

      const operation = allOperations.find(
        (item) => String(item.id) === String(destination.resourceOperationId),
      );
      if (!operation) {
        return;
      }

      let inputSchema = operation.input_schema;
      if (typeof inputSchema === "string") {
        try {
          inputSchema = JSON.parse(inputSchema);
        } catch {
          return;
        }
      }

      const properties = inputSchema?.properties || inputSchema;
      if (!properties || typeof properties !== "object" || Array.isArray(properties)) {
        return;
      }

      const currentPayload = destination.dataMapping?.payloadTemplate;
      if (currentPayload && Object.keys(currentPayload).length > 0) {
        return;
      }

      const template: Record<string, unknown> = {};
      Object.entries(properties).forEach(([key, schema]: [string, any]) => {
        const schemaObject =
          schema && typeof schema === "object" ? schema : { type: schema };
        template[key] = inferTemplateDefault(schemaObject);
      });

      setValue(`destinations.${index}.dataMapping.payloadTemplate`, template, {
        shouldValidate: true,
        shouldDirty: true,
      });
    });
  }, [allOperations, destinations, setValue]);

  const combinedPayload: Record<string, unknown> = {};
  destinations.forEach((destination, index) => {
    const operation = allOperations.find(
      (item) => String(item.id) === String(destination.resourceOperationId),
    );
    const key = `${index + 1}. ${operation?.name || "Destino"}`;
    combinedPayload[key] = destination.dataMapping?.payloadTemplate || {};
  });

  const getDestinationIndex = (path: string) => {
    const match = path.match(/^(\d+)\./);
    return match ? Number.parseInt(match[1], 10) - 1 : -1;
  };

  const getDestinationSubPath = (fullPath: string) => {
    const destinationIndex = getDestinationIndex(fullPath);
    if (destinationIndex === -1) {
      return "";
    }

    const operation = allOperations.find(
      (item) =>
        String(item.id) === String(destinations[destinationIndex]?.resourceOperationId),
    );
    const prefix = `${destinationIndex + 1}. ${operation?.name || "Destino"}`;
    if (fullPath === prefix) {
      return "";
    }
    if (fullPath.startsWith(`${prefix}.`)) {
      return fullPath.slice(prefix.length + 1);
    }
    return "";
  };

  const updateNestedValue = (obj: any, path: string, value: any) => {
    if (!path) {
      return value;
    }
    const keys = path.split(".");
    const current = JSON.parse(JSON.stringify(obj || {}));
    let ref = current;
    for (let i = 0; i < keys.length - 1; i++) {
      if (!ref[keys[i]] || typeof ref[keys[i]] !== "object") {
        ref[keys[i]] = {};
      }
      ref = ref[keys[i]];
    }
    ref[keys[keys.length - 1]] = value;
    return current;
  };

  const deleteNestedValue = (obj: any, path: string) => {
    if (!path) {
      return {};
    }
    const keys = path.split(".");
    const lastKey = keys.pop();
    const current = JSON.parse(JSON.stringify(obj || {}));
    let ref = current;
    for (let i = 0; i < keys.length; i++) {
      if (!ref[keys[i]]) {
        return obj;
      }
      ref = ref[keys[i]];
    }
    if (lastKey) {
      delete ref[lastKey];
    }
    return current;
  };

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-8">
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <ArrowRightLeft className="w-4 h-4 text-violet-400" />
            <h3 className="font-medium text-primary uppercase text-xs tracking-wider">
              Mapeamento Unificado de Destinos
            </h3>
          </div>
          <p className="text-sm text-secondary">
            Mapeie os dados para todos os destinos em uma unica visualizacao.
            Arraste os campos da esquerda para os valores da direita.
          </p>
        </div>

        {!destinations.length ? (
          <Card className="p-6 bg-muted/40 border-border text-secondary text-center">
            <p className="text-sm">
              Adicione destinos primeiro para configurar o mapeamento de dados.
            </p>
          </Card>
        ) : (
          <div className="space-y-8">
            <div className="grid grid-cols-2 gap-4 h-[500px]">
              <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0">
                <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                  <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">
                    Fonte de Dados Disponivel
                  </span>
                  <span className="text-[9px] px-1.5 py-0.5 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20">
                    Read Only
                  </span>
                </div>
                <div className="flex-1 overflow-auto p-2">
                  <JsonViewMain
                    data={sourceData}
                    draggableValue={true}
                    rootPath="$"
                    pProps={{
                      className: "bg-transparent border-none text-xs text-primary",
                    }}
                  />
                </div>
              </Card>

              <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0">
                <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                  <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">
                    Payload de Todos os Destinos
                  </span>
                  <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    Unified View
                  </span>
                </div>
                <div className="flex-1 overflow-auto p-2">
                  <JsonViewMain
                    data={combinedPayload}
                    templateContext={sourceData as Record<string, unknown>}
                    onParamChange={(path, key, value) => {
                      const fullPath = path ? `${path}.${key}` : key;
                      const destinationIndex = getDestinationIndex(fullPath);
                      if (destinationIndex === -1) {
                        return;
                      }

                      const subPath = getDestinationSubPath(fullPath);
                      if (!subPath) {
                        return;
                      }

                      const currentTemplate =
                        destinations[destinationIndex].dataMapping?.payloadTemplate || {};
                      setValue(
                        `destinations.${destinationIndex}.dataMapping.payloadTemplate`,
                        updateNestedValue(currentTemplate, subPath, value),
                        { shouldValidate: true, shouldDirty: true },
                      );
                    }}
                    onAddField={(path, key, type) => {
                      const fullPath = path ? `${path}.${key}` : key;
                      const destinationIndex = getDestinationIndex(fullPath);
                      if (destinationIndex === -1) {
                        return;
                      }

                      const subPath = getDestinationSubPath(fullPath);
                      if (!subPath) {
                        return;
                      }

                      const defaultValue =
                        type === "object"
                          ? {}
                          : type === "array"
                            ? []
                            : type === "number"
                              ? 0
                              : type === "boolean"
                                ? false
                                : "";
                      const currentTemplate =
                        destinations[destinationIndex].dataMapping?.payloadTemplate || {};
                      setValue(
                        `destinations.${destinationIndex}.dataMapping.payloadTemplate`,
                        updateNestedValue(currentTemplate, subPath, defaultValue),
                        { shouldValidate: true, shouldDirty: true },
                      );
                    }}
                    onDeleteField={(path, key) => {
                      const fullPath = path ? `${path}.${key}` : key;
                      const destinationIndex = getDestinationIndex(fullPath);
                      if (destinationIndex === -1) {
                        return;
                      }

                      const subPath = getDestinationSubPath(fullPath);
                      if (!subPath) {
                        return;
                      }

                      const currentTemplate =
                        destinations[destinationIndex].dataMapping?.payloadTemplate || {};
                      setValue(
                        `destinations.${destinationIndex}.dataMapping.payloadTemplate`,
                        deleteNestedValue(currentTemplate, subPath),
                        { shouldValidate: true, shouldDirty: true },
                      );
                    }}
                    onDestructure={(path, key, value) => {
                      const fullPath = path ? `${path}.${key}` : key;
                      const destinationIndex = getDestinationIndex(fullPath);
                      if (destinationIndex === -1 || typeof value !== "object" || value === null) {
                        return;
                      }

                      const subPath = getDestinationSubPath(fullPath);
                      let currentTemplate = {
                        ...(destinations[destinationIndex].dataMapping?.payloadTemplate || {}),
                      };

                      Object.entries(value).forEach(([childKey, childValue]) => {
                        const childPath = subPath ? `${subPath}.${childKey}` : childKey;
                        currentTemplate = updateNestedValue(
                          currentTemplate,
                          childPath,
                          childValue,
                        );
                      });

                      setValue(
                        `destinations.${destinationIndex}.dataMapping.payloadTemplate`,
                        currentTemplate,
                        { shouldValidate: true, shouldDirty: true },
                      );
                    }}
                    draggableValue={false}
                    pProps={{
                      className: "bg-transparent border-none text-xs text-primary",
                    }}
                  />
                </div>
              </Card>
            </div>

            <div className="space-y-4">
              <h4 className="text-[10px] font-bold text-secondary uppercase tracking-widest px-1">
                Scripts de Pos-Processamento
              </h4>
              <div className="grid grid-cols-1 gap-4">
                {destinations.map((destination: any, index: number) => {
                  const operation = allOperations.find(
                    (item) =>
                      String(item.id) === String(destination.resourceOperationId),
                  );
                  return (
                    <Card
                      key={`${destination.resourceOperationId}-${index}`}
                      className="p-4 bg-foreground border-border space-y-3"
                    >
                      <div className="flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-violet-500/10 text-violet-400 flex items-center justify-center text-[9px] font-bold border border-violet-500/20">
                          {index + 1}
                        </span>
                        <Label className="text-xs font-semibold text-primary">
                          {operation?.name || "Destino"}
                        </Label>
                      </div>
                      <Textarea
                        placeholder="Ex: return { ...data, timestamp: Date.now() }"
                        {...register(`destinations.${index}.dataMapping.transformScript`)}
                        rows={3}
                        className="font-mono text-[10px] text-primary bg-black/40 border-border focus:ring-violet-500/30"
                      />
                    </Card>
                  );
                })}
              </div>
            </div>
          </div>
        )}

        <div className="p-4 bg-muted/20 border border-border rounded-lg flex items-start gap-3">
          <Info className="w-4 h-4 text-violet-400 mt-0.5" />
          <div className="space-y-1">
            <p className="text-xs text-primary font-medium">Dica de Mapeamento</p>
            <p className="text-[11px] text-secondary leading-relaxed">
              O painel da esquerda reflete o mesmo contexto usado pelo runtime:
              `system`, `instance`, `device`, `script` e `output`.
            </p>
          </div>
        </div>
      </div>
    </ScrollArea>
  );
}
