import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { updateService, type UpdateApplyResult, type UpdateCheckResult, type UpdateDownloadResult, type UpdateHistoryEntry, type UpdateRestartResult, type UpdateRollbackResult, type UpdateState } from '@/rest-api-client/update.service';
import { AlertTriangle, CheckCircle2, Clock3, Download, Loader2, RefreshCw, Rocket, ShieldCheck, RotateCcw, Wrench } from 'lucide-react';
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

function historyTone(status?: string) {
  switch (status) {
    case 'completed':
    case 'applied':
    case 'executed':
      return 'border-emerald-500/20 bg-emerald-500/[0.06] text-emerald-200';
    case 'prepared':
    case 'planned':
    case 'manual_required':
      return 'border-amber-500/20 bg-amber-500/[0.06] text-amber-200';
    default:
      return 'border-white/[0.06] bg-black/20 text-zinc-300';
  }
}

function historyLabel(entry: UpdateHistoryEntry) {
  const kind = entry.type === 'download'
    ? 'Download'
    : entry.type === 'apply'
      ? 'Apply'
      : entry.type === 'rollback'
        ? 'Rollback'
        : entry.type === 'restart'
          ? 'Restart'
          : entry.type;
  return `${kind} | ${entry.status}`;
}

export default function UpdateSettingsPage() {
  const [loading, setLoading] = useState(true);
  const [checking, setChecking] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [applying, setApplying] = useState(false);
  const [rollingBack, setRollingBack] = useState(false);
  const [restarting, setRestarting] = useState(false);
  const [result, setResult] = useState<UpdateCheckResult | null>(null);
  const [downloadResult, setDownloadResult] = useState<UpdateDownloadResult | null>(null);
  const [applyResult, setApplyResult] = useState<UpdateApplyResult | null>(null);
  const [updateState, setUpdateState] = useState<UpdateState | null>(null);
  const [rollbackResult, setRollbackResult] = useState<UpdateRollbackResult | null>(null);
  const [restartResult, setRestartResult] = useState<UpdateRestartResult | null>(null);

  const loadStatus = async () => {
    const response = await updateService.status();
    setUpdateState(response.data.data);
    if (response.data.data.last_download_result) {
      setDownloadResult(response.data.data.last_download_result);
    }
    if (response.data.data.last_apply_result) {
      setApplyResult(response.data.data.last_apply_result);
    }
    if (response.data.data.last_rollback_result) {
      setRollbackResult(response.data.data.last_rollback_result);
    }
    if (response.data.data.last_restart_result) {
      setRestartResult(response.data.data.last_restart_result);
    }
  };

  const loadCheck = async () => {
    setChecking(true);
    try {
      const [checkResponse] = await Promise.all([
        updateService.check(),
        loadStatus(),
      ]);
      setResult(checkResponse.data.data);
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

  const handleDownload = async () => {
    setDownloading(true);
    try {
      const response = await updateService.download();
      setDownloadResult(response.data.data);
      toast.success('Pacote de atualizacao baixado com sucesso.');
      await loadCheck();
    } catch (error: any) {
      toast.error(`Falha ao baixar pacote: ${error?.message ?? 'Erro desconhecido'}`);
    } finally {
      setDownloading(false);
    }
  };

  const handleApply = async () => {
    if (!downloadResult?.downloaded_path) {
      toast.error('Baixe um pacote compativel antes de aplicar a atualizacao.');
      return;
    }
    setApplying(true);
    try {
      const response = await updateService.apply(downloadResult.downloaded_path);
      setApplyResult(response.data.data);
      toast.success('Atualizacao assistida preparada com sucesso.');
    } catch (error: any) {
      toast.error(`Falha ao preparar aplicacao: ${error?.message ?? 'Erro desconhecido'}`);
    } finally {
      setApplying(false);
    }
  };

  const handleRollback = async () => {
    setRollingBack(true);
    try {
      const response = await updateService.rollback();
      setRollbackResult(response.data.data);
      toast.success('Rollback assistido processado.');
      await loadStatus();
    } catch (error: any) {
      toast.error(`Falha ao preparar rollback: ${error?.message ?? 'Erro desconhecido'}`);
    } finally {
      setRollingBack(false);
    }
  };

  const handleRestart = async (execute: boolean) => {
    setRestarting(true);
    try {
      const response = await updateService.restart(execute);
      setRestartResult(response.data.data);
      toast.success(execute ? 'Reinicio disparado.' : 'Plano de reinicio gerado.');
      await loadStatus();
    } catch (error: any) {
      toast.error(`Falha ao processar reinicio: ${error?.message ?? 'Erro desconhecido'}`);
    } finally {
      setRestarting(false);
    }
  };

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
              O fluxo cobre consulta, preparo, rollback e reinicio assistido com estado persistido para acompanhar cada etapa.
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
              <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4">
                <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500 mb-1">Canal</p>
                <p className="text-sm font-medium capitalize text-zinc-200">{result?.channel ?? 'stable'}</p>
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
                <p className="text-xs text-zinc-500">
                  Integridade: {result?.integrity_ready ? 'manifesto e checksum disponiveis' : 'manifesto/checksum ainda indisponiveis'}
                </p>
              </div>

              <Button
                type="button"
                onClick={() => void handleDownload()}
                disabled={downloading || !result?.update_available || !result?.integrity_ready}
                className="w-full gap-2 bg-sky-400 text-zinc-950 hover:bg-sky-300 font-semibold"
              >
                {downloading ? <Loader2 size={16} className="animate-spin" /> : <Download size={16} />}
                {downloading ? 'Baixando pacote...' : 'Baixar pacote compativel'}
              </Button>

              {downloadResult && (
                <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/[0.07] p-4 space-y-2">
                  <p className="text-sm font-semibold text-emerald-300">Pacote pronto para uso assistido</p>
                  <p className="text-xs text-emerald-100/80 break-all">Arquivo: {downloadResult.downloaded_path}</p>
                  <p className="text-xs text-emerald-100/80">SHA256 verificado: {downloadResult.sha256}</p>
                </div>
              )}

              <Button
                type="button"
                onClick={() => void handleApply()}
                disabled={applying || !downloadResult?.downloaded_path}
                variant="outline"
                className="w-full gap-2 border-white/10 text-white hover:bg-white/5"
              >
                {applying ? <Loader2 size={16} className="animate-spin" /> : <Wrench size={16} />}
                {applying ? 'Preparando aplicacao...' : 'Preparar aplicacao assistida'}
              </Button>

              {applyResult && (
                <div className="rounded-xl border border-amber-500/20 bg-amber-500/[0.07] p-4 space-y-2">
                  <p className="text-sm font-semibold text-amber-300">Aplicacao preparada</p>
                  <p className="text-xs text-amber-100/80">{applyResult.message}</p>
                  <p className="text-xs text-amber-100/80 break-all">Backup: {applyResult.backup_path}</p>
                  <p className="text-xs text-amber-100/80 break-all">Stage: {applyResult.staging_path}</p>
                  {applyResult.script_path && (
                    <p className="text-xs text-amber-100/80 break-all">Script: {applyResult.script_path}</p>
                  )}
                  {applyResult.rollback_path && (
                    <p className="text-xs text-amber-100/80 break-all">Rollback: {applyResult.rollback_path}</p>
                  )}
                  {applyResult.next_steps?.length ? (
                    <div className="pt-2 space-y-1">
                      {applyResult.next_steps.map((step, index) => (
                        <p key={`${index}-${step}`} className="text-xs text-amber-100/80">
                          {index + 1}. {step}
                        </p>
                      ))}
                    </div>
                  ) : null}
                </div>
              )}

              <div className="grid gap-3 md:grid-cols-2">
                <Button
                  type="button"
                  onClick={() => void handleRollback()}
                  disabled={rollingBack || !applyResult}
                  variant="outline"
                  className="gap-2 border-white/10 text-white hover:bg-white/5"
                >
                  {rollingBack ? <Loader2 size={16} className="animate-spin" /> : <RotateCcw size={16} />}
                  {rollingBack ? 'Processando rollback...' : 'Rollback assistido'}
                </Button>
                <Button
                  type="button"
                  onClick={() => void handleRestart(false)}
                  disabled={restarting}
                  variant="outline"
                  className="gap-2 border-white/10 text-white hover:bg-white/5"
                >
                  {restarting ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                  Plano de reinicio
                </Button>
              </div>

              <Button
                type="button"
                onClick={() => void handleRestart(true)}
                disabled={restarting}
                className="w-full gap-2 bg-emerald-400 text-zinc-950 hover:bg-emerald-300 font-semibold"
              >
                {restarting ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                {restarting ? 'Disparando reinicio...' : 'Executar reinicio assistido'}
              </Button>

              {rollbackResult && (
                <div className="rounded-xl border border-rose-500/20 bg-rose-500/[0.07] p-4 space-y-2">
                  <p className="text-sm font-semibold text-rose-300">Rollback</p>
                  <p className="text-xs text-rose-100/80">{rollbackResult.message}</p>
                  <p className="text-xs text-rose-100/80 break-all">Backup: {rollbackResult.backup_path}</p>
                  {rollbackResult.script_path && (
                    <p className="text-xs text-rose-100/80 break-all">Script: {rollbackResult.script_path}</p>
                  )}
                </div>
              )}

              {restartResult && (
                <div className="rounded-xl border border-cyan-500/20 bg-cyan-500/[0.07] p-4 space-y-2">
                  <p className="text-sm font-semibold text-cyan-300">Reinicio</p>
                  <p className="text-xs text-cyan-100/80">{restartResult.message}</p>
                  {restartResult.command && (
                    <p className="text-xs text-cyan-100/80 break-all">Comando: {restartResult.command}</p>
                  )}
                </div>
              )}

              {updateState?.updated_at && (
                <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4 space-y-2">
                  <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500">Estado persistido</p>
                  <p className="text-xs text-zinc-400">
                    Ultima atualizacao do estado: {new Date(updateState.updated_at).toLocaleString('pt-BR')}
                  </p>
                  {updateState.last_prepared_version && (
                    <p className="text-xs text-zinc-400">
                      Ultima versao preparada: {updateState.last_prepared_version} ({updateState.last_prepared_target})
                    </p>
                  )}
                </div>
              )}

              {updateState?.history?.length ? (
                <div className="rounded-xl border border-white/[0.06] bg-black/20 p-4 space-y-3">
                  <p className="text-[11px] uppercase tracking-[0.2em] text-zinc-500">Historico</p>
                  <div className="space-y-2">
                    {[...updateState.history].reverse().map((entry, index) => (
                      <div key={`${entry.created_at}-${index}`} className={`rounded-xl border p-3 ${historyTone(entry.status)}`}>
                        <div className="flex items-start justify-between gap-3">
                          <div>
                            <p className="text-sm font-semibold">{historyLabel(entry)}</p>
                            <p className="text-xs opacity-80">
                              {entry.version || '-'} {entry.target ? `| ${entry.target}` : ''}
                            </p>
                          </div>
                          <p className="text-[11px] opacity-70">
                            {new Date(entry.created_at).toLocaleString('pt-BR')}
                          </p>
                        </div>
                        {entry.message && (
                          <p className="mt-2 text-xs opacity-90">{entry.message}</p>
                        )}
                        {entry.artifact && (
                          <p className="mt-1 break-all text-[11px] opacity-75">{entry.artifact}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}

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
