import React, { useEffect, useMemo, useState, useRef } from "react";
import { useFieldArray, useFormContext } from "react-hook-form";
import {
  Plus,
  Trash2,
  Settings2,
  Activity,
  Database,
  Box,
  ChevronDown,
  Search,
  Sparkles,
  Loader2,
  Check,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { DynamicField } from "./dynamic-field";
import type { ServerBuilderValues } from "./schemas";
import { cn } from "@/lib/utils";
import { api, type AdapterMetadata } from "@/server/server.service";
// instance-form-styles removed — using design tokens directly
import { useCredentialsStore } from "@/stores/credentials-store";
import { v4 as uuidv4 } from 'uuid';
import { JsonViewMain } from "../json-view/json-view";

interface ServerBuilderFormProps {
  metadata: AdapterMetadata[];
}

type DiscoveredResourceField = {
  name: string;
  type?: string;
};

type DiscoveredResource = {
  name: string;
  type?: string;
  config: Record<string, unknown>;
  fields: DiscoveredResourceField[];
};

const normalizeDiscoveredResource = (
  raw: any,
): DiscoveredResource | null => {
  const name = raw?.name ?? raw?.Name;
  if (typeof name !== "string" || !name.trim()) {
    return null;
  }

  const config = raw?.config ?? raw?.Config;
  const fields = raw?.fields ?? raw?.Fields;

  return {
    name,
    type: raw?.type ?? raw?.Type,
    config: config && typeof config === "object" ? config : {},
    fields: Array.isArray(fields)
      ? fields
          .map((field) => {
            const fieldName = field?.name ?? field?.Name;
            if (typeof fieldName !== "string" || !fieldName.trim()) {
              return null;
            }

            return {
              name: fieldName,
              type: field?.type ?? field?.Type,
            };
          })
          .filter(
            (field): field is DiscoveredResourceField => field !== null,
          )
      : [],
  };
};

export const ServerBuilderForm: React.FC<ServerBuilderFormProps> = ({
  metadata,
}) => {
  const { control, watch } = useFormContext<ServerBuilderValues>();
  const id = watch("server.type");
  const credentialId = watch("server.credential_id");
  const credentials = useCredentialsStore((s) => s.credentials);
  const validMetadata = useMemo(
    () => metadata.filter((item): item is AdapterMetadata => !!item && typeof item.id === "string"),
    [metadata],
  );

  const [discovering, setDiscovering] = useState(false);
  const [discoveredResources, setDiscoveredResources] = useState<
    DiscoveredResource[] | null
  >(null);

  const selectedMetadata = useMemo(
    () => validMetadata.find((m) => m.id === id),
    [validMetadata, id],
  );

  const {
    fields: resourceFields,
    append: appendResource,
    remove: removeResource,
  } = useFieldArray({
    control,
    name: "resources",
  });

  const initializedFor = useRef<string | null>(null);

  const handleAddResource = () => {
    appendResource({
      id: uuidv4(),
      name: `Novo Recurso ${resourceFields.length + 1}`,
      type: id,
      config: {},
      operations: [],
    });
  };

  const handleDiscover = async () => {
    if (!id || !credentialId) return;
    setDiscovering(true);
    try {
      const cred = credentials.find((c) => c.id === credentialId);
      if (!cred) throw new Error("Credencial não encontrada");

      const res = await api.discoverResources(id, cred.data);
      const resources = (res.data.resources || [])
        .map(normalizeDiscoveredResource)
        .filter((resource): resource is DiscoveredResource => resource !== null);
      setDiscoveredResources(resources);
      console.log("resources", res);
    } catch (e) {
      console.error(e);
      alert("Falha ao descobrir recursos: " + (e as any).message);
    } finally {
      setDiscovering(false);
    }
  };

  const quickAdd = (res: DiscoveredResource) => {
    const resourceId = uuidv4();
    const opField = selectedMetadata?.operationFields?.[0];
    const opKey = opField?.key || "operation";
    const defaultVal = opField?.options?.find(o => String(o.label).toLowerCase().includes('select') || String(o.label).toLowerCase().includes('find'))?.value 
      || opField?.options?.[0]?.value 
      || (id === "mongo" ? "find" : "select");

    appendResource({
      id: resourceId,
      name: res.name,
      type: id,
      config: res.config || {
        table: res.name, // Fallback SQL
        collection: res.name, // Fallback NoSQL
      },
      operations: [
        {
          id: uuidv4(),
          name: "Listar Registros",
          type: defaultVal,
          config: {
            [opKey]: defaultVal,
          },
        },
      ],
    });
    setDiscoveredResources((prev) =>
      prev
        ? prev.filter(
            (r) =>
              (r.config?.table || r.name) !== (res.config?.table || res.name),
          )
        : null,
    );
  };

  // Only auto-init IF nothing exists and we haven't initialized for this ID yet
  useEffect(() => {
    if (
      selectedMetadata &&
      resourceFields.length === 0 &&
      initializedFor.current !== id
    ) {
      initializedFor.current = id;
      const resId = uuidv4();
      
      const opField = selectedMetadata?.operationFields?.[0];
      const opKey = opField?.key || "operation";
      const defaultVal = opField?.options?.find(o => String(o.label).toLowerCase().includes('insert') || String(o.label).toLowerCase().includes('create'))?.value 
        || opField?.options?.[0]?.value 
        || (id === "mongo" ? "insertOne" : "insert");
        
      appendResource({
        id: resId,
        name: "Recurso Principal",
        type: id,
        config: {},
        operations: [
          {
            id: uuidv4(),
            name: "Operação Principal",
            type: defaultVal,
            config: {
              [opKey]: defaultVal,
            },
          },
        ],
      });
    }
    // If it's already initialized but for a DIFFERENT id, we might want to reset?
    // The user didn't ask for auto-reset, so I'll leave it.
  }, [id, selectedMetadata, resourceFields.length]);

  if (!id) {
    return (
      <div className="flex items-center justify-center h-64 border-2 border-dashed border border-input rounded-lg">
        <p className="text-secondary">
          Selecione um tipo de servidor e credencial para começar.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-1">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium text-primary">
            Recursos e Operações
          </h3>
          <p className="text-sm text-secondary">
            Configure as tabelas, tópicos ou endpoints que você deseja acessar.
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleDiscover}
            disabled={discovering || !credentialId}
            className="h-7 text-xs bg-secondary/10 text-secondary rounded px-2 py-1"
          >
            {discovering ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Sparkles className="w-4 h-4 mr-2" />
            )}
            Explorar Banco de Dados
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleAddResource}
            className="h-7 text-xs bg-muted border-muted-foreground/20 text-secondary"
          >
            <Plus className="w-4 h-4 mr-2" />
            Adicionar Manualmente
          </Button>
        </div>
      </div>

      {discoveredResources && discoveredResources.length > 0 && (
        <Card className="bg-muted/20 border-none text-primary mb-6">
          <CardHeader className="">
            <div className="flex items-center justify-between space-y-0">

            <div className="flex items-center gap-2">
              <Search className="w-4 h-4" />
              <h4 className="text-sm font-medium text-primary">
                Resultados da Exploração ({discoveredResources.length})
              </h4>
            </div>
            <div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setDiscoveredResources(null)}
                className="h-7 text-xs"
              >
                Fechar
              </Button>
            </div>
            </div>
          </CardHeader>
          <CardContent className="px-4 pb-4">
            <div className="flex flex-wrap gap-2">
              {discoveredResources.map((res) => (
                <Button
                  key={res.name}
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => quickAdd(res)}
                  className="h-7 text-[11px] bg-secondary/20 text-secondary rounded-lg px-2 py-1"
                >
                  <Plus className="w-3 h-3 mr-1" />
                  <div className="text-primary">{res.name}</div>
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <div className="space-y-4">
        {resourceFields.map((resource, resourceIndex) => (
          <ResourceItem
            key={resource.id}
            index={resourceIndex}
            remove={removeResource}
            metadata={selectedMetadata}
            serverType={id}
          />
        ))}
      </div>

      {resourceFields.length === 0 && !discoveredResources && (
        <div className="text-center py-12 bg-white/[0.02] border border-white/[0.05] rounded-xl">
          <Database className="w-12 h-12 mx-auto text-zinc-700 mb-4" />
          <p className="text-zinc-500">Nenhum recurso configurado ainda.</p>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleDiscover}
            className="mt-4 text-primary hover:text-primary/80"
          >
            Explorar banco de dados agora
          </Button>
        </div>
      )}
    </div>
  );
};

const ResourceItem = ({
  index,
  remove,
  metadata,
  serverType,
}: {
  index: number;
  remove: (i: number) => void;
  metadata: AdapterMetadata | undefined;
  serverType: string;
}) => {
  const { control, watch } = useFormContext<ServerBuilderValues>();
  const resourceName = watch(`resources.${index}.name`);
  const [isOpen, setIsOpen] = useState(index === 0);

  const {
    fields: operationFields,
    append: appendOperation,
    remove: removeOperation,
  } = useFieldArray({
    control,
    name: `resources.${index}.operations`,
  });

  const handleAddOperation = () => {
    const opField = metadata?.operationFields?.[0];
    const opKey = opField?.key || "operation";
    const defaultVal = opField?.options?.find(o => String(o.label).toLowerCase().includes('select') || String(o.label).toLowerCase().includes('find'))?.value 
      || opField?.options?.[0]?.value 
      || (serverType === "mongo" ? "find" : "select");

    appendOperation({
      id: uuidv4(),
      name: `Operação ${operationFields.length + 1}`,
      type: defaultVal,
      config: {
        [opKey]: defaultVal,
      },
    });
  };

  return (
    <Collapsible
      open={isOpen}
      onOpenChange={setIsOpen}
      className="bg-muted/30 text-primary-foreground rounded-xl overflow-hidden px-4"
    >
      <div className="flex items-center gap-4">
        <CollapsibleTrigger asChild>
          <button className="flex items-center gap-3 text-left py-4 flex-1 outline-none">
            <ChevronDown
              className={cn(
                "w-4 h-4 transition-transform text-primary",
                isOpen && "rotate-180",
              )}
            />
            <div className="w-8 h-8 rounded-lg flex items-center justify-center bg-primary/10 border-primary/20 text-primary-foreground">
              <Box className="w-4 h-4 text-primary" />
            </div>
            <div>
              <p className="font-medium text-primary">
                {resourceName || "Sem nome"}
              </p>
              <p className="text-[10px] text-secondary uppercase tracking-wider">
                {metadata?.name || "Recurso"}
              </p>
            </div>
          </button>
        </CollapsibleTrigger>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => remove(index)}
          className="text-secondary hover:text-destructive"
        >
          <Trash2 className="w-4 h-4" />
        </Button>
      </div>

      <CollapsibleContent className="pb-6 space-y-6 pt-2">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2 col-span-2">
            <Label className="text-secondary text-xs">
              NOME EXIBIÇÃO NO SPARK EDGE
            </Label>
            <Input
              {...control.register(`resources.${index}.name`)}
              placeholder="Ex: Tabela de Usuários"
              className={`h-10 bg-input text-secondary`}
            />
          </div>
        </div>

        {/* Dynamic Resource Fields */}
        {metadata?.resourceFields && metadata.resourceFields.length > 0 && (
          <div className="space-y-4 pt-2">
            <h4 className="text-xs font-semibold text-secondary uppercase flex items-center gap-2">
              <Settings2 className="w-3 h-3" />
              Parâmetros do Recurso
            </h4>
            <div className="grid grid-cols-2 gap-4 p-4 rounded-lg bg-foreground border-primary/20 text-primary-foreground">
              {metadata.resourceFields.map((f: any) => (
                <DynamicField
                  key={f.key}
                  field={f}
                  path={`resources.${index}.config`}
                />
              ))}
            </div>
          </div>
        )}

        {/* Operations */}
        <div className="space-y-4 pt-4 border-t border-white/[0.05]">
          <div className="flex items-center justify-between">
            <h4 className="text-xs font-semibold text-secondary uppercase flex items-center gap-2">
              <Activity className="w-3 h-3" />
              Ações Disponíveis
            </h4>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleAddOperation}
              className="h-7 text-[10px] uppercase tracking-wider bg-muted text-primary rounded-lg px-2 py-1"
            >
              <Plus className="w-3 h-3 mr-1" />
              Adicionar Ação
            </Button>
          </div>

          <div className="space-y-3">
            {operationFields.map((op, opIndex) => (
              <OperationItem
                key={op.id}
                resourceIndex={index}
                index={opIndex}
                remove={removeOperation}
                metadata={metadata}
              />
            ))}
          </div>
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
};

