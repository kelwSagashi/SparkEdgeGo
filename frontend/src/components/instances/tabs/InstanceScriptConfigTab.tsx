/**
 * Instance Script Configuration Tab
 * Select script and configure parameters
 */

import { useCallback, useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { AlertCircle, HelpCircle } from "lucide-react";

interface InstanceScriptConfigTabProps {
  data: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceScriptConfigTab({
  data,
  onChange,
  errors,
}: InstanceScriptConfigTabProps) {
  const [scripts, setScripts] = useState<any[]>([]);
  const [selectedScriptDetails, setSelectedScriptDetails] = useState<any>(null);
  const [isLoadingScripts, setIsLoadingScripts] = useState(true);

  useEffect(() => {
    // Fetch available scripts
    const fetchScripts = async () => {
      try {
        const response = await fetch("/api/scripts/downloaded");
        if (response.ok) {
          const result = await response.json();
          setScripts(result.data || []);

          // Load selected script details
          if (data.scriptId) {
            const selected = result.data?.find(
              (s: any) => s.id === data.scriptId,
            );
            setSelectedScriptDetails(selected);
          }
        }
      } catch (error) {
        console.error("Failed to load scripts:", error);
      } finally {
        setIsLoadingScripts(false);
      }
    };

    fetchScripts();
  }, [data.scriptId]);

  const handleScriptSelect = useCallback(
    (scriptId: string) => {
      const selected = scripts.find((s) => s.id === scriptId);
      onChange({
        ...data,
        scriptId,
        scriptParameters:
          selected?.schema_config?.inputs?.map((input: any) => ({
            key: input.key,
            value: input.default || null,
            sourceType: "manual",
          })) || [],
      });
      setSelectedScriptDetails(selected);
    },
    [scripts, data, onChange],
  );

  const handleParameterChange = useCallback(
    (paramKey: string, value: any, sourceType: string) => {
      onChange({
        ...data,
        scriptParameters: data.scriptParameters.map((p: any) =>
          p.key === paramKey ? { ...p, value, sourceType } : p,
        ),
      });
    },
    [data, onChange],
  );

  return (
    <div className="space-y-6">
      <Card className="bg-zinc-900/50 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-lg">Configuração do Script</CardTitle>
          <CardDescription>
            Selecione o script Python a ser executado e configure seus
            parâmetros
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Script Selection */}
          <div className="space-y-3">
            <Label className="text-sm font-medium">
              Script <span className="text-red-500">*</span>
            </Label>

            {isLoadingScripts ? (
              <div className="text-sm text-zinc-400 p-4 text-center">
                Carregando scripts...
              </div>
            ) : scripts.length === 0 ? (
              <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4 flex gap-3">
                <AlertCircle className="text-yellow-400 shrink-0" size={20} />
                <div>
                  <h4 className="font-medium text-yellow-300 text-sm">
                    Nenhum script encontrado
                  </h4>
                  <p className="text-xs text-yellow-200 mt-1">
                    Você precisa fazer download de um script primeiro. Vá para
                    Script Hub.
                  </p>
                </div>
              </div>
            ) : (
              <div className="space-y-2">
                {scripts.map((script) => (
                  <button
                    key={script.id}
                    onClick={() => handleScriptSelect(script.id)}
                    className={`w-full text-left p-3 rounded-lg border transition-colors ${
                      data.scriptId === script.id
                        ? "bg-blue-500/20 border-blue-500/50"
                        : "bg-zinc-800/50 border-zinc-700 hover:border-zinc-600"
                    }`}
                  >
                    <div className="font-medium text-sm">{script.name}</div>
                    <div className="text-xs text-zinc-400 mt-1">
                      {script.description}
                    </div>
                    <div className="text-xs text-zinc-500 mt-2">
                      {script.author} • v{script.version}
                    </div>
                  </button>
                ))}
              </div>
            )}

            {errors.scriptId && (
              <p className="text-xs text-red-400">{errors.scriptId}</p>
            )}
          </div>

          {/* Script Details and Parameters */}
          {selectedScriptDetails && (
            <div className="space-y-4 pt-4 border-t border-zinc-700">
              <div>
                <h4 className="text-sm font-medium mb-2">Detalhes do Script</h4>
                <div className="bg-zinc-800/50 rounded-lg p-3 space-y-2 text-sm">
                  <div>
                    <span className="text-zinc-400">Arquivo Principal:</span>{" "}
                    <span className="text-zinc-200">
                      {selectedScriptDetails.main_file}
                    </span>
                  </div>
                  <div>
                    <span className="text-zinc-400">Linguagem:</span>{" "}
                    <span className="text-zinc-200">
                      {selectedScriptDetails.language}
                    </span>
                  </div>
                  {selectedScriptDetails.requirements_file && (
                    <div>
                      <span className="text-zinc-400">Dependências:</span>{" "}
                      <span className="text-zinc-200">
                        {selectedScriptDetails.requirements_file}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* Script Parameters */}
              {selectedScriptDetails.schema_config?.inputs &&
                selectedScriptDetails.schema_config.inputs.length > 0 && (
                  <div className="space-y-3">
                    <h4 className="text-sm font-medium flex items-center gap-2">
                      Parâmetros de Entrada
                      <HelpCircle size={14} className="text-zinc-400" />
                    </h4>

                    <div className="space-y-3 bg-zinc-800/30 rounded-lg p-4">
                      {selectedScriptDetails.schema_config.inputs.map(
                        (input: any) => {
                          const param = data.scriptParameters.find(
                            (p: any) => p.key === input.key,
                          );
                          return (
                            <div key={input.key} className="space-y-2">
                              <div className="flex items-center justify-between">
                                <Label className="text-sm">
                                  {input.label || input.key}
                                  {input.required && (
                                    <span className="text-red-500 ml-1">*</span>
                                  )}
                                </Label>
                              </div>

                              {input.description && (
                                <p className="text-xs text-zinc-400">
                                  {input.description}
                                </p>
                              )}

                              {/* Parameter input would go here - placeholder for now */}
                              <div className="text-xs text-zinc-500 p-2 bg-zinc-900/50 rounded">
                                Controle de entrada será implementado baseado no
                                tipo do parâmetro
                              </div>
                            </div>
                          );
                        },
                      )}
                    </div>
                  </div>
                )}
            </div>
          )}

          {/* Info Box */}
          <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 space-y-2">
            <h4 className="text-sm font-medium text-blue-300">
              💡 Sobre Scripts
            </h4>
            <p className="text-sm text-blue-200">
              Scripts são programas Python que coletam dados do seu dispositivo.
              Cada script possui parâmetros que você pode configurar para
              adaptá-lo à sua necessidade.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

