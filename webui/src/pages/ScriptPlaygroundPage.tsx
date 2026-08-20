import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ChevronLeft, Play } from "lucide-react";
import ScriptPlayground from "@/components/ScriptPlayground";
import { Button } from "@/components/ui/button";
import { scriptsApi } from "@/rest-api-client/scripts.service";
import { useScriptsStore } from "@/stores/scripts-store";
import type { DownloadedScriptReturningValues, SchemaConfigIO } from "@/types/db";
import { inferSchema } from "@/utils/schema-inference";

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

function jsonSchemaToNamedOutput(name: "stdout" | "stderr", schema: any): SchemaConfigIO[] {
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

export default function ScriptPlaygroundPage() {
  const { id, sampleName } = useParams<{ id?: string; sampleName?: string }>();
  const navigate = useNavigate();
  const upsertScript = useScriptsStore((state) => state.upsertScript);
  const [script, setScript] = useState<DownloadedScriptReturningValues | null>(null);
  const [schema, setSchema] = useState<any>(null);
  const [inputs, setInputs] = useState<Record<string, any>>({});
  const [output, setOutput] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [loadingPage, setLoadingPage] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoadingPage(true);
      setError(null);
      setOutput(null);
      try {
        if (sampleName) {
          const res: any = await scriptsApi.getSampleSchema(sampleName);
          if (!cancelled) {
            setScript(null);
            setSchema(res.data || { inputs: [], outputs: [] });
          }
          return;
        }

        if (!id) {
          setError("Script nao informado.");
          return;
        }

        const res: any = await scriptsApi.getById(id);
        const nextScript = res.data as DownloadedScriptReturningValues;
        if (!cancelled) {
          setScript(nextScript);
          setSchema(nextScript.schema_config || { inputs: [], outputs: [] });
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err.response?.data?.error || err.response?.data?.message || err.message || "Erro ao carregar playground.");
        }
      } finally {
        if (!cancelled) {
          setLoadingPage(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [id, sampleName]);

  const handleRun = async () => {
    setLoading(true);
    setError(null);
    try {
      const res: any = await scriptsApi.runPlayground({
        script_id: script?.id,
        sample_name: sampleName,
        inputs,
      });
      setOutput(res.data);
    } catch (err: any) {
      setOutput({
        stdout: null,
        stderr: err.response?.data?.error || err.response?.data?.message || err.message,
      });
    } finally {
      setLoading(false);
    }
  };

  const persistOutputSchema = async (
    outputName: "stdout" | "stderr",
    outputData: any,
    successMessage: string,
  ) => {
    if (!script?.id) return;

    setLoading(true);
    try {
      const outputSchema = inferSchema(outputData);
      const nextOutputFields = jsonSchemaToNamedOutput(outputName, outputSchema);
      const currentOutputs = Array.isArray(script.schema_config?.outputs)
        ? [...script.schema_config.outputs]
        : [];
      const replacedNames = new Set(nextOutputFields.map((field) => field.name));
      const preservedOutputs = currentOutputs.filter(
        (field) => !replacedNames.has(field.name || ""),
      );
      const newConfig = {
        ...(script.schema_config || { inputs: [] }),
        outputs: [...nextOutputFields, ...preservedOutputs],
      };

      const updated = await scriptsApi.update(script.id, {
        schema_config: newConfig,
      });
      const updatedScript = updated?.data ?? { ...script, schema_config: newConfig };
      setScript(updatedScript);
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

  const stdoutCandidate = resolveStdoutCandidate(output);
  const stderrCandidate = resolveStderrCandidate(output);
  const title = script?.name || sampleName || "Playground";

  return (
    <main className="flex h-full min-h-screen flex-col bg-[#09090b] text-zinc-400">
      <div className="border-b border-white/[0.08] bg-white/[0.02] px-6 py-5">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div>
            <button
              type="button"
              onClick={() => navigate(script ? `/script-hub/${script.id}` : "/script-hub")}
              className="mb-4 inline-flex items-center gap-2 text-xs text-zinc-500 transition-colors hover:text-white"
            >
              <ChevronLeft size={14} />
              Voltar
            </button>
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-violet-500/20 bg-violet-500/10">
                <Play className="h-5 w-5 text-violet-300" />
              </div>
              <div>
                <h1 className="text-2xl font-semibold tracking-tight text-white">Playground: {title}</h1>
                <p className="text-sm text-zinc-500">
                  Teste entradas e visualize a saida em paineis redimensionaveis.
                </p>
              </div>
            </div>
          </div>
          <Button
            variant="outline"
            asChild
            className="border-white/[0.1] bg-transparent text-white hover:bg-white/[0.06]"
          >
            <Link to={script ? `/script-hub/${script.id}/files/edit` : "/script-hub/new"}>
              {script ? "Editar Arquivos" : "Criar Script"}
            </Link>
          </Button>
        </div>
      </div>

      <div className="mx-auto flex min-h-0 w-full max-w-7xl flex-1 flex-col px-6 py-6">
        {error && (
          <div className="mb-4 rounded-xl border border-red-400/20 bg-red-400/10 px-4 py-3 text-sm text-red-200">
            {error}
          </div>
        )}
        <div className="min-h-[680px] flex-1 overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.02]">
          {loadingPage ? (
            <div className="flex h-full items-center justify-center text-sm text-zinc-500">
              Carregando playground...
            </div>
          ) : (
            <ScriptPlayground
              handleRun={handleRun}
              setInputs={setInputs}
              inputs={inputs}
              output={output}
              stdoutCandidate={stdoutCandidate}
              stderrCandidate={stderrCandidate}
              schema={schema}
              loading={loading}
              handleSaveStdoutSchema={script ? (data) => persistOutputSchema("stdout", data, "Esquema de Stdout gravado com sucesso!") : undefined}
              handleSaveStderrSchema={script ? (data) => persistOutputSchema("stderr", data, "Esquema de Stderr gravado com sucesso!") : undefined}
            />
          )}
        </div>
      </div>
    </main>
  );
}
