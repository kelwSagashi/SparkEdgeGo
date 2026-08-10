import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { updateService, type UpdateCheckResult } from '@/rest-api-client/update.service';
import { AlertTriangle, CheckCircle2, Clock3, Download, Loader2, RefreshCw, Rocket, ShieldCheck } from 'lucide-react';
import { toast } from 'sonner';

function formatBytes(size?: number) {
  if (!size || size <= 0) return '-';
  const units = ['B', 'KB', 'MB', 'GB'];
  let value = size;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`;
}

export default function UpdateSettingsPage() {
  const [loading, setLoading] = useState(true);
  const [checking, setChecking] = useState(false);
  const [result, setResult] = useState<UpdateCheckResult | null>(null);

  const loadCheck = async () => {
    setChecking(true);
    try {
      const response = await updateService.check();
      setResult(response.data.data);
    } catch (error: any) {
      toast.error(`Falha ao verificar atualizacao: ${error?.message ?? 'Erro desconhecido'}`);
    } finally {
      setLoading(false);
      setChecking(false);
    }
  };

  useEffect(() => {
    void loadCheck();
  }, []);

  return (
    <main className="grow px-8 py-6 w-full max-w-[860px] mx-auto pb-20">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-10 h-10 rounded-2xl bg-gradient-to-br from-sky-500 to-cyan-400 flex items-center justify-center shadow-lg shadow-sky-500/20">
            <Rocket size={18} className="text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold text-white tracking-tight">Atualizacao Assistida</h1>
            <p className="text-sm text-zinc-500">Consulta segura de novas versoes publicadas no GitHub.</p>
          </div>
        </div>
      </div>

      <div className="mb-6 flex items-start gap-3 bg-sky-500/[0.07] border border-sky-500/20 rounded-2xl px-5 py-4">
        <ShieldCheck size={18} className="text-sky-300 shrink-0 mt-0.5" />
        <div>
          <p className="text-sm font-semibold text-sky-200">Modo assistido</p>
          <p className="text-xs text-sky-100/70 leading-relaxed">
            Nesta fase, o SparkEdge apenas consulta releases compativeis. Nenhum binario e substituido automaticamente.
          </p>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <section className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-white">Versao atual</h2>
            <Button
              type="button"
              onClick={() => void loadCheck()}
              disabled={checking}
              variant="outline"
              className="gap-2 border-white/10 text-white hover:bg-white/5"
            >
              {checking ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
              Verificar
            </Button>
          </div>

          {loading ? (
            <div className="flex items-center gap-2 text-zinc-500">
              <Loader2 size={16} className="animate-spin" />
              <span className="text-sm">Carregando status de atualizacao...</span>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4">
                <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500 mb-1">Versao</p>
                <p className="text-lg font-semibold text-white">{result?.current_version ?? '-'}</p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4">
                <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500 mb-1">Target</p>
                <p className="text-sm font-medium text-zinc-200">{result?.current_target ?? '-'}</p>
              </div>
              <div className="flex items-center gap-2 text-xs text-zinc-500">
                <Clock3 size={13} />
                <span>Ultima consulta: {result?.checked_at ? new Date(result.checked_at).toLocaleString('pt-BR') : '-'}</span>
              </div>
            </div>
          )}
        </section>

        <section className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-6">
          <h2 className="text-sm font-semibold text-white mb-4">Release compativel</h2>

          {loading ? (
            <div className="flex items-center gap-2 text-zinc-500">
              <Loader2 size={16} className="animate-spin" />
              <span className="text-sm">Buscando release...</span>
            </div>
          ) : !result?.enabled ? (
            <div className="rounded-xl border border-amber-500/20 bg-amber-500/[0.06] p-4 text-sm text-amber-200">
              A checagem de atualizacao esta desabilitada no `config.yml`.
            </div>
          ) : (
            <div className="space-y-4">
              <div className={`rounded-xl border p-4 ${result?.update_available ? 'border-emerald-500/30 bg-emerald-500/[0.08]' : 'border-white/[0.06] bg-black/20'}`}>
                <div className="flex items-center gap-2 mb-2">
                  {result?.update_available ? (
                    <CheckCircle2 size={16} className="text-emerald-300" />
                  ) : (
                    <AlertTriangle size={16} className="text-zinc-400" />
                  )}
                  <p className="text-sm font-semibold text-white">
                    {result?.update_available ? 'Atualizacao disponivel' : 'Sem atualizacao pendente'}
                  </p>
                </div>
                <p className="text-xs text-zinc-300/80 leading-relaxed">
                  {result?.compatibility_message ?? 'A release mais recente compativel com a plataforma atual foi identificada com sucesso.'}
                </p>
              </div>

              <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4 space-y-2">
                <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500">Release</p>
                <p className="text-lg font-semibold text-white">{result?.latest_version ?? '-'}</p>
                <p className="text-sm text-zinc-400">{result?.release_name ?? '-'}</p>
              </div>

              <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4 space-y-2">
                <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500">Asset compativel</p>
                <p className="text-sm font-medium text-zinc-200 break-all">{result?.compatible_asset?.name ?? '-'}</p>
                <p className="text-xs text-zinc-500">Tamanho: {formatBytes(result?.compatible_asset?.size)}</p>
              </div>

              {result?.release_url && (
                <a
                  href={result.release_url}
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-2 text-sm text-sky-300 hover:text-sky-200 transition-colors"
                >
                  <Download size={14} />
                  Abrir release no GitHub
                </a>
              )}
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
