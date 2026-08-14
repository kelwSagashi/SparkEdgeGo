import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  AlertTriangle,
  BookOpenText,
  CheckCircle2,
  ChevronLeft,
  Code2,
  Clock,
  Download,
  Edit,
  FileText,
  Globe,
  Package,
  Play,
  RotateCcw,
  Tag,
  TerminalSquare,
  Upload,
  User,
} from "lucide-react";
import ReactMarkdown from "react-markdown";
import ScriptPlayground from "@/components/ScriptPlayground";
import { AddScriptDialog } from "@/components/script-hub/add-script-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { scriptsApi } from "@/rest-api-client/scripts.service";
import type { DownloadedScriptHistoryEntry, DownloadedScriptReturningValues } from "@/types/db";

function markdownComponents() {
  return {
    h1: ({ children }: any) => (
      <h1 className="text-3xl font-bold text-white tracking-tight mt-0 mb-6 border-b border-white/[0.08] pb-4">
        {children}
      </h1>
    ),
    h2: ({ children }: any) => (
      <h2 className="text-2xl font-semibold text-white mt-10 mb-4">{children}</h2>
    ),
    h3: ({ children }: any) => (
      <h3 className="text-xl font-semibold text-zinc-100 mt-8 mb-3">{children}</h3>
    ),
    p: ({ children }: any) => (
      <p className="text-[15px] leading-8 text-zinc-300 mb-4">{children}</p>
    ),
    ul: ({ children }: any) => (
      <ul className="list-disc pl-6 space-y-2 text-zinc-300 mb-5">{children}</ul>
    ),
    ol: ({ children }: any) => (
      <ol className="list-decimal pl-6 space-y-2 text-zinc-300 mb-5">{children}</ol>
    ),
    li: ({ children }: any) => <li className="leading-7">{children}</li>,
    blockquote: ({ children }: any) => (
      <blockquote className="border-l-4 border-violet-500/50 bg-violet-500/[0.08] px-4 py-3 rounded-r-xl text-zinc-200 mb-5">
        {children}
      </blockquote>
    ),
    a: ({ href, children }: any) => (
      <a
        href={href}
        target="_blank"
        rel="noreferrer"
        className="text-sky-300 hover:text-sky-200 underline underline-offset-4"
      >
        {children}
      </a>
    ),
    code: ({ inline, children }: any) =>
      inline ? (
        <code className="px-1.5 py-0.5 rounded-md bg-white/[0.06] text-violet-300 text-[0.95em]">
          {children}
        </code>
      ) : (
        <code className="text-zinc-100">{children}</code>
      ),
    pre: ({ children }: any) => (
      <pre className="bg-black/50 border border-white/[0.08] rounded-2xl p-5 overflow-x-auto mb-6 text-sm leading-7">
        {children}
      </pre>
    ),
    hr: () => <hr className="border-white/[0.08] my-8" />,
  };
}

