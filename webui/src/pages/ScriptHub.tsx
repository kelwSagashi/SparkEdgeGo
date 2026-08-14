import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Package, Download, Search, Check, FolderGit2, Play, Plus, BookOpen, Clock, Activity, Code
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import { useScriptsStore } from '@/stores/scripts-store';
import { scriptsApi } from '@/rest-api-client/scripts.service';
import { AddScriptDialog } from '@/components/script-hub/add-script-dialog';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import type { DownloadedScriptReturningValues } from "@/types/db";

import ScriptPlayground from '@/components/ScriptPlayground';
import { inferSchema } from '@/utils/schema-inference';
import type { SchemaConfigIO } from "@/types/db";

function jsonSchemaToField(name: string, schema: any, required = false): SchemaConfigIO {
  const field: SchemaConfigIO = {
    name,
    type: schema?.type || "string",
    required,
  };

  if (schema?.type === "object" && schema?.properties) {
    field.fields = Object.entries(schema.properties).map(([childName, childSchema]: [string, any]) =>
      jsonSchemaToField(childName, childSchema, schema.required?.includes(childName)),
    );
  }
  return field;
}

function jsonSchemaToOutputFields(schema: any): SchemaConfigIO[] {
  if (schema?.type === "object" && schema?.properties) {
    return Object.entries(schema.properties).map(([name, childSchema]: [string, any]) =>
      jsonSchemaToField(name, childSchema, schema.required?.includes(name)),
    );
  }
  return [jsonSchemaToField("stdout", schema)];
}

function jsonSchemaToNamedOutput(name: string, schema: any): SchemaConfigIO[] {
  if (name === "stdout") {
    return jsonSchemaToOutputFields(schema);
  }
  return [jsonSchemaToField(name, schema)];
}

function resolveStdoutCandidate(output: any) {
  if (output === null || output === undefined) {
    return null;
  }
  if (
    typeof output === "object" &&
    !Array.isArray(output) &&
    Object.prototype.hasOwnProperty.call(output, "stdout") &&
    Object.prototype.hasOwnProperty.call(output, "stderr") &&
    Object.keys(output).length === 2
  ) {
    return output.stdout ?? null;
  }
  return output;
}

function resolveStderrCandidate(output: any) {
  if (
    output &&
    typeof output === "object" &&
    !Array.isArray(output) &&
    Object.prototype.hasOwnProperty.call(output, "stdout") &&
    Object.prototype.hasOwnProperty.call(output, "stderr") &&
    Object.keys(output).length === 2
  ) {
    return output.stderr ?? null;
  }
  return null;
}

