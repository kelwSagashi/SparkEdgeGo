import { useEffect } from "react";
import { useFormContext, Controller } from "react-hook-form";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card } from "@/components/ui/card";
import { Play, ArrowRightLeft, Info, Settings2 } from "lucide-react";
import { JsonViewMain } from "../json-view/json-view";
import type {
  DownloadedScriptReturningValues,
  DeviceReturningValues,
} from "@/types/db";
import type { InstanceFormValues } from "./instance-form.schemas";

interface InstanceScriptFormProps {
  scripts: DownloadedScriptReturningValues[];
  selectedDevice?: DeviceReturningValues;
  includeDeviceData: boolean;
}

export function InstanceScriptForm({
  scripts,
  selectedDevice,
  includeDeviceData,
}: InstanceScriptFormProps) {
  const {
    control,
    watch,
    setValue,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();

  const selectedScriptId = watch("script_id");
  const scriptInputs = watch("scriptInputs");
  const instanceName = watch("name");

  const selectedScript = scripts.find((s) => s.id === selectedScriptId);

  // 1. Construir o objeto de dados de origem (Source Data)
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
    system: {
      now: new Date().toISOString(),
      instance_name: instanceName || "Nova Instância",
    },
  };

  // Se o script selecionado já tiver um esquema de saída (stdout) salvo, mostramos na fonte para mapear entre scripts se necessário
  // (Embora aqui seja entrada do script, talvez o usuário queira mapear saídas de outros scripts no futuro. 
  // Por enquanto, focamos em device e system).

  // 2. Inicializar campos se não existirem no scriptInputs baseado no schema
  useEffect(() => {
    if (selectedScript?.schema_config?.inputs) {
      const currentInputs = { ...(scriptInputs || {}) };
      let changed = false;

      selectedScript.schema_config.inputs.forEach((inp: any) => {
        if (currentInputs[inp.name] === undefined) {
          currentInputs[inp.name] =
            inp.type === "number"
              ? 0
              : inp.type === "boolean"
              ? false
              : inp.type === "object" || inp.type === "json"
              ? {}
              : "";
          changed = true;
        }
      });

      if (changed) {
        setValue("scriptInputs", currentInputs);
      }
    }
  }, [selectedScriptId, setValue]);

  // Helper para atualizar objeto aninhado por path
  const updateNestedValue = (obj: any, path: string, value: any) => {
    if (!path) return { ...obj };
    const keys = path.split(".");
    let current = JSON.parse(JSON.stringify(obj || {}));
    let ref = current;
    for (let i = 0; i < keys.length - 1; i++) {
        ref[keys[i]] = ref[keys[i]] || {};
        ref = ref[keys[i]];
    }
    ref[keys[keys.length - 1]] = value;
    return current;
  };

  // Helper para deletar campo por path
  const deleteNestedValue = (obj: any, path: string) => {
    if (!path) return { ...obj };
    const keys = path.split(".");
    const lastKey = keys.pop();
    let current = JSON.parse(JSON.stringify(obj || {}));
    let ref = current;
    for (let i = 0; i < keys.length; i++) {
      if (!ref[keys[i]]) return obj;
      ref = ref[keys[i]];
    }
    if (lastKey) delete ref[lastKey];
    return current;
  };

  // Helper para desestruturar objeto
  const handleDestructure = (path: string, key: string, value: any) => {
    if (typeof value !== 'object' || value === null) return;
    
    let currentInputs = { ...(scriptInputs || {}) };
    const parentPath = path;
    
    // Pegar todas as chaves do objeto e adicionar ao nível atual (ou pai)
    Object.entries(value).forEach(([childKey, childValue]) => {
        currentInputs = updateNestedValue(currentInputs, parentPath ? `${parentPath}.${childKey}` : childKey, childValue);
    });
    
    setValue("scriptInputs", currentInputs);
  };

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-6">
        <div className="md:col-span-1 space-y-4">
          <div className="space-y-1">
            <Label className="text-[10px] font-bold text-primary uppercase tracking-wider">
              Script
            </Label>
            <Controller
              name="script_id"
              control={control}
              render={({ field }) => (
                <Select onValueChange={field.onChange} value={field.value}>
                  <SelectTrigger className="border-border text-primary h-9">
                    <SelectValue placeholder="Selecione um script" />
                  </SelectTrigger>
                  <SelectContent className="border-border text-primary">
                    {scripts.map((script) => (
                      <SelectItem key={script.id} value={script.id}>
                        {script.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
            {errors.script_id && (
              <p className="text-[10px] text-red-400 mt-1">
                {errors.script_id.message}
              </p>
            )}
          </div>

          {selectedScript && (
            <Card className="p-3 bg-violet-500/5 border-violet-500/20">
              <div className="flex items-center gap-2">
                <Play className="w-3 h-3 text-violet-400" />
                <span className="text-xs font-semibold text-violet-300">
                  {selectedScript.name}
                </span>
                <span className="text-[8px] px-1.5 py-0.5 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20">
                  {selectedScript.language || "python"}
                </span>
                <span className="text-[8px] px-1.5 py-0.5 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20">
                  v{selectedScript.version || "1.0.0"}
                </span>
              </div>
              <p className="text-[10px] text-violet-300/70 line-clamp-2">
                {selectedScript.description || "Sem descrição disponível."}
              </p>
            </Card>
          )}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="md:col-span-2 space-y-4">
            <div className="flex items-center gap-2 mb-2">
              <ArrowRightLeft className="w-4 h-4 text-violet-400" />
              <h3 className="font-medium text-primary uppercase text-[10px] tracking-wider">
                Mapeamento de Entrada do Script
              </h3>
            </div>

            {!selectedScript ? (
              <Card className="p-8 bg-black/20 border-border border-dashed text-center flex flex-col items-center justify-center gap-3">
                <Settings2 className="w-8 h-8 text-secondary" />
                <p className="text-xs text-secondary max-w-[200px]">
                  Selecione um script para configurar as entradas de execução.
                </p>
              </Card>
            ) : (
              <div className="grid grid-cols-2 gap-4 h-[400px]">
                {/* FONTE DE DADOS */}
                <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0">
                  <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                    <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">
                      Fonte de Dados
                    </span>
                    <span className="text-[9px] px-1.5 py-0.5 rounded bg-violet-500/10 text-violet-400 border border-violet-500/20">
                      Read Only
                    </span>
                  </div>
                  <div className="flex-1 overflow-auto p-2">
                    <JsonViewMain
                      data={sourceData}
                      draggableValue={true}
                      pProps={{
                        className:
                          "bg-transparent border-none text-xs text-primary",
                      }}
                    />
                  </div>
                </Card>

                {/* ENTRADAS DO SCRIPT */}
                <Card className="flex flex-col bg-foreground border-border overflow-hidden py-0">
                  <div className="p-2 border-b border-border bg-muted flex items-center justify-between">
                    <span className="text-[10px] font-bold text-secondary uppercase tracking-widest">
                      Argumentos do Script
                    </span>
                    <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Mappable
                    </span>
                  </div>
                  <div className="flex-1 overflow-auto p-2">
                    <Controller
                      name="scriptInputs"
                      control={control}
                      render={({ field }) => (
                        <JsonViewMain
                          data={field.value || {}}
                          onParamChange={(path, key, val) => {
                            const fullPath = path ? `${path}.${key}` : key;
                            setValue(
                              "scriptInputs",
                              updateNestedValue(field.value, fullPath, val)
                            );
                          }}
                          onAddField={(path, key, type) => {
                            const defaultVal =
                              type === "object"
                                ? {}
                                : type === "array"
                                ? []
                                : type === "number"
                                ? 0
                                : type === "boolean"
                                ? false
                                : "";
                            const fullPath = path ? `${path}.${key}` : key;
                            setValue(
                              "scriptInputs",
                              updateNestedValue(field.value, fullPath, defaultVal)
                            );
                          }}
                          onDeleteField={(path, key) => {
                            const fullPath = path ? `${path}.${key}` : key;
                            setValue(
                              "scriptInputs",
                              deleteNestedValue(field.value, fullPath)
                            );
                          }}
                          onDestructure={handleDestructure}
                          draggableValue={false}
                          pProps={{
                            className:
                              "bg-transparent border-none text-xs text-primary",
                          }}
                        />
                      )}
                    />
                  </div>
                </Card>
              </div>
            )}
          </div>
        </div>

        {selectedScript && (
          <div className="p-4 bg-violet-500/5 border border-violet-500/10 rounded-lg flex items-start gap-3">
            <Info className="w-4 h-4 text-violet-400 mt-0.5" />
            <div className="space-y-1">
              <p className="text-xs text-primary font-medium">Instruções de Mapeamento</p>
              <p className="text-[10px] text-secondary leading-relaxed">
                As entradas do script podem ser mapeadas dinamicamente usando a sintaxe <code className="text-violet-400">{"{{device.ip_address}}"}</code>.
                O sistema substituirá esses valores em tempo de execução com os dados reais do dispositivo monitorado.
              </p>
            </div>
          </div>
        )}
      </div>
    </ScrollArea>
  );
}