export default function ScriptDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [script, setScript] = useState<DownloadedScriptReturningValues | null>(null);
  const [readme, setReadme] = useState<string | null>(null);
  const [mainCode, setMainCode] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [playgroundOpen, setPlaygroundOpen] = useState(false);
  const [playgroundInputs, setPlaygroundInputs] = useState<Record<string, any>>({});
  const [playgroundOutput, setPlaygroundOutput] = useState<any>(null);
  const [playgroundLoading, setPlaygroundLoading] = useState(false);
  const [updateBundleOpen, setUpdateBundleOpen] = useState(false);
  const [history, setHistory] = useState<DownloadedScriptHistoryEntry[]>([]);
  const [restoringHistoryId, setRestoringHistoryId] = useState<string | null>(null);

  const fetchScriptData = async () => {
    if (!id) {
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const scriptRes: any = await scriptsApi.getById(id);
      if (!scriptRes.data) {
        setError("Script não encontrado");
        return;
      }

      const nextScript = scriptRes.data as DownloadedScriptReturningValues;
      setScript(nextScript);

      const reads: Promise<void>[] = [
        scriptsApi
          .getHistory(id)
          .then((historyRes: any) => {
            setHistory((historyRes.data || []) as DownloadedScriptHistoryEntry[]);
          })
          .catch(() => {
            setHistory([]);
          }),
        scriptsApi
          .getFileContent(id, "README.md")
          .then((readmeRes: any) => {
            setReadme(readmeRes.data || null);
          })
          .catch(() => {
            setReadme(null);
          }),
      ];

      if (nextScript.main_file) {
        reads.push(
          scriptsApi
            .getFileContent(id, nextScript.main_file)
            .then((codeRes: any) => {
              setMainCode(codeRes.data || null);
            })
            .catch(() => {
              setMainCode(null);
            }),
        );
      } else {
        setMainCode(null);
      }

      await Promise.all(reads);
    } catch (err: any) {
      setError(err.message || "Erro ao carregar script");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchScriptData();
  }, [id]);

  const handleRunPlayground = async () => {
    if (!script?.id) {
      return;
    }

    setPlaygroundLoading(true);
    try {
      const response: any = await scriptsApi.runPlayground({
        script_id: script.id,
        inputs: playgroundInputs,
      });
      setPlaygroundOutput(response.data);
    } catch (err: any) {
      setPlaygroundOutput({
        stdout: null,
        stderr: err?.response?.data?.error || err?.message || "Erro ao executar script",
      });
    } finally {
      setPlaygroundLoading(false);
    }
  };

  const handleRestoreHistory = async (entry: DownloadedScriptHistoryEntry) => {
    if (!script?.id || !entry.id || !entry.can_restore) {
      return;
    }

    const confirmed = window.confirm(
      `Restaurar o script para a versão registrada em ${entry.created_at ? new Date(entry.created_at).toLocaleString("pt-BR") : "um ponto anterior do histórico"}?`,
    );
    if (!confirmed) {
      return;
    }

    setRestoringHistoryId(entry.id);
    setError(null);
    try {
      await scriptsApi.restoreHistory(script.id, entry.id);
      await fetchScriptData();
    } catch (err: any) {
      setError(err?.response?.data?.error || err?.message || "Erro ao restaurar versão do script");
    } finally {
      setRestoringHistoryId(null);
    }
  };

  const schema = useMemo(
    () => script?.schema_config || { inputs: [], outputs: [] },
    [script],
  );

  const actionLabel = (action: string) => {
    switch (action) {
      case "installed":
        return "Instalação inicial";
      case "bundle_updated":
        return "Atualização de bundle";
      case "metadata_updated":
        return "Edição de metadados";
      case "restored":
        return "Restauração";
      default:
        return action;
    }
  };

  if (loading) {
    return (
      <div className="p-8 space-y-6 max-w-6xl mx-auto">
        <Skeleton className="h-10 w-48 mb-6" />
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          <div className="lg:col-span-3 space-y-4">
            <Skeleton className="h-12 w-3/4" />
            <Skeleton className="h-[400px] w-full" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
          </div>
        </div>
      </div>
    );
  }

  if (error || !script) {
    return (
      <div className="p-20 text-center space-y-4">
        <AlertTriangle className="w-12 h-12 text-amber-500 mx-auto" />
        <h2 className="text-xl font-semibold text-white">{error || "Algo deu errado"}</h2>
        <Button onClick={() => navigate("/script-hub")}>Voltar ao Hub</Button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-400 pb-20">
      <div className="border-b border-white/[0.06] bg-white/[0.02]">
        <div className="max-w-6xl mx-auto px-6 py-6">
          <Link
            to="/script-hub"
            className="inline-flex items-center gap-2 text-xs text-zinc-500 hover:text-white transition-colors mb-6 group"
          >
            <ChevronLeft size={14} className="group-hover:-translate-x-0.5 transition-transform" />
            Voltar para Script Hub
          </Link>

          <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
            <div className="flex items-start gap-5">
              <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-600 to-blue-600 p-[1px]">
                <div className="w-full h-full rounded-2xl bg-[#09090b] flex items-center justify-center">
                  <Package size={32} className="text-white" />
                </div>
              </div>
              <div className="space-y-1">
                <h1 className="text-3xl font-bold text-white tracking-tight">{script.name}</h1>
                <div className="flex items-center gap-4 text-sm">
                  <span className="flex items-center gap-1.5">
                    <User size={14} /> {script.author}
                  </span>
                  <span className="text-zinc-600">•</span>
                  <span className="flex items-center gap-1.5">
                    <Clock size={14} /> v{script.version}
                  </span>
                  {script.source === "local" && (
                    <Badge
                      variant="outline"
                      className="bg-green-500/10 text-green-400 border-green-500/20 gap-1 capitalize"
                    >
                      <CheckCircle2 size={10} /> {script.source}
                    </Badge>
                  )}
                </div>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <Button
                variant="outline"
                onClick={() => setUpdateBundleOpen(true)}
                className="border-white/[0.1] bg-white/[0.02] hover:bg-white/[0.06] text-white"
              >
                <Upload className="w-4 h-4 mr-2" /> Atualizar Bundle
              </Button>
              <Button
                variant="outline"
                asChild
                className="border-white/[0.1] bg-white/[0.02] hover:bg-white/[0.06] text-white"
              >
                <Link to={`/script-hub/${script.id}/edit`}>
                  <Edit className="w-4 h-4 mr-2" /> Editar
                </Link>
              </Button>
              <Button
                onClick={() => setPlaygroundOpen(true)}
                className="bg-violet-600 hover:bg-violet-700 text-white shadow-lg shadow-violet-600/20"
              >
                <Play className="w-4 h-4 mr-2" /> Playground
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-6xl mx-auto px-6 pt-10">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-10">
          <div className="lg:col-span-3 space-y-8">
            <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4 md:p-8 min-h-[560px]">
              <Tabs defaultValue="readme" className="w-full">
                <TabsList className="mb-8 bg-white/[0.03] border border-white/[0.06]">
                  <TabsTrigger value="readme" className="gap-2">
                    <BookOpenText size={14} />
                    Documentação
                  </TabsTrigger>
                  <TabsTrigger value="code" className="gap-2">
                    <TerminalSquare size={14} />
                    Código
                  </TabsTrigger>
                </TabsList>

                <TabsContent value="readme">
                  {readme ? (
                    <article className="max-w-none">
                      <ReactMarkdown components={markdownComponents()}>{readme}</ReactMarkdown>
                    </article>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-20 text-center space-y-4">
                      <FileText className="w-12 h-12 text-zinc-700" />
                      <div>
                        <h3 className="text-lg font-medium text-zinc-300">Sem documentação</h3>
                        <p className="text-sm text-zinc-500">
                          Este script não possui um arquivo README.md identificável.
                        </p>
                      </div>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="code">
                  {mainCode ? (
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-xs uppercase tracking-[0.2em] text-zinc-500 mb-1">
                            Entrypoint
                          </p>
                          <p className="text-sm font-mono text-zinc-200">{script.main_file}</p>
                        </div>
                      </div>
                      <pre className="bg-black/50 border border-white/[0.08] rounded-2xl p-5 overflow-x-auto text-sm leading-7 text-zinc-100">
                        <code>{mainCode}</code>
                      </pre>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-20 text-center space-y-4">
                      <Code2 className="w-12 h-12 text-zinc-700" />
                      <div>
                        <h3 className="text-lg font-medium text-zinc-300">Código indisponível</h3>
                        <p className="text-sm text-zinc-500">
                          Não foi possível carregar o arquivo principal do script.
                        </p>
                      </div>
                    </div>
                  )}
                </TabsContent>
              </Tabs>
            </div>
          </div>

          <div className="space-y-6">
            <Card className="bg-white/[0.02] border border-white/[0.06] overflow-hidden">
              <CardContent className="p-6 space-y-6">
                <div>
                  <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-4">
                    Informações
                  </h4>
                  <div className="space-y-4">
                    <div className="flex items-center justify-between text-sm gap-4">
                      <span className="flex items-center gap-2">
                        <Globe size={14} /> Source
                      </span>
                      <span className="text-white capitalize">{script.source}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm gap-4">
                      <span className="flex items-center gap-2">
                        <Code2 size={14} /> Entrypoint
                      </span>
                      <span className="text-white font-mono text-[10px] text-right break-all">
                        {script.main_file}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm gap-4">
                      <span className="flex items-center gap-2">
                        <Tag size={14} /> Tags
                      </span>
                      <div className="flex flex-wrap justify-end gap-1">
                        {script.tags?.map((tag) => (
                          <Badge
                            key={tag}
                            variant="secondary"
                            className="text-[10px] py-0 px-1.5 bg-white/[0.05] text-zinc-400 border-none"
                          >
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>

                <Separator className="bg-white/[0.06]" />

                <div>
                  <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">
                    Descrição curta
                  </h4>
                  <p className="text-sm text-zinc-400 leading-relaxed">
                    {script.description || "Nenhuma descrição fornecida."}
                  </p>
                </div>

                <div className="pt-4">
                  <Button
                    variant="ghost"
                    className="w-full text-zinc-500 hover:text-white hover:bg-white/[0.05] text-xs"
                  >
                    <Download className="w-3 h-3 mr-2" /> Download Bundle
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card className="bg-gradient-to-br from-violet-600/10 to-blue-600/10 border border-violet-500/20">
              <CardContent className="p-6">
                <h4 className="text-sm font-semibold text-white mb-2">Instalação via CLI</h4>
                <div className="relative group">
                  <pre className="bg-black/40 p-3 rounded-lg text-[10px] font-mono text-zinc-300 overflow-x-auto border border-white/5">
                    spark-edge install {script.id}
                  </pre>
                </div>
              </CardContent>
            </Card>

            <Card className="bg-white/[0.02] border border-white/[0.06] overflow-hidden">
              <CardContent className="p-6 space-y-4">
                <div>
                  <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-1">
                    Histórico
                  </h4>
                  <p className="text-xs text-zinc-500">
                    Linha do tempo com resumo das mudanças e restauração de versões anteriores.
                  </p>
                </div>

                {history.length === 0 ? (
                  <p className="text-sm text-zinc-500">Nenhum evento registrado ainda.</p>
                ) : (
                  <div className="space-y-3">
                    {history.map((entry) => (
                      <div
                        key={entry.id}
                        className="rounded-xl border border-white/[0.06] bg-black/20 p-4 space-y-3"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div>
                            <p className="text-sm font-medium text-white">{actionLabel(entry.action)}</p>
                            <p className="text-xs text-zinc-500">
                              {entry.created_at
                                ? new Date(entry.created_at).toLocaleString("pt-BR")
                                : "Sem data"}
                            </p>
                          </div>
                          <Badge
                            variant="outline"
                            className="border-violet-500/20 bg-violet-500/10 text-violet-200"
                          >
                            v{entry.version || "?"}
                          </Badge>
                        </div>

                        <div className="space-y-1 text-xs text-zinc-400">
                          <p><span className="text-zinc-500">Nome:</span> {entry.name}</p>
                          <p><span className="text-zinc-500">Autor:</span> {entry.author || "-"}</p>
                          <p><span className="text-zinc-500">Entrypoint:</span> <span className="font-mono">{entry.main_file || "-"}</span></p>
                        </div>

                        {entry.change_summary && entry.change_summary.length > 0 && (
                          <div className="space-y-2 pt-1">
                            <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500">
                              Resumo das mudanças
                            </p>
                            <ul className="space-y-1">
                              {entry.change_summary.map((item, index) => (
                                <li key={`${entry.id}-${index}`} className="text-xs text-zinc-300">
                                  {item}
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}

                        <div className="pt-1">
                          <Button
                            type="button"
                            variant="outline"
                            disabled={!entry.can_restore || restoringHistoryId === entry.id}
                            onClick={() => void handleRestoreHistory(entry)}
                            className="border-white/[0.1] bg-white/[0.02] hover:bg-white/[0.06] text-white text-xs"
                          >
                            <RotateCcw className="w-3 h-3 mr-2" />
                            {restoringHistoryId === entry.id ? "Restaurando..." : "Restaurar esta versão"}
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <Dialog open={playgroundOpen} onOpenChange={setPlaygroundOpen}>
        <DialogContent className="sm:max-w-[800px] h-[600px] flex flex-col bg-[#09090b] border-white/[0.08] p-0 overflow-hidden">
          <DialogHeader className="p-4 py-3 border-b border-white/[0.08] bg-white/[0.02]">
            <DialogTitle className="text-white flex items-center gap-2">
              <Play className="w-4 h-4 text-violet-400" />
              Playground: {script.name}
            </DialogTitle>
          </DialogHeader>
          <ScriptPlayground
            handleRun={handleRunPlayground}
            setInputs={setPlaygroundInputs}
            inputs={playgroundInputs}
            output={playgroundOutput}
            stdoutCandidate={null}
            stderrCandidate={null}
            schema={schema}
            loading={playgroundLoading}
          />
        </DialogContent>
      </Dialog>

      <AddScriptDialog
        open={updateBundleOpen}
        onOpenChange={setUpdateBundleOpen}
        existingScript={script}
        onSuccess={() => {
          setUpdateBundleOpen(false);
          void fetchScriptData();
        }}
      />
    </div>
  );
}
