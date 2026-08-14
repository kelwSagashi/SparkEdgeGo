import { useEffect, useMemo } from "react";
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
import { resolveTemplatePreview } from "@/lib/template-preview";

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

  const combinedPayload = useMemo(() => {
    const payload: Record<string, unknown> = {};
    destinations.forEach((destination, index) => {
      const operation = allOperations.find(
        (item) => String(item.id) === String(destination.resourceOperationId),
      );
      const key = `${index + 1}. ${operation?.name || "Destino"}`;
      payload[key] = destination.dataMapping?.payloadTemplate || {};
    });
    return payload;
  }, [allOperations, destinations]);

  const resolvedPayload = useMemo(() => {
    return resolveTemplatePreview(combinedPayload, sourceData as Record<string, unknown>);
  }, [combinedPayload, sourceData]);

  const mappingStats = useMemo(() => {
    const visit = (input: unknown): { placeholders: number; textTemplates: number; objects: number } => {
      if (typeof input === "string") {
        const placeholders = (input.match(/\{\{/g) || []).length;
        return {
          placeholders,
          textTemplates: placeholders > 0 && input.trim() !== "" ? 1 : 0,
          objects: 0,
        };
      }

      if (Array.isArray(input)) {
        return input.reduce(
          (acc, item) => {
            const nested = visit(item);
            return {
              placeholders: acc.placeholders + nested.placeholders,
              textTemplates: acc.textTemplates + nested.textTemplates,
              objects: acc.objects + nested.objects,
            };
          },
          { placeholders: 0, textTemplates: 0, objects: 0 },
        );
      }

      if (input && typeof input === "object") {
        return Object.values(input as Record<string, unknown>).reduce(
          (acc, item) => {
            const nested = visit(item);
            return {
              placeholders: acc.placeholders + nested.placeholders,
              textTemplates: acc.textTemplates + nested.textTemplates,
              objects: acc.objects + nested.objects + 1,
            };
          },
          { placeholders: 0, textTemplates: 0, objects: 0 },
        );
      }

      return { placeholders: 0, textTemplates: 0, objects: 0 };
    };

    return visit(combinedPayload);
  }, [combinedPayload]);

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

  const defaultValueForFieldType = (type: "string" | "number" | "boolean" | "object" | "array") => {
    if (type === "object") {
      return {};
    }
    if (type === "array") {
      return [];
    }
    if (type === "number") {
      return 0;
    }
    if (type === "boolean") {
      return false;
    }
    return "";
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
            <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
              <Card className="p-3 bg-muted/30 border-border">
                <p className="text-[10px] font-bold uppercase tracking-widest text-secondary">
                  Destinos
                </p>
                <p className="mt-1 text-2xl font-semibold text-primary">{destinations.length}</p>
                <p className="text-[11px] text-secondary">
                  payloads configurados na instancia
                </p>
              </Card>
              <Card className="p-3 bg-muted/30 border-border">
                <p className="text-[10px] font-bold uppercase tracking-widest text-secondary">
                  Placeholders
                </p>
                <p className="mt-1 text-2xl font-semibold text-primary">{mappingStats.placeholders}</p>
                <p className="text-[11px] text-secondary">
                  referencias e expressoes detectadas
                </p>
              </Card>
              <Card className="p-3 bg-muted/30 border-border">
                <p className="text-[10px] font-bold uppercase tracking-widest text-secondary">
                  Estruturas
                </p>
                <p className="mt-1 text-2xl font-semibold text-primary">{mappingStats.objects}</p>
                <p className="text-[11px] text-secondary">
                  objetos/layers no template de payload
                </p>
              </Card>
            </div>

            <div className="grid grid-cols-1 gap-4 xl:grid-cols-3 h-[720px]">
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
                  <div className="space-y-3">
                    {destinations.map((destination: any, destinationIndex: number) => {
                      const operation = allOperations.find(
                        (item) =>
                          String(item.id) === String(destination.resourceOperationId),
                      );
                      const currentTemplate =
                        destination.dataMapping?.payloadTemplate || {};
                      const setTemplate = (template: unknown) => {
                        setValue(
                          `destinations.${destinationIndex}.dataMapping.payloadTemplate`,
                          template,
                          { shouldValidate: true, shouldDirty: true },
                        );
                      };

                      return (
                        <div
                          key={`${destination.resourceOperationId || "destination"}-${destinationIndex}`}
                          className="rounded-md border border-border/70 bg-black/15 p-2"
                        >
                          <div className="mb-2 flex items-center justify-between gap-2">
                            <span className="truncate text-[10px] font-bold uppercase tracking-widest text-secondary">
                              {destinationIndex + 1}. {operation?.name || "Destino"}
                            </span>
                            <span className="shrink-0 text-[9px] text-secondary">
                              {Object.keys(currentTemplate).length} campos
                            </span>
                          </div>
                          <JsonViewMain
                            data={currentTemplate}
                            templateContext={sourceData as Record<string, unknown>}
                            onParamChange={(path, key, value) => {
                              const subPath = path ? `${path}.${key}` : key;
                              setTemplate(updateNestedValue(currentTemplate, subPath, value));
                            }}
                            onAddField={(path, key, type) => {
                              const subPath = path ? `${path}.${key}` : key;
                              setTemplate(
                                updateNestedValue(
                                  currentTemplate,
                                  subPath,
                                  defaultValueForFieldType(type),
                                ),
                              );
                            }}
                            onDeleteField={(path, key) => {
                              const subPath = path ? `${path}.${key}` : key;
                              setTemplate(deleteNestedValue(currentTemplate, subPath));
                            }}
                            onDestructure={(path, key, value) => {
                              if (typeof value !== "object" || value === null) {
                                return;
                              }

                              let nextTemplate = { ...currentTemplate };
                              Object.entries(value).forEach(([childKey, childValue]) => {
                                const childPath = path
                                  ? `${path}.${key}.${childKey}`
                                  : `${key}.${childKey}`;
                                nextTemplate = updateNestedValue(
                                  nextTemplate,
                                  childPath,
                                  childValue,
                                );
                              });
                              setTemplate(nextTemplate);
                            }}
                            draggableValue={false}
                            pProps={{
                              className: "bg-transparent border-none text-xs text-primary",
                            }}
                          />
                        </div>
                      );
                    })}
                  </div>
                </div>
              </Card>

              <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0">
                <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                  <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">
                    Preview Resolvido
                  </span>
                  <span className="text-[9px] px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20">
                    Runtime View
                  </span>
                </div>
                <div className="px-3 py-2 border-b border-border bg-black/20 space-y-1">
                  <p className="text-[11px] text-primary">
                    O preview usa o mesmo contexto exibido na coluna da esquerda.
                  </p>
                  <p className="text-[10px] text-secondary">
                    Misture texto com variaveis, solte objetos inteiros e confira o resultado final antes de salvar.
                  </p>
                </div>
                <div className="flex-1 overflow-auto p-2">
                  <JsonViewMain
                    data={resolvedPayload}
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
