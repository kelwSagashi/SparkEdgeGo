import { useFormContext } from "react-hook-form";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Info, ArrowRightLeft } from "lucide-react";
import { JsonViewMain } from "../json-view/json-view";
import type {
  ResourceOperationReturningValues,
  DownloadedScriptReturningValues,
  DeviceReturningValues,
  SchemaConfigIO,
} from "@/types/db";
import type { InstanceFormValues } from "./instance-form.schemas";
import { useEffect } from "react";

interface InstanceMappingFormProps {
  allOperations: ResourceOperationReturningValues[];
  selectedScript?: DownloadedScriptReturningValues;
  selectedDevice?: DeviceReturningValues;
  includeDeviceData: boolean;
}

export function InstanceMappingForm({
  allOperations,
  selectedScript,
  selectedDevice,
  includeDeviceData,
}: InstanceMappingFormProps) {
  const {
    register,
    watch,
    setValue,
  } = useFormContext<InstanceFormValues>();
  
  const destinations = watch("destinations") || [];
  const instanceName = watch("name");

  // 1. Construir o objeto de dados de origem (Source Data) - Unificado para todos os destinos
  const sourceData: any = {
    device: selectedDevice
      ? {
          id: selectedDevice.id,
          name: selectedDevice.name,
          ip_address: selectedDevice.ip_address,
          ...(typeof selectedDevice.others === "object"
            ? (selectedDevice.others as any)
            : {}),
          include_data_on_send: includeDeviceData,
        }
      : null,
    script: {},
    system: {
      now: new Date().toISOString(),
      instance_name: instanceName || "Nova Instância",
      destinations: destinations.map((d, i) => ({
        index: i + 1,
        serverId: d.serverId,
        operationId: d.resourceOperationId
      }))
    },
  };

  // Transform outputs recursively to build dynamic source pattern
  const buildOutputTree = (fields: SchemaConfigIO[], prefix = '$.script') => {
    const tree: any = {};
    fields.forEach((field) => {
      if (field.fields && Array.isArray(field.fields) && field.fields.length > 0) {
        tree[field.name] = buildOutputTree(field.fields, `${prefix}.${field.name}`);
      } else {
        tree[field.name] = `{{${prefix}.${field.name}}}`;
      }
    });
    return tree;
  };

  const outputs = Array.isArray(selectedScript?.schema_config?.outputs) 
    ? selectedScript?.schema_config.outputs 
    : [];

  sourceData.script = buildOutputTree(outputs);


  // 2. Efeito para inicializar o payloadTemplate baseado no esquema da operação
  useEffect(() => {
    if (!allOperations || allOperations.length === 0 || !destinations) return;

    destinations.forEach((dest, index) => {
      if (!dest.resourceOperationId) return;

      const operation = allOperations.find(
        (op) => op.id === dest.resourceOperationId,
      );
      if (!operation) return;

      // Tentar obter o esquema de entrada (pode vir como string do banco)
      let inputSchema = operation.input_schema;
      if (typeof inputSchema === "string") {
        try {
          inputSchema = JSON.parse(inputSchema);
        } catch (e) {
          return;
        }
      }

      const properties = inputSchema?.properties || inputSchema;
      if (!properties || typeof properties !== "object" || Array.isArray(properties))
        return;

      // Verificar se precisamos inicializar
      const currentPayload = dest.dataMapping?.payloadTemplate;
      const isEmpty =
        !currentPayload || Object.keys(currentPayload).length === 0;

      if (isEmpty) {
        const template: any = {};
        Object.entries(properties).forEach(([key, schema]: [string, any]) => {
          // Suporta tanto formato JSON Schema quanto formato simplificado { field: 'type' }
          const schemaObj =
            schema && typeof schema === "object" ? schema : { type: schema };
          const type = schemaObj.type || "string";

          template[key] =
            type === "number"
              ? 0
              : type === "boolean"
                ? false
                : type === "object"
                  ? {}
                  : type === "array"
                    ? []
                    : "";
        });
        setValue(`destinations.${index}.dataMapping.payloadTemplate`, template, {
          shouldValidate: true,
          shouldDirty: true,
        });
      }
    });
  }, [
    allOperations,
    setValue,
    // biome-ignore lint/correctness/useExhaustiveDependencies: Gatilho principal por stringify
    JSON.stringify(destinations?.map((d) => d.resourceOperationId)),
  ]);

  // 3. Construir o objeto de payload combinado (Unified Payload)
  const combinedPayload: Record<string, any> = {};
  destinations.forEach((dest, idx) => {
    const operation = allOperations.find(
      (op) => op.id === dest.resourceOperationId,
    );
    const key = `${idx + 1}. ${operation?.name || "Destino"}`;
    combinedPayload[key] = dest.dataMapping?.payloadTemplate || {};
  });

  // 4. Helpers para sincronizar as mudanças de volta para o array de destinos original
  const getDestIndexFromPath = (path: string) => {
    // O path sempre começa com "N. " onde N é o índice + 1
    const match = path.match(/^(\d+)\./);
    if (match) return parseInt(match[1], 10) - 1;
    return -1;
  };

  const getSubPath = (fullPath: string) => {
    const destIdx = getDestIndexFromPath(fullPath);
    if (destIdx === -1) return "";
    
    // O prefixo é dinâmico baseado no nome da operação
    const op = allOperations.find(o => o.id === destinations[destIdx]?.resourceOperationId);
    const prefix = `${destIdx + 1}. ${op?.name || "Destino"}`;
    
    if (fullPath === prefix) return "";
    if (fullPath.startsWith(prefix + ".")) {
       return fullPath.substring(prefix.length + 1);
    }
    return "";
  };

  // Helper para atualizar objeto aninhado por path
  const updateNestedValue = (obj: any, path: string, value: any) => {
    if (!path) return value;
    const keys = path.split(".");
    const current = JSON.parse(JSON.stringify(obj || {}));
    let ref = current;
    for (let i = 0; i < keys.length - 1; i++) {
      const key = keys[i];
      if (!ref[key] || typeof ref[key] !== 'object') {
        ref[key] = {};
      }
      ref = ref[key];
    }
    ref[keys[keys.length - 1]] = value;
    return current;
  };

  // Helper para deletar campo por path
  const deleteNestedValue = (obj: any, path: string) => {
    if (!path) return {};
    const keys = path.split(".");
    const lastKey = keys.pop();
    const current = JSON.parse(JSON.stringify(obj || {}));
    let ref = current;
    for (let i = 0; i < keys.length; i++) {
      if (!ref[keys[i]]) return obj;
      ref = ref[keys[i]];
    }
    if (lastKey) delete ref[lastKey];
    return current;
  };

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-8">
        {/* Header */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
             <ArrowRightLeft className="w-4 h-4 text-violet-400" />
             <h3 className="font-medium text-primary uppercase text-xs tracking-wider">Mapeamento Unificado de Destinos</h3>
          </div>
          <p className="text-sm text-secondary">
            Mapeie os dados para todos os destinos em uma única visualização. 
            Arraste os nomes dos campos da esquerda para os valores da direita.
          </p>
        </div>

        {(!destinations || destinations.length === 0) ? (
          <Card className="p-6 bg-muted/40 border-border text-secondary text-center">
            <p className="text-sm">
              Adicione destinos primeiro para configurar o mapeamento de dados.
            </p>
          </Card>
        ) : (
          <div className="space-y-8">
            {/* GRID PRINCIPAL DE MAPEAMENTO */}
            <div className="grid grid-cols-2 gap-4 h-[500px]">
              {/* LADO ESQUERDO: FONTE DE DADOS UNIFICADA */}
              <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0 scroll-smooth scrollbar-thin scrollbar-thumb-border scrollbar-track-transparent">
                <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                  <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">Fonte de Dados Disponível</span>
                  <span className="text-[9px] px-1.5 py-0.5 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20">Read Only</span>
                </div>
                <div className="flex-1 overflow-auto p-2">
                    <JsonViewMain
                      data={sourceData}
                      draggableValue={true}
                      rootPath="$"
                      pProps={{
                        className: "bg-transparent border-none text-xs text-primary"
                      }}
                    />
                </div>
              </Card>

              {/* LADO DIREITO: PAYLOAD GLOBAL (TODOS OS DESTINOS) */}
              <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0 scroll-smooth scrollbar-thin scrollbar-thumb-border scrollbar-track-transparent">
                <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                  <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">Payload de Todos os Destinos</span>
                  <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">Unified View</span>
                </div>
                <div className="flex-1 overflow-auto p-2">
                    <JsonViewMain
                      data={combinedPayload}
                      onParamChange={(path, key, val) => {
                        const fullPath = path ? `${path}.${key}` : key;
                        const destIdx = getDestIndexFromPath(fullPath);
                        if (destIdx === -1) return;
                        
                        const subPath = getSubPath(fullPath);
                        if (!subPath) return; // Não permitir editar a raiz do destino

                        const currentTemplate = destinations[destIdx].dataMapping?.payloadTemplate || {};
                        setValue(`destinations.${destIdx}.dataMapping.payloadTemplate`, updateNestedValue(currentTemplate, subPath, val), { 
                          shouldValidate: true,
                          shouldDirty: true 
                        });
                      }}
                      onAddField={(path, key, type) => {
                        const fullPath = path ? `${path}.${key}` : key;
                        const destIdx = getDestIndexFromPath(fullPath);
                        if (destIdx === -1) return;
                        
                        const subPath = getSubPath(fullPath);
                        if (!subPath) return;

                        const defaultVal = type === 'object' ? {} : type === 'array' ? [] : type === 'number' ? 0 : type === 'boolean' ? false : "";
                        const currentTemplate = destinations[destIdx].dataMapping?.payloadTemplate || {};
                        setValue(`destinations.${destIdx}.dataMapping.payloadTemplate`, updateNestedValue(currentTemplate, subPath, defaultVal), { 
                          shouldValidate: true,
                          shouldDirty: true
                        });
                      }}
                      onDeleteField={(path, key) => {
                        const fullPath = path ? `${path}.${key}` : key;
                        const destIdx = getDestIndexFromPath(fullPath);
                        if (destIdx === -1) return;
                        
                        const subPath = getSubPath(fullPath);
                        if (!subPath) return;

                        const currentTemplate = destinations[destIdx].dataMapping?.payloadTemplate || {};
                        setValue(`destinations.${destIdx}.dataMapping.payloadTemplate`, deleteNestedValue(currentTemplate, subPath), { 
                          shouldValidate: true,
                          shouldDirty: true
                        });
                      }}
                      onDestructure={(path, key, val) => {
                        const fullPath = path ? `${path}.${key}` : key;
                        const destIdx = getDestIndexFromPath(fullPath);
                        if (destIdx === -1) return;
                        
                        const subPath = getSubPath(fullPath);
                        // Subpath pode ser vazio se estiver desestruturando na raiz do destino
                        const currentTemplate = { ...(destinations[destIdx].dataMapping?.payloadTemplate || {}) };
                        
                        Object.entries(val).forEach(([childKey, childValue]) => {
                           const childPath = subPath ? `${subPath}.${childKey}` : childKey;
                           const updated = updateNestedValue(currentTemplate, childPath, childValue);
                           // Sincronizar de volta para o currentTemplate local para as próximas iterações
                           Object.assign(currentTemplate, updated);
                        });
                        
                        setValue(`destinations.${destIdx}.dataMapping.payloadTemplate`, currentTemplate, { 
                          shouldValidate: true,
                          shouldDirty: true
                        });
                      }}
                      draggableValue={false}
                      pProps={{
                        className: "bg-transparent border-none text-xs text-primary"
                      }}
                    />
                </div>
              </Card>
            </div>

            {/* SEÇÃO DE SCRIPTS DE TRANFORMAÇÃO (POR DESTINO) */}
            <div className="space-y-4">
               <h4 className="text-[10px] font-bold text-secondary uppercase tracking-widest px-1">Scripts de Pós-Processamento</h4>
               <div className="grid grid-cols-1  gap-4">
                  {destinations.map((dest: any, idx: number) => {
                    const op = allOperations.find(o => o.id === dest.resourceOperationId);
                    return (
                      <Card key={idx} className="p-4 bg-foreground border-border space-y-3">
                         <div className="flex items-center gap-2">
                           <span className="w-5 h-5 rounded-full bg-violet-500/10 text-violet-400 flex items-center justify-center text-[9px] font-bold border border-violet-500/20">
                             {idx + 1}
                           </span>
                           <Label className="text-xs font-semibold text-primary">
                             {op?.name || "Destino"}
                           </Label>
                         </div>
                         <Textarea
                           placeholder="Ex: return { ...data, timestamp: Date.now() }"
                           {...register(`destinations.${idx}.dataMapping.transformScript`)}
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

        {/* Footer Info */}
        <div className="p-4 bg-muted/20 border border-border rounded-lg flex items-start gap-3">
          <Info className="w-4 h-4 text-violet-400 mt-0.5" />
          <div className="space-y-1">
             <p className="text-xs text-primary font-medium">Dica de Mapeamento</p>
             <p className="text-[11px] text-secondary leading-relaxed">
                Cada destino aparece como um objeto numerado no painel à direita. 
                Mudanças feitas em cada sub-objeto serão salvas automaticamente no respectivo destino.
             </p>
          </div>
        </div>
      </div>
    </ScrollArea>
  );
}

