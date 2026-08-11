import React, { useEffect, useState } from 'react';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { CheckCircle2, ChevronRight, UploadCloud, X } from 'lucide-react';
import { scriptsApi } from '@/rest-api-client/scripts.service';
import type { DownloadedScriptReturningValues } from '@/types/db';

interface AddScriptDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
  existingScript?: DownloadedScriptReturningValues | null;
}

export function AddScriptDialog({
  open,
  onOpenChange,
  onSuccess,
  existingScript = null,
}: AddScriptDialogProps) {
  const [step, setStep] = useState(1);
  const [file, setFile] = useState<File | null>(null);
  const [inspectData, setInspectData] = useState<{ tempFolder: string; pyFiles: string[] } | null>(null);
  const [uploading, setUploading] = useState(false);
  const [installError, setInstallError] = useState<string | null>(null);
  const [progress, setProgress] = useState<
    { step: string; status: 'pending' | 'loading' | 'success' | 'error' }[]
  >([]);
  const [mainFile, setMainFile] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [author, setAuthor] = useState('');
  const [version, setVersion] = useState('1.0.0');

  const isReplaceMode = !!existingScript;

  useEffect(() => {
    if (!open) {
      return;
    }

    if (existingScript) {
      setName(existingScript.name || '');
      setDescription(existingScript.description || '');
      setAuthor(existingScript.author || '');
      setVersion(existingScript.version || '1.0.0');
      setMainFile(existingScript.main_file || '');
    }
  }, [open, existingScript]);

  const reset = () => {
    setStep(1);
    setFile(null);
    setInspectData(null);
    setMainFile('');
    setName('');
    setDescription('');
    setAuthor('');
    setVersion('1.0.0');
    setUploading(false);
    setInstallError(null);
    setProgress([]);
  };

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
      const msg = err.response?.data?.error || err.message;
      setInstallError(msg);
      setProgress((current) => current.map((item) => (
        item.status === 'loading' ? { ...item, status: 'error' } : item
      )));
    } finally {
      setUploading(false);
    }
  };

  const handleFinalize = async () => {
    if (!inspectData?.tempFolder || !mainFile || !name) {
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

      setTimeout(() => {
        onSuccess();
        onOpenChange(false);
        reset();
      }, 1000);
    } catch (err: any) {
      const msg = err.response?.data?.error || err.message;
      setInstallError(msg);
      setProgress((current) => current.map((item) => (
        item.status === 'loading' ? { ...item, status: 'error' } : item
      )));
    } finally {
      setUploading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen);
        if (!nextOpen) {
          reset();
        }
      }}
    >
      <DialogContent className="sm:max-w-md bg-[#09090b] border-white/[0.08]">
        <DialogHeader>
          <DialogTitle className="text-white">
            {isReplaceMode ? 'Atualizar Bundle do Script' : 'Adicionar Novo Script'}
          </DialogTitle>
        </DialogHeader>

        {step === 1 && (
          <div className="space-y-4 py-4">
            <div className="flex justify-center border-2 border-dashed border-white/[0.1] rounded-xl p-8 hover:bg-white/[0.02] transition-colors relative">
              <input
                type="file"
                accept=".zip"
                onChange={(event) => setFile(event.target.files?.[0] || null)}
                className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
              />
              <div className="flex flex-col items-center gap-2 pointer-events-none">
                <UploadCloud className="w-8 h-8 text-zinc-400" />
                <p className="text-sm font-medium text-white">
                  {file ? file.name : 'Clique ou arraste o .zip do script'}
                </p>
                {!file && (
                  <p className="text-xs text-zinc-500">
                    {isReplaceMode
                      ? 'O novo bundle vai substituir os arquivos atuais do script'
                      : 'Deve conter o código Python'}
                  </p>
                )}
              </div>
            </div>

            {(uploading || installError || progress.length > 0) && (
              <div className="mt-4 space-y-2 bg-white/[0.04] p-4 rounded-xl border border-white/[0.08]">
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
                  <div className="mt-2 text-[10px] text-red-400 font-mono bg-red-400/10 p-2 rounded border border-red-400/20">
                    {installError}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {step === 2 && inspectData && (
          <div className="space-y-4 py-4">
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
              <Label className="text-zinc-400">Entrypoint (Main File)</Label>
              <Select value={mainFile} onValueChange={setMainFile}>
                <SelectTrigger className="bg-white/[0.04] border-white/[0.1] text-white">
                  <SelectValue placeholder="Selecione o arquivo principal" />
                </SelectTrigger>
                <SelectContent className="bg-[#121214] border-white/[0.1]">
                  {inspectData.pyFiles.map((pathValue) => (
                    <SelectItem
                      key={pathValue}
                      value={pathValue}
                      className="text-white focus:bg-white/[0.06]"
                    >
                      {pathValue}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
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

            {(uploading || installError || progress.length > 0) && (
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
                  <div className="mt-2 text-[10px] text-red-400 font-mono bg-red-400/10 p-2 rounded border border-red-400/20">
                    {installError}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        <DialogFooter className="flex items-center gap-2 mt-4">
          {step === 1 ? (
            <Button
              className="w-full bg-violet-600 hover:bg-violet-700 text-white"
              disabled={!file || uploading}
              onClick={handleInspect}
            >
              {uploading ? 'Processando...' : 'Avançar'}
              {!uploading && <ChevronRight className="ml-2 w-4 h-4" />}
            </Button>
          ) : (
            <div className="flex w-full gap-2">
              <Button
                variant="outline"
                onClick={() => setStep(1)}
                className="flex-1 bg-transparent border-white/[0.1] text-white hover:bg-white/[0.05]"
              >
                Voltar
              </Button>
              <Button
                onClick={handleFinalize}
                disabled={uploading || !name || !mainFile}
                className="flex-1 bg-violet-600 hover:bg-violet-700 text-white"
              >
                {uploading ? 'Instalando...' : isReplaceMode ? 'Atualizar Script' : 'Finalizar e Instalar'}
                {!uploading && <CheckCircle2 className="ml-2 w-4 h-4" />}
              </Button>
            </div>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
