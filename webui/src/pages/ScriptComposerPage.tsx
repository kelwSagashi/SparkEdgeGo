import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { BookOpen, CheckCircle2, ChevronLeft, ChevronRight, Code2, FileText, Play, Plus, Trash2, UploadCloud, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { scriptsApi, type ScriptDraftFile, type ScriptDraftPayload } from '@/rest-api-client/scripts.service';
import type { DownloadedScriptReturningValues } from '@/types/db';

type InstallMode = 'zip' | 'editor';

const defaultDraftCode = `#!/usr/bin/env python3
from datetime import datetime, timezone

from Sparkit import Input, MainOut, Node, Run, sparkit


@Input(name="asset_id", required=False, type=str)
@Node
class InlineSparkEdgeScript:
    asset_id: str = "edge-device"

    @Run
    @MainOut
    def get_main_output(self) -> dict:
        return {
            "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "asset_id": self.asset_id,
            "message": "Hello from SparkEdge",
        }


if __name__ == "__main__":
    sparkit.run(InlineSparkEdgeScript)
`;

const defaultDraftFiles: ScriptDraftFile[] = [
  { path: 'main.py', content: defaultDraftCode },
  { path: 'requirements.txt', content: 'sparkit\n' },
  {
    path: 'README.md',
    content: '# Novo Script SparkEdge\n\nDescreva aqui o objetivo, entradas, saidas e como testar este script.\n',
  },
];

const cloneDefaultDraftFiles = () => defaultDraftFiles.map((file) => ({ ...file }));

export default function ScriptComposerPage() {
  const { id } = useParams<{ id?: string }>();
  const navigate = useNavigate();
  const [existingScript, setExistingScript] = useState<DownloadedScriptReturningValues | null>(null);
  const [mode, setMode] = useState<InstallMode>(id ? 'editor' : 'zip');
  const [step, setStep] = useState(1);
  const [file, setFile] = useState<File | null>(null);
  const [inspectData, setInspectData] = useState<{ tempFolder: string; pyFiles: string[] } | null>(null);
  const [uploading, setUploading] = useState(false);
  const [installError, setInstallError] = useState<string | null>(null);
  const [progress, setProgress] = useState<
    { step: string; status: 'pending' | 'loading' | 'success' | 'error' }[]
  >([]);
  const [mainFile, setMainFile] = useState('main.py');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [author, setAuthor] = useState('');
  const [version, setVersion] = useState('1.0.0');
  const [draftFiles, setDraftFiles] = useState<ScriptDraftFile[]>(cloneDefaultDraftFiles);
  const [selectedDraftPath, setSelectedDraftPath] = useState('main.py');
  const [newDraftPath, setNewDraftPath] = useState('');
  const [draftInputs, setDraftInputs] = useState('{\n  "asset_id": "edge-test-device"\n}');
  const [draftOutput, setDraftOutput] = useState<unknown>(null);
  const [draftRunning, setDraftRunning] = useState(false);
  const [generatingReadme, setGeneratingReadme] = useState(false);
  const [loadingExistingFiles, setLoadingExistingFiles] = useState(false);

  const isReplaceMode = !!id;
  const selectedDraftFile = draftFiles.find((draftFile) => draftFile.path === selectedDraftPath) || draftFiles[0];
  const draftPyFiles = draftFiles
    .filter((draftFile) => draftFile.path.toLowerCase().endsWith('.py'))
    .map((draftFile) => draftFile.path);
  const backPath = existingScript ? `/script-hub/${existingScript.id}` : '/script-hub';

  useEffect(() => {
    if (!id) {
      setExistingScript(null);
      setMode('zip');
      setMainFile('main.py');
      return;
    }

    let cancelled = false;
    const loadExistingScript = async () => {
      setLoadingExistingFiles(true);
      setInstallError(null);
      try {
        const scriptRes: any = await scriptsApi.getById(id);
        const script = scriptRes.data as DownloadedScriptReturningValues;
        if (cancelled) return;
        setExistingScript(script);
        setMode('editor');
        setName(script.name || '');
        setDescription(script.description || '');
        setAuthor(script.author || '');
        setVersion(script.version || '1.0.0');
        setMainFile(script.main_file || 'main.py');

        const filesRes: any = await scriptsApi.listFiles(id);
        if (cancelled) return;
        const files = Array.isArray(filesRes.data) ? filesRes.data as ScriptDraftFile[] : [];
        if (files.length > 0) {
          setDraftFiles(files);
          const preferred = files.find((draftFile) => draftFile.path === script.main_file) || files[0];
          setSelectedDraftPath(preferred.path);
        }
      } catch (err: any) {
        if (!cancelled) {
          setInstallError(err.response?.data?.error || err.response?.data?.message || err.message || 'Erro ao carregar arquivos do script.');
        }
      } finally {
        if (!cancelled) {
          setLoadingExistingFiles(false);
        }
      }
    };

    void loadExistingScript();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const updateSelectedDraftContent = (content: string) => {
    setDraftFiles((current) => current.map((draftFile) => (
      draftFile.path === selectedDraftPath ? { ...draftFile, content } : draftFile
    )));
  };

  const addDraftFile = () => {
    const cleanPath = newDraftPath.trim().replace(/\\/g, '/');
    if (!cleanPath || draftFiles.some((draftFile) => draftFile.path === cleanPath)) {
      return;
    }
    const content = cleanPath.toLowerCase().endsWith('requirements.txt') ? 'sparkit\n' : '';
    setDraftFiles((current) => [...current, { path: cleanPath, content }]);
    setSelectedDraftPath(cleanPath);
    setNewDraftPath('');
  };

  const removeDraftFile = (pathValue: string) => {
    if (draftFiles.length <= 1 || pathValue === mainFile) {
      return;
    }
    setDraftFiles((current) => current.filter((draftFile) => draftFile.path !== pathValue));
    if (selectedDraftPath === pathValue) {
      const nextFile = draftFiles.find((draftFile) => draftFile.path !== pathValue);
      setSelectedDraftPath(nextFile?.path || 'main.py');
    }
  };

  const parseDraftInputs = () => {
    const trimmed = draftInputs.trim();
    if (!trimmed) {
      return {};
    }
    return JSON.parse(trimmed);
  };

  const buildDraftPayload = (inputs?: Record<string, unknown>): ScriptDraftPayload => ({
    files: draftFiles,
    mainFile,
    name: name || 'Novo Script SparkEdge',
    description,
    author: author || 'unknown',
    version: version || '1.0.0',
    inputs,
  });

  const handleInspect = async () => {
    if (!file) return;

    setUploading(true);
    setInstallError(null);
    setProgress([
      { step: 'Descompactando e validando', status: 'loading' },
      { step: 'Lendo arquivos', status: 'pending' },
    ]);

    try {
      const formData = new FormData();
      formData.append('file', file);

      const res: any = await scriptsApi.uploadInspect(formData);
      if (res.data) {
        setInspectData(res.data);
        setProgress([
          { step: 'Descompactando e validando', status: 'success' },
          { step: 'Lendo arquivos', status: 'success' },
        ]);
        if (res.data.pyFiles && res.data.pyFiles.length > 0) {
          const hasMain = res.data.pyFiles.find((pathValue: string) => (
            pathValue === 'main.py' || pathValue.endsWith('/main.py')
          ));
          setMainFile(hasMain || res.data.pyFiles[0]);
        }
        setTimeout(() => setStep(2), 600);
      }
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setInstallError(msg);
      setProgress((current) => current.map((item) => (
        item.status === 'loading' ? { ...item, status: 'error' } : item
      )));
    } finally {
      setUploading(false);
    }
  };

  const finishAndNavigate = (scriptID?: string) => {
    const target = scriptID || existingScript?.id;
    navigate(target ? `/script-hub/${target}` : '/script-hub');
  };

  const handleFinalize = async () => {
    if (!inspectData?.tempFolder || !mainFile) {
      return;
    }

    setUploading(true);
    setInstallError(null);
    setProgress([
      { step: 'Criando ambiente virtual', status: 'loading' },
      { step: 'Instalando dependências (Sparkit)', status: 'pending' },
      { step: 'Analisando schema', status: 'pending' },
    ]);

    try {
      const payload = {
        tempFolder: inspectData.tempFolder,
        mainFile,
        name,
        description,
        author,
        version,
      };
      const res: any = isReplaceMode && existingScript
        ? await scriptsApi.replaceUploadFinalize(existingScript.id, payload)
        : await scriptsApi.uploadFinalize(payload);

      if (res.error) {
        throw new Error(res.error);
      }

      setProgress([
        { step: 'Criando ambiente virtual', status: 'success' },
        { step: 'Instalando dependências (Sparkit)', status: 'success' },
        { step: 'Analisando schema', status: 'success' },
      ]);

      setTimeout(() => finishAndNavigate(res.data?.script?.id), 700);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setInstallError(msg);
      setProgress((current) => current.map((item) => (
        item.status === 'loading' ? { ...item, status: 'error' } : item
      )));
    } finally {
      setUploading(false);
    }
  };

  const handleDraftFinalize = async () => {
    if (!mainFile) {
      setInstallError('Informe o arquivo principal do script.');
      return;
    }

    setUploading(true);
    setInstallError(null);
    setProgress([
      { step: 'Preparando arquivos do editor', status: 'loading' },
      { step: 'Criando ambiente virtual', status: 'pending' },
      { step: 'Instalando dependências e extraindo schema', status: 'pending' },
    ]);

    try {
      const payload = buildDraftPayload();
      const res: any = isReplaceMode && existingScript
        ? await scriptsApi.replaceDraftFinalize(existingScript.id, payload)
        : await scriptsApi.draftFinalize(payload);

      if (res.error) {
        throw new Error(res.error);
      }

      setProgress([
        { step: 'Preparando arquivos do editor', status: 'success' },
        { step: 'Criando ambiente virtual', status: 'success' },
        { step: 'Instalando dependências e extraindo schema', status: 'success' },
      ]);

      setTimeout(() => finishAndNavigate(res.data?.script?.id), 700);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setInstallError(msg);
      setProgress((current) => current.map((item) => (
        item.status === 'loading' ? { ...item, status: 'error' } : item
      )));
    } finally {
      setUploading(false);
    }
  };

  const handleRunDraft = async () => {
    setDraftRunning(true);
    setInstallError(null);
    setDraftOutput(null);
    try {
      const inputs = parseDraftInputs();
      const res: any = await scriptsApi.runDraftPlayground(buildDraftPayload(inputs));
      setDraftOutput(res.data ?? res);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setInstallError(msg);
      setDraftOutput({ error: msg });
    } finally {
      setDraftRunning(false);
    }
  };

  const handleGenerateReadme = async () => {
    setGeneratingReadme(true);
    setInstallError(null);
    try {
      const res: any = await scriptsApi.generateDraftReadme(buildDraftPayload());
      const readme = res.data?.readme || res.readme || '';
      if (!readme) {
        throw new Error('O script nao retornou conteudo para o README.');
      }
      const readmePath = draftFiles.find((draftFile) => draftFile.path.toLowerCase() === 'readme.md')?.path || 'README.md';
      setDraftFiles((current) => {
        const hasReadme = current.some((draftFile) => draftFile.path.toLowerCase() === 'readme.md');
        if (hasReadme) {
          return current.map((draftFile) => (
            draftFile.path.toLowerCase() === 'readme.md' ? { ...draftFile, content: readme } : draftFile
          ));
        }
        return [...current, { path: 'README.md', content: readme }];
      });
      setSelectedDraftPath(readmePath);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      setInstallError(msg);
    } finally {
      setGeneratingReadme(false);
    }
  };

  const renderProgress = () => (
    (uploading || installError || progress.length > 0) && (
      <div className="mt-2 space-y-2 bg-white/[0.04] p-4 rounded-xl border border-white/[0.08]">
        {progress.map((item, index) => (
          <div key={index} className="flex items-center justify-between text-xs">
            <span className={item.status === 'pending' ? 'text-zinc-600' : 'text-zinc-300'}>
              {item.step}
            </span>
            {item.status === 'loading' && (
              <div className="w-3 h-3 border border-violet-400 border-t-transparent rounded-full animate-spin" />
            )}
            {item.status === 'success' && <CheckCircle2 className="w-3 h-3 text-green-500" />}
            {item.status === 'error' && <X className="w-3 h-3 text-red-500" />}
          </div>
        ))}
        {installError && (
          <div className="mt-2 text-[10px] text-red-400 font-mono bg-red-400/10 p-2 rounded border border-red-400/20 whitespace-pre-wrap">
            {installError}
          </div>
        )}
      </div>
    )
  );

  const renderMetadataFields = () => (
    <div className="space-y-4">
      <div className="grid gap-2">
        <Label className="text-zinc-400">Nome do Script</Label>
        <Input
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="Ex.: Sensor Data Collector"
          className="bg-white/[0.04] border-white/[0.1] text-white"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="grid gap-2">
          <Label className="text-zinc-400">Autor</Label>
          <Input
            value={author}
            onChange={(event) => setAuthor(event.target.value)}
            placeholder="Nome/Empresa"
            className="bg-white/[0.04] border-white/[0.1] text-white"
          />
        </div>
        <div className="grid gap-2">
          <Label className="text-zinc-400">Versão</Label>
          <Input
            value={version}
            onChange={(event) => setVersion(event.target.value)}
            placeholder="1.0.0"
            className="bg-white/[0.04] border-white/[0.1] text-white"
          />
        </div>
      </div>

      <div className="grid gap-2">
        <Label className="text-zinc-400">Descrição (opcional)</Label>
        <Input
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          placeholder="Exemplo de script..."
          className="bg-white/[0.04] border-white/[0.1] text-white"
        />
      </div>
    </div>
  );

  return (
    <main className="min-h-screen bg-[#09090b] px-6 py-8 text-zinc-400">
      <div className="mx-auto max-w-7xl space-y-6">
        <div className="flex flex-col gap-4 border-b border-white/[0.06] pb-6 md:flex-row md:items-end md:justify-between">
          <div>
            <Link
              to={backPath}
              className="mb-5 inline-flex items-center gap-2 text-xs text-zinc-500 transition-colors hover:text-white"
            >
              <ChevronLeft size={14} />
              Voltar
            </Link>
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/[0.08] bg-gradient-to-br from-violet-500/20 to-blue-500/20">
                <Code2 className="h-6 w-6 text-violet-300" />
              </div>
              <div>
                <h1 className="text-3xl font-bold tracking-tight text-white">
                  {isReplaceMode ? 'Editar Arquivos do Script' : 'Adicionar Novo Script'}
                </h1>
                <p className="text-sm text-zinc-500">
                  {isReplaceMode
                    ? 'Edite o bundle instalado, teste no playground e salve uma nova versão no histórico.'
                    : 'Importe um .zip ou escreva o script diretamente na WebUI.'}
                </p>
              </div>
            </div>
          </div>
          <Button
            variant="outline"
            asChild
            className="border-white/[0.1] bg-transparent text-white hover:bg-white/[0.06]"
          >
            <Link to={backPath}>Cancelar</Link>
          </Button>
        </div>

        <Tabs value={mode} onValueChange={(value) => setMode(value as InstallMode)} className="space-y-6">
          {!isReplaceMode && (
            <TabsList className="grid w-full grid-cols-2 border border-white/[0.08] bg-white/[0.04] md:w-[520px]">
              <TabsTrigger value="zip" className="data-[state=active]:bg-white/[0.08] data-[state=active]:text-white">
                <UploadCloud className="h-4 w-4" />
                Importar .zip
              </TabsTrigger>
              <TabsTrigger value="editor" className="data-[state=active]:bg-white/[0.08] data-[state=active]:text-white">
                <Code2 className="h-4 w-4" />
                Escrever Script
              </TabsTrigger>
            </TabsList>
          )}

          <TabsContent value="zip" className="space-y-6">
            {step === 1 && (
              <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
                <div className="relative flex min-h-[360px] justify-center rounded-2xl border-2 border-dashed border-white/[0.1] p-10 transition-colors hover:bg-white/[0.02]">
                  <input
                    type="file"
                    accept=".zip"
                    onChange={(event) => setFile(event.target.files?.[0] || null)}
                    className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
                  />
                  <div className="pointer-events-none flex flex-col items-center justify-center gap-3 text-center">
                    <UploadCloud className="h-12 w-12 text-zinc-400" />
                    <p className="text-base font-medium text-white">
                      {file ? file.name : 'Clique ou arraste o .zip do script'}
                    </p>
                    <p className="max-w-md text-sm text-zinc-500">
                      Deve conter o código Python, requirements.txt e README.md. Depois você escolhe o arquivo principal.
                    </p>
                  </div>
                </div>

                <div className="rounded-2xl border border-white/[0.08] bg-white/[0.03] p-5">
                  <h2 className="mb-2 text-sm font-semibold text-white">Instalação via pacote</h2>
                  <p className="mb-4 text-xs leading-6 text-zinc-500">
                    O SparkEdge descompacta, valida a presença do Sparkit, cria a venv e extrai o schema antes de salvar.
                  </p>
                  {renderProgress()}
                  <Button
                    className="mt-4 w-full bg-violet-600 text-white hover:bg-violet-700"
                    disabled={!file || uploading}
                    onClick={handleInspect}
                  >
                    {uploading ? 'Processando...' : 'Avançar'}
                    {!uploading && <ChevronRight className="ml-2 h-4 w-4" />}
                  </Button>
                </div>
              </div>
            )}

            {step === 2 && inspectData && (
              <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
                <div className="rounded-2xl border border-white/[0.08] bg-white/[0.03] p-6">
                  <div className="mb-6">
                    <h2 className="text-xl font-semibold text-white">Configurar pacote</h2>
                    <p className="text-sm text-zinc-500">Complete os metadados e escolha o entrypoint.</p>
                  </div>
                  <div className="space-y-4">
                    {renderMetadataFields()}
                    <div className="grid gap-2">
                      <Label className="text-zinc-400">Entrypoint (Main File)</Label>
                      <Select value={mainFile} onValueChange={setMainFile}>
                        <SelectTrigger className="border-white/[0.1] bg-white/[0.04] text-white">
                          <SelectValue placeholder="Selecione o arquivo principal" />
                        </SelectTrigger>
                        <SelectContent className="border-white/[0.1] bg-[#121214]">
                          {inspectData.pyFiles.map((pathValue) => (
                            <SelectItem key={pathValue} value={pathValue} className="text-white focus:bg-white/[0.06]">
                              {pathValue}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                </div>

                <div className="rounded-2xl border border-white/[0.08] bg-white/[0.03] p-5">
                  {renderProgress()}
                  <div className="mt-4 flex gap-2">
                    <Button
                      variant="outline"
                      onClick={() => setStep(1)}
                      className="flex-1 border-white/[0.1] bg-transparent text-white hover:bg-white/[0.05]"
                    >
                      Voltar
                    </Button>
                    <Button
                      onClick={handleFinalize}
                      disabled={uploading || !mainFile}
                      className="flex-1 bg-violet-600 text-white hover:bg-violet-700"
                    >
                      {uploading ? 'Instalando...' : isReplaceMode ? 'Atualizar Script' : 'Finalizar e Instalar'}
                      {!uploading && <CheckCircle2 className="ml-2 h-4 w-4" />}
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </TabsContent>

          <TabsContent value="editor" className="space-y-6">
            {loadingExistingFiles && (
              <div className="rounded-xl border border-violet-400/20 bg-violet-400/10 px-4 py-3 text-sm text-violet-100">
                Carregando arquivos atuais do script para edição...
              </div>
            )}

            <ResizablePanelGroup direction="horizontal" className="min-h-[760px] rounded-2xl border border-white/[0.08] bg-white/[0.02]">
              <ResizablePanel defaultSize={22} minSize={16} maxSize={34} className="min-w-0">
              <div className="h-full space-y-4 overflow-y-auto border-r border-white/[0.08] bg-white/[0.03] p-4">
                <div>
                  <Label className="text-zinc-400">Arquivos</Label>
                  <div className="mt-2 space-y-2">
                    {draftFiles.map((draftFile) => (
                      <button
                        key={draftFile.path}
                        type="button"
                        onClick={() => setSelectedDraftPath(draftFile.path)}
                        className={`flex w-full items-center justify-between rounded-lg border px-3 py-2 text-left text-xs transition-colors ${
                          selectedDraftPath === draftFile.path
                            ? 'border-violet-400/50 bg-violet-400/10 text-white'
                            : 'border-white/[0.08] bg-black/20 text-zinc-400 hover:text-white'
                        }`}
                      >
                        <span className="flex min-w-0 items-center gap-2">
                          <FileText className="h-3.5 w-3.5 shrink-0" />
                          <span className="truncate">{draftFile.path}</span>
                        </span>
                        <span
                          role="button"
                          tabIndex={0}
                          onClick={(event) => {
                            event.stopPropagation();
                            removeDraftFile(draftFile.path);
                          }}
                          onKeyDown={(event) => {
                            if (event.key === 'Enter' || event.key === ' ') {
                              event.preventDefault();
                              event.stopPropagation();
                              removeDraftFile(draftFile.path);
                            }
                          }}
                          className={`rounded p-1 ${
                            draftFile.path === mainFile || draftFiles.length <= 1
                              ? 'cursor-not-allowed text-zinc-700'
                              : 'text-zinc-500 hover:bg-red-500/10 hover:text-red-300'
                          }`}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </span>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="flex gap-2">
                  <Input
                    value={newDraftPath}
                    onChange={(event) => setNewDraftPath(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === 'Enter') {
                        event.preventDefault();
                        addDraftFile();
                      }
                    }}
                    placeholder="utils/helper.py"
                    className="border-white/[0.1] bg-white/[0.04] text-white"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={addDraftFile}
                    className="border-white/[0.1] bg-transparent text-white hover:bg-white/[0.06]"
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>

                <div className="grid gap-2">
                  <Label className="text-zinc-400">Entrypoint</Label>
                  <Select value={mainFile} onValueChange={setMainFile}>
                    <SelectTrigger className="border-white/[0.1] bg-white/[0.04] text-white">
                      <SelectValue placeholder="Selecione o arquivo principal" />
                    </SelectTrigger>
                    <SelectContent className="border-white/[0.1] bg-[#121214]">
                      {draftPyFiles.map((pathValue) => (
                        <SelectItem key={pathValue} value={pathValue} className="text-white focus:bg-white/[0.06]">
                          {pathValue}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className="text-[11px] text-zinc-500">
                    O SparkEdge cria a venv, instala o requirements e extrai o schema desse arquivo.
                  </p>
                </div>
              </div>
              </ResizablePanel>

              <ResizableHandle withHandle className="bg-white/[0.08]" />
              <ResizablePanel defaultSize={50} minSize={30} className="min-w-0">
              <div className="h-full space-y-3 overflow-y-auto p-4">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium text-white">{selectedDraftFile?.path}</p>
                    <p className="text-xs text-zinc-500">Editor simples para código, requirements ou README.</p>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleGenerateReadme}
                    disabled={generatingReadme || !mainFile}
                    className="border-white/[0.1] bg-transparent text-white hover:bg-white/[0.06]"
                  >
                    <BookOpen className="mr-2 h-4 w-4" />
                    {generatingReadme ? 'Gerando...' : 'Gerar README'}
                  </Button>
                </div>
                <Textarea
                  value={selectedDraftFile?.content || ''}
                  onChange={(event) => updateSelectedDraftContent(event.target.value)}
                  spellCheck={false}
                  className="min-h-[660px] resize-y border-white/[0.1] bg-[#050506] font-mono text-xs leading-5 text-zinc-100"
                />
              </div>
              </ResizablePanel>

              <ResizableHandle withHandle className="bg-white/[0.08]" />
              <ResizablePanel defaultSize={28} minSize={20} maxSize={42} className="min-w-0">
              <div className="h-full space-y-4 overflow-y-auto border-l border-white/[0.08] bg-white/[0.03] p-4">
                {renderMetadataFields()}

                <div className="grid gap-2">
                  <Label className="text-zinc-400">Inputs do Playground (JSON)</Label>
                  <Textarea
                    value={draftInputs}
                    onChange={(event) => setDraftInputs(event.target.value)}
                    spellCheck={false}
                    className="min-h-28 resize-y border-white/[0.1] bg-black/30 font-mono text-xs text-zinc-100"
                  />
                </div>

                <Button
                  type="button"
                  onClick={handleRunDraft}
                  disabled={draftRunning || !mainFile}
                  className="w-full bg-emerald-600 text-white hover:bg-emerald-700"
                >
                  <Play className="mr-2 h-4 w-4" />
                  {draftRunning ? 'Executando...' : 'Rodar no Playground'}
                </Button>

                {draftOutput !== null && (
                  <div className="rounded-xl border border-white/[0.08] bg-black/40 p-3">
                    <p className="mb-2 text-xs font-medium text-zinc-300">Resultado</p>
                    <pre className="max-h-56 overflow-auto whitespace-pre-wrap text-[11px] leading-5 text-zinc-200">
                      {JSON.stringify(draftOutput, null, 2)}
                    </pre>
                  </div>
                )}

                {renderProgress()}

                <Button
                  onClick={handleDraftFinalize}
                  disabled={uploading || !mainFile || draftPyFiles.length === 0}
                  className="w-full bg-violet-600 text-white hover:bg-violet-700"
                >
                  {uploading ? 'Instalando...' : isReplaceMode ? 'Salvar Nova Versão' : 'Finalizar e Instalar'}
                  {!uploading && <CheckCircle2 className="ml-2 h-4 w-4" />}
                </Button>
              </div>
              </ResizablePanel>
            </ResizablePanelGroup>
          </TabsContent>
        </Tabs>
      </div>
    </main>
  );
}