// Playground Dialog Component
function ScriptPlaygroundDialog({ open, onOpenChange, script, sampleName }: { open: boolean, onOpenChange: (open: boolean) => void, script?: DownloadedScriptReturningValues, sampleName?: string }) {
  const upsertScript = useScriptsStore((state) => state.upsertScript);
  const [currentScript, setCurrentScript] = useState<DownloadedScriptReturningValues | undefined>(script);
  const [schema, setSchema] = useState<any>(null);
  const [inputs, setInputs] = useState<Record<string, any>>({});
  const [output, setOutput] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setCurrentScript(script);
  }, [script]);

  useEffect(() => {
    if (!open) {
      setCurrentScript(script);
      setSchema(null);
      setInputs({});
      setOutput(null);
      return;
    }

    const fetchSchema = async () => {
      if (sampleName) {
        setLoading(true);
        const res: any = await scriptsApi.getSampleSchema(sampleName);
        setSchema(res.data);
        setLoading(false);
      } else if (currentScript) {
        // Assume schema is saved in DB under script.schema_config
        setSchema((currentScript as any).schema_config || { inputs: [], outputs: [] });
      }
    };
    fetchSchema();
  }, [open, currentScript, sampleName, script]);

  const handleRun = async () => {
    setLoading(true);
    try {
      const res: any = await scriptsApi.runPlayground({ script_id: script?.id, sample_name: sampleName, inputs });
      setOutput(res.data);
    } catch(err: any) {
      setOutput({ stdout: null, stderr: err.message });
    } finally {
        setLoading(false);
    }
  };

  const persistOutputSchema = async (
    outputName: "stdout" | "stderr",
    outputData: any,
    successMessage: string,
  ) => {
    if (!currentScript?.id) return;

    setLoading(true);
    try {
      const outputSchema = inferSchema(outputData);
      const nextOutputFields = jsonSchemaToNamedOutput(outputName, outputSchema);

      const currentOutputs = Array.isArray(currentScript.schema_config?.outputs)
        ? [...currentScript.schema_config.outputs]
        : [];

      const replacedNames = new Set(nextOutputFields.map((field) => field.name));
      const preservedOutputs = currentOutputs.filter(
        (field) => !replacedNames.has(field.name || ""),
      );

      const newConfig = {
        ...(currentScript.schema_config || { inputs: [] }),
        outputs: [...nextOutputFields, ...preservedOutputs],
      };

      const updated = await scriptsApi.update(currentScript.id, {
        schema_config: newConfig,
      });
      const updatedScript = updated?.data ?? { ...currentScript, schema_config: newConfig };
      setCurrentScript(updatedScript);
      setSchema(newConfig);
      upsertScript(updatedScript);
      window.dispatchEvent(
        new CustomEvent("sparkedge-script-schema-updated", {
          detail: { script: updatedScript },
        }),
      );

      alert(successMessage);
    } catch (err: any) {
      alert("Erro ao gravar esquema: " + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveStdoutSchema = async (stdoutData: any) => {
    await persistOutputSchema("stdout", stdoutData, "Esquema de Stdout gravado com sucesso!");
  };

  const handleSaveStderrSchema = async (stderrData: any) => {
    await persistOutputSchema("stderr", stderrData, "Esquema de Stderr gravado com sucesso!");
  };

  const stdoutCandidate = resolveStdoutCandidate(output);
  const stderrCandidate = resolveStderrCandidate(output);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-[800px] h-[600px] flex flex-col bg-[#09090b] border-white/[0.08] p-0 overflow-hidden">
            <DialogHeader className="p-4 py-3 border-b border-white/[0.08] bg-white/[0.02]">
               <DialogTitle className="text-white flex items-center gap-2">
                 <Play className="w-4 h-4 text-violet-400" />
                 Playground: {currentScript?.name || sampleName}
               </DialogTitle>
            </DialogHeader>
            <ScriptPlayground
              handleRun={handleRun}
              setInputs={setInputs}
              inputs={inputs}
              output={output}
              schema={schema}
              loading={loading}
              stdoutCandidate={stdoutCandidate}
              stderrCandidate={stderrCandidate}
              handleSaveStdoutSchema={currentScript ? handleSaveStdoutSchema : undefined}
              handleSaveStderrSchema={currentScript ? handleSaveStderrSchema : undefined}
            />
        </DialogContent>
    </Dialog>
  );
}


function ScriptCard({ id, title, subtitle, info, badges, onPlay, onDelete }: { id?: string, title: string, subtitle: string, info: string, badges: string[], onPlay: () => void, onDelete?: () => void }) {
  const navigate = useNavigate();
  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      onClick={() => id && navigate(`/script-hub/${id}`)}
      className="bg-white/[0.03] hover:bg-white/[0.06] border border-white/[0.08] hover:border-white/[0.15] rounded-xl p-5 transition-all duration-300 flex flex-col cursor-pointer group"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500/20 to-blue-500/20 border border-white/[0.08] flex items-center justify-center group-hover:scale-105 transition-transform">
            <Code size={18} className="text-violet-400" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white group-hover:text-violet-400 transition-colors">{title}</h3>
            <p className="text-[11px] text-zinc-500">{subtitle}</p>
          </div>
        </div>
      </div>
      <p className="text-xs text-zinc-500 mb-4 line-clamp-2 flex-1">{info}</p>
      {badges.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-4">
          {badges.map(tag => (
            <span key={tag} className="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-medium bg-white/[0.06] text-zinc-400 border border-white/[0.06]">
              {tag}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
        <Button 
          size="sm" 
          variant="outline" 
          onClick={(e) => { e.stopPropagation(); onPlay(); }} 
          className="gap-1.5 text-xs h-8 border-violet-500/20 text-violet-400 hover:bg-violet-500/10"
        >
          <Play size={12} /> Playground
        </Button>
        {onDelete && (
             <Button 
              size="sm" 
              variant="ghost" 
              onClick={(e) => { e.stopPropagation(); onDelete(); }} 
              className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs"
             >
               Excluir
             </Button>
        )}
      </div>
    </motion.div>
  );
}

export default function ScriptHubPage() {
  const { scripts, samples, loading, fetchAll, fetchSamples, deleteScript } = useScriptsStore();
  const [search, setSearch] = useState('');
  const [uploadOpen, setUploadOpen] = useState(false);
  const [activePlayground, setActivePlayground] = useState<{ script?: DownloadedScriptReturningValues, sampleName?: string } | null>(null);

  useEffect(() => {
    fetchAll();
    fetchSamples();
  }, [fetchAll, fetchSamples]);

  const filteredScripts = scripts.filter(s => !search || s.name.toLowerCase().includes(search.toLowerCase()) || s.tags?.some((t: string) => t.toLowerCase().includes(search.toLowerCase())));
  const filteredSamples = samples.filter(s => !search || s.toLowerCase().includes(search.toLowerCase()));

  return (
    <main className="grow px-8 py-6 w-full max-w-[1200px] mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500/20 to-blue-500/20 border border-white/[0.08] flex items-center justify-center">
              <Package size={20} className="text-violet-400" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-white tracking-tight">Script Hub</h1>
              <p className="text-sm text-zinc-500">Desenvolva, teste e instale seus scripts.</p>
            </div>
          </div>
        </div>
        <Button onClick={() => setUploadOpen(true)} className="bg-violet-600 hover:bg-violet-700 text-white gap-2">
          Novo Script
        </Button>
      </div>

      {/* Search */}
      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={14} />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar scripts..."
            className="w-full pl-9 pr-3 py-2.5 bg-white/[0.04] border border-white/[0.1] rounded-lg text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] transition-colors"
          />
        </div>
      </div>

      <Tabs defaultValue="my-scripts" className="w-full">
         <TabsList className="mb-6 inline-flex h-9 items-center justify-center p-1 text-zinc-400">
             <TabsTrigger value="my-scripts" className="data-[state=active]:text-white px-3 py-1 text-sm transition-all focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50">
                 Instalados ({scripts.length})
             </TabsTrigger>
             <TabsTrigger value="samples" className="data-[state=active]:text-white px-3 py-1 text-sm transition-all focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50">
                 Exemplos Nativos ({samples.length})
             </TabsTrigger>
         </TabsList>

         <TabsContent value="my-scripts">
             {loading && scripts.length === 0 ? (
                <div className="flex items-center justify-center py-20"><div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" /></div>
             ) : filteredScripts.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 text-center bg-white/[0.02] border border-white/[0.08] rounded-xl border-dashed">
                  <h3 className="text-sm font-medium text-white mb-1">Nenhum script instalado</h3>
                  <p className="text-xs text-zinc-500 max-w-sm mb-4">Faça o upload do seu primeiro script Node ou Python compactado em ZIP.</p>
                  <Button variant="outline" onClick={() => setUploadOpen(true)} className="border-white/[0.1] bg-white/[0.05] hover:bg-white/[0.1] text-white">Upload</Button>
                </div>
             ) : (
                <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    <AnimatePresence mode="popLayout">
                        {filteredScripts.map(script => (
                            <ScriptCard 
                                key={script.id} 
                                id={script.id}
                                title={script.name} 
                                subtitle={`by ${script.author} v${script.version}`} 
                                info={script.description || 'Sem descrição'} 
                                badges={script.tags || []} 
                                onPlay={() => setActivePlayground({ script: script as any })}
                                onDelete={() => deleteScript(script.id)}
                            />
                        ))}
                    </AnimatePresence>
                </motion.div>
             )}
         </TabsContent>
         <TabsContent value="samples">
             {loading && samples.length === 0 ? (
                <div className="flex items-center justify-center py-20"><div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" /></div>
             ) : filteredSamples.length === 0 ? (
                <div className="flex w-full h-32 items-center justify-center text-sm text-zinc-500">Nenhum exemplo.</div>
             ) : (
                <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    <AnimatePresence mode="popLayout">
                        {filteredSamples.map(sample => (
                            <ScriptCard 
                                key={sample} 
                                title={sample} 
                                subtitle="Biblioteca Nativa" 
                                info="Este é um script de exemplo empacotado nativamente." 
                                badges={['python']} 
                                onPlay={() => setActivePlayground({ sampleName: sample })}
                            />
                        ))}
                    </AnimatePresence>
                </motion.div>
             )}
         </TabsContent>
      </Tabs>

      <AddScriptDialog 
         open={uploadOpen} 
         onOpenChange={setUploadOpen} 
         onSuccess={() => fetchAll()}
      />

      <ScriptPlaygroundDialog 
         open={!!activePlayground}
         onOpenChange={(v) => !v && setActivePlayground(null)}
         script={activePlayground?.script}
         sampleName={activePlayground?.sampleName}
      />
    </main>
  );
}