const OperationItem = ({
  resourceIndex,
  index,
  remove,
  metadata,
}: {
  resourceIndex: number;
  index: number;
  remove: (i: number) => void;
  metadata: AdapterMetadata | undefined;
}) => {
  const { control, setValue, watch } = useFormContext<ServerBuilderValues>();
  const inputSchema = watch(
    `resources.${resourceIndex}.operations.${index}.input_schema` as any,
  );
  const outputSchema = watch(
    `resources.${resourceIndex}.operations.${index}.output_schema` as any,
  );
  const operationType = watch(
    `resources.${resourceIndex}.operations.${index}.config.operation` as any,
  );
  const method = watch(
    `resources.${resourceIndex}.operations.${index}.config.method` as any,
  );
  const [showSchemas, setShowSchemas] = useState(false);

  useEffect(() => {
    const finalType = method || operationType;
    if (finalType) {
      setValue(
        `resources.${resourceIndex}.operations.${index}.type` as any,
        finalType,
      );
    }
  }, [operationType, method, setValue, resourceIndex, index]);

  const updateDeep = (obj: any, path: string, key: string, value: any) => {
    const newObj = JSON.parse(JSON.stringify(obj || {}));
    if (!path && key === "") return value; // Root replacement
    if (!path) {
      newObj[key] = value;
      return newObj;
    }
    const parts = path.split(".");
    let temp = newObj;
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      if (temp[part] === undefined) {
        // Guess if next level should be array or object
        const nextPart = i < parts.length - 1 ? parts[i + 1] : key;
        temp[part] = /^\d+$/.test(nextPart) ? [] : {};
      }
      temp = temp[part];
    }
    temp[key] = value;
    return newObj;
  };

  const deleteDeep = (obj: any, path: string, key: string) => {
    let newObj = JSON.parse(JSON.stringify(obj || {}));
    if (!path) {
      if (Array.isArray(newObj)) {
        newObj.splice(Number(key), 1);
      } else {
        delete newObj[key];
      }
      return newObj;
    }
    const parts = path.split(".");
    let temp = newObj;
    for (const part of parts) {
      if (!temp || temp[part] === undefined) return newObj;
      temp = temp[part];
    }
    if (Array.isArray(temp)) {
      temp.splice(Number(key), 1);
    } else {
      delete temp[key];
    }
    return newObj;
  };

  const handleSchemaChange = (
    type: "input_schema" | "output_schema",
    path: string,
    key: string,
    value: any,
  ) => {
    const current = watch(
      `resources.${resourceIndex}.operations.${index}.${type}` as any,
    );
    setValue(
      `resources.${resourceIndex}.operations.${index}.${type}` as any,
      updateDeep(current, path, key, value),
    );
  };

  const handleAddSchemaField = (
    type: "input_schema" | "output_schema",
    path: string,
    key: string,
    fieldType: any,
  ) => {
    const current = watch(
      `resources.${resourceIndex}.operations.${index}.${type}` as any,
    );
    let defaultValue: any = "";
    if (fieldType === "string") defaultValue = "string";
    if (fieldType === "number") defaultValue = "number";
    if (fieldType === "boolean") defaultValue = "boolean";
    if (fieldType === "object") defaultValue = {};
    if (fieldType === "array") defaultValue = [];

    // If we are initializing from null/empty root
    if (!current && !path) {
      setValue(
        `resources.${resourceIndex}.operations.${index}.${type}` as any,
        fieldType === "array" ? [defaultValue] : { [key]: defaultValue },
      );
      return;
    }

    let finalKey = key;
    let target = current;
    if (path) {
      const parts = path.split(".");
      for (const part of parts) target = target[part];
    }

    if (Array.isArray(target)) {
      finalKey = target.length.toString();
    }

    setValue(
      `resources.${resourceIndex}.operations.${index}.${type}` as any,
      updateDeep(current, path, finalKey, defaultValue),
    );
  };

  const handleRemoveSchemaField = (
    type: "input_schema" | "output_schema",
    path: string,
    key: string,
  ) => {
    const current = watch(
      `resources.${resourceIndex}.operations.${index}.${type}` as any,
    );
    setValue(
      `resources.${resourceIndex}.operations.${index}.${type}` as any,
      deleteDeep(current, path, key),
    );
  };

  return (
    <Card className="bg-foreground border-none text-primary-foreground">
      <CardHeader className="py-3 px-4 flex-row items-center space-y-0 gap-2">
        <div className="flex-1 flex items-center bg-muted border-muted-foreground/20 text-secondary rounded-lg px-3 group transition-colors">
          <Input
            {...control.register(
              `resources.${resourceIndex}.operations.${index}.name`,
            )}
            className={`bg-input text-secondary h-8 font-medium p-0 border-none bg-transparent focus-visible:ring-0 shadow-none text-sm placeholder:text-secondary`}
            placeholder="Nome da Ação (Ex: Listar Usuários)"
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => remove(index)}
            className={`h-7 w-7 text-secondary hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity`}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      </CardHeader>

      {metadata?.operationFields && metadata.operationFields.length > 0 && (
        <CardContent className="px-4 pb-4 grid grid-cols-2 gap-4 pt-0">
          {metadata.operationFields.map((f: any) => (
            <DynamicField
              key={f.key}
              field={f}
              path={`resources.${resourceIndex}.operations.${index}.config`}
            />
          ))}
        </CardContent>
      )}

      <div className="px-4 pb-3 border-t border-white/[0.03] pt-2">
        <button
          type="button"
          onClick={() => setShowSchemas(!showSchemas)}
          className="flex items-center gap-2 text-[10px] text-zinc-500 hover:text-zinc-300 uppercase tracking-widest font-semibold transition-colors"
        >
          <Settings2
            className={cn(
              "w-3 h-3 transition-transform",
              showSchemas && "rotate-90",
            )}
          />
          Definir Esquemas (Input/Output)
        </button>

        {showSchemas && (
          <div className="grid grid-cols-2 gap-4 mt-3 animate-in fade-in slide-in-from-top-1 duration-200">
            <div className="space-y-2">
              <Label className="text-[10px] text-secondary uppercase">
                Esquema de Entrada
              </Label>
              <div
                className={`min-h-[100px] rounded-lg p-2 bg-muted border-muted-foreground/20 text-secondary`}
              >
                <JsonViewMain
                  data={inputSchema}
                  onParamChange={(path: string, key: string, val: any) =>
                    handleSchemaChange("input_schema", path, key, val)
                  }
                  onAddField={(path: string, key: string, type: any) =>
                    handleAddSchemaField("input_schema", path, key, type)
                  }
                  onDeleteField={(path: string, key: string) =>
                    handleRemoveSchemaField("input_schema", path, key)
                  }
                  draggableValue={false}
                  pProps={{ className: "text-[10px] font-mono" }}
                />

                {!inputSchema && (
                  <div className="flex gap-1 mt-2">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className={`flex-1 text-[10px] text-secondary h-7`}
                      onClick={() =>
                        setValue(
                          `resources.${resourceIndex}.operations.${index}.input_schema` as any,
                          {},
                        )
                      }
                    >
                      {"{ }"} Objeto
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className={`flex-1 text-[10px] text-secondary h-7`}
                      onClick={() =>
                        setValue(
                          `resources.${resourceIndex}.operations.${index}.input_schema` as any,
                          [],
                        )
                      }
                    >
                      {"[ ]"} Lista
                    </Button>
                  </div>
                )}
              </div>
            </div>
            <div className="space-y-2">
              <Label className="text-[10px] text-zinc-500 uppercase">
                Esquema de Saída
              </Label>
              <div
                className={`min-h-[100px] rounded-lg p-2 bg-muted border-muted-foreground/20 text-secondary`}
              >
                <JsonViewMain
                  data={outputSchema}
                  onParamChange={(path: string, key: string, val: any) =>
                    handleSchemaChange("output_schema", path, key, val)
                  }
                  onAddField={(path: string, key: string, type: any) =>
                    handleAddSchemaField("output_schema", path, key, type)
                  }
                  onDeleteField={(path: string, key: string) =>
                    handleRemoveSchemaField("output_schema", path, key)
                  }
                  draggableValue={false}
                  pProps={{ className: "text-[10px] font-mono" }}
                />

                {!outputSchema && (
                  <div className="flex gap-1 mt-2">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className={`flex-1 text-[10px] text-secondary h-7`}
                      onClick={() =>
                        setValue(
                          `resources.${resourceIndex}.operations.${index}.output_schema` as any,
                          {},
                        )
                      }
                    >
                      {"{ }"} Objeto
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className={`flex-1 text-[10px] text-secondary h-7`}
                      onClick={() =>
                        setValue(
                          `resources.${resourceIndex}.operations.${index}.output_schema` as any,
                          [],
                        )
                      }
                    >
                      {"[ ]"} Lista
                    </Button>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </Card>
  );
};

