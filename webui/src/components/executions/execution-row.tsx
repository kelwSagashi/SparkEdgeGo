import { useMemo, useState } from 'react';
import {
  CheckCircle2,
  ChevronDown,
  Clock,
  GitBranch,
  Loader2,
  PlayCircle,
  RadioTower,
  ShieldAlert,
  Timer,
  XCircle,
} from 'lucide-react';
import { AnimatePresence, motion } from 'framer-motion';

export type ExecutionStatus = 'queued' | 'running' | 'success' | 'failed' | 'timeout';
export type DestinationExecutionStatus = 'success' | 'failed' | 'fallback' | 'skipped' | string;

export interface InstanceExecution {
  id: string;
  instance_id: string;
  status: ExecutionStatus;
  started_at: string;
  finished_at?: string;
  duration_ms?: number;
  trigger_type: string;
  destination_sent: boolean;
  fallback_used: boolean;
  error_message?: string;
  output?: unknown;
  input_payload?: Record<string, unknown> | null;
  output_payload?: Record<string, unknown> | null;
  logs?: Array<{
    timestamp: string;
    level: 'info' | 'warn' | 'error';
    message: string;
  }>;
  destination_details?: Array<{
    destination_id: string;
    resource_operation_id: string;
    server_name?: string;
    resource_name?: string;
    operation_name?: string;
    status: DestinationExecutionStatus;
    payload?: Record<string, unknown>;
    error?: string | null;
    used_fallback?: boolean;
    timestamp?: string;
  }>;
}

type FlowNode = {
  id: string;
  title: string;
  subtitle: string;
  status: 'success' | 'failed' | 'running' | 'pending' | 'warning';
  icon: React.ElementType;
  payload?: Record<string, unknown> | null;
  error?: string | null;
  timestamp?: string;
};

const statusMap: Record<ExecutionStatus, { icon: React.ElementType; color: string; bg: string; label: string }> = {
  queued: { icon: Clock, color: 'text-zinc-400', bg: 'bg-zinc-500/10', label: 'Na fila' },
  running: { icon: Loader2, color: 'text-blue-400', bg: 'bg-blue-500/10', label: 'Executando' },
  success: { icon: CheckCircle2, color: 'text-emerald-400', bg: 'bg-emerald-500/10', label: 'Sucesso' },
  failed: { icon: XCircle, color: 'text-red-400', bg: 'bg-red-500/10', label: 'Falha' },
  timeout: { icon: Timer, color: 'text-amber-400', bg: 'bg-amber-500/10', label: 'Timeout' },
};

const nodeStatusStyles: Record<FlowNode['status'], { dot: string; card: string; text: string }> = {
  success: {
    dot: 'bg-emerald-400 shadow-[0_0_18px_rgba(52,211,153,0.45)]',
    card: 'border-emerald-500/20 bg-emerald-500/8',
    text: 'text-emerald-300',
  },
  failed: {
    dot: 'bg-red-400 shadow-[0_0_18px_rgba(248,113,113,0.45)]',
    card: 'border-red-500/20 bg-red-500/8',
    text: 'text-red-300',
  },
  running: {
    dot: 'bg-blue-400 shadow-[0_0_18px_rgba(96,165,250,0.45)]',
    card: 'border-blue-500/20 bg-blue-500/8',
    text: 'text-blue-300',
  },
  pending: {
    dot: 'bg-zinc-500',
    card: 'border-white/[0.08] bg-white/[0.03]',
    text: 'text-zinc-200',
  },
  warning: {
    dot: 'bg-amber-400 shadow-[0_0_18px_rgba(251,191,36,0.4)]',
    card: 'border-amber-500/20 bg-amber-500/8',
    text: 'text-amber-200',
  },
};

const destinationStatusLabel: Record<string, string> = {
  success: 'Enviado com sucesso',
  failed: 'Falha no envio',
  fallback: 'Enviado para fallback',
  skipped: 'Destino ignorado',
};

const destinationNodeStatus = (status: DestinationExecutionStatus): FlowNode['status'] => {
  switch (status) {
    case 'success':
      return 'success';
    case 'failed':
      return 'failed';
    case 'fallback':
      return 'warning';
    case 'skipped':
      return 'pending';
    default:
      return 'pending';
  }
};

function formatDate(value?: string) {
  if (!value) return '--';
  return new Date(value).toLocaleString('pt-BR');
}

function prettyJson(value: unknown) {
  if (value == null) return '--';
  if (typeof value === 'string') return value;
  return JSON.stringify(value, null, 2);
}

export function ExecutionRow({ execution }: { execution: InstanceExecution }) {
  const [expanded, setExpanded] = useState(false);
  const cfg = statusMap[execution.status] || statusMap.queued;
  const Icon = cfg.icon;

  const duration = execution.duration_ms
    ? execution.duration_ms >= 1000
      ? `${(execution.duration_ms / 1000).toFixed(1)}s`
      : `${execution.duration_ms}ms`
    : '--';

  const flowNodes = useMemo<FlowNode[]>(() => {
    const destinationNodes = (execution.destination_details || []).map((detail, index) => {
      const labelParts = [detail.server_name, detail.resource_name, detail.operation_name].filter(Boolean);
      const title = labelParts.length > 0
        ? labelParts.join(' / ')
        : `Destino ${detail.destination_id.slice(0, 8)}`;

      return {
        id: `${execution.id}-destination-${index}`,
        title,
        subtitle: `${destinationStatusLabel[detail.status] || detail.status} | ${detail.resource_operation_id || 'sem operacao'}`,
        status: destinationNodeStatus(detail.status),
        icon: detail.used_fallback ? ShieldAlert : RadioTower,
        payload: detail.payload || null,
        error: detail.error || null,
        timestamp: detail.timestamp,
      };
    });

    const completionStatus: FlowNode['status'] = execution.status === 'failed'
      ? 'failed'
      : execution.fallback_used
        ? 'warning'
        : execution.status === 'running'
          ? 'running'
          : 'success';

    return [
      {
        id: `${execution.id}-trigger`,
        title: 'Disparo',
        subtitle: `Trigger ${execution.trigger_type}`,
        status: execution.status === 'queued' ? 'pending' : 'success',
        icon: GitBranch,
        timestamp: execution.started_at,
      },
      {
        id: `${execution.id}-script`,
        title: 'Execucao do Script',
        subtitle: execution.error_message ? 'Script terminou com erro' : 'Script executado pelo Sparkit',
        status: execution.status === 'failed' ? 'failed' : execution.status === 'running' ? 'running' : 'success',
        icon: PlayCircle,
      },
      ...destinationNodes,
      {
        id: `${execution.id}-completion`,
        title: 'Conclusao',
        subtitle: execution.fallback_used
          ? 'Execucao concluida com fallback'
          : execution.destination_sent
            ? 'Execucao concluida com envio aos destinos'
            : 'Execucao concluida sem envio',
        status: completionStatus,
        icon: execution.status === 'failed' ? XCircle : CheckCircle2,
        timestamp: execution.finished_at,
      },
    ];
  }, [execution]);

  return (
    <motion.div layout className="overflow-hidden border-b border-white/[0.04] last:border-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-4 px-5 py-4 text-left transition-colors hover:bg-white/[0.02]"
      >
        <div className={`rounded-lg p-2 ${cfg.bg}`}>
          <Icon size={14} className={`${cfg.color} ${execution.status === 'running' ? 'animate-spin' : ''}`} />
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-zinc-500">{execution.id.slice(0, 8)}</span>
            <span className={`text-xs font-medium ${cfg.color}`}>{cfg.label}</span>
          </div>
          <p className="mt-0.5 truncate text-[11px] text-zinc-600">
            Iniciado em: {formatDate(execution.started_at)}
          </p>
        </div>

        <div className="flex items-center gap-6 text-[11px] text-zinc-500">
          <div className="flex items-center gap-1">
            <Clock size={11} />
            <span>{duration}</span>
          </div>
          <span className="rounded-md bg-white/[0.04] px-2 py-0.5 capitalize text-zinc-500">{execution.trigger_type}</span>
          <ChevronDown size={14} className={`transition-transform ${expanded ? 'rotate-180' : ''}`} />
        </div>
      </button>

      <AnimatePresence>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            className="overflow-hidden"
          >
            <div className="space-y-4 px-5 pb-5 pt-1">
              <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
                {[
                  ['Iniciado', execution.started_at],
                  ['Finalizado', execution.finished_at],
                  ['Duracao', duration],
                ].map(([label, value]) => (
                  <div key={label as string} className="rounded-lg bg-white/[0.02] p-3">
                    <p className="text-[10px] uppercase tracking-wider text-zinc-600">{label}</p>
                    <p className="mt-1 font-mono text-xs text-white">
                      {label === 'Duracao' ? value : formatDate(value as string | undefined)}
                    </p>
                  </div>
                ))}
              </div>

              <div className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
                <div className="mb-4">
                  <p className="text-xs font-medium text-white">Fluxo da Execucao</p>
                  <p className="mt-1 text-[11px] text-zinc-500">
                    Visualizacao dos passos da instancia em formato de nos conectados.
                  </p>
                </div>

                <div className="space-y-0">
                  {flowNodes.map((node, index) => {
                    const style = nodeStatusStyles[node.status];
                    const NodeIcon = node.icon;

                    return (
                      <div key={node.id} className="flex gap-4">
                        <div className="shrink-0 flex flex-col items-center">
                          <div className={`mt-1 h-3 w-3 rounded-full ${style.dot}`} />
                          {index < flowNodes.length - 1 && (
                            <div className="min-h-[42px] w-px flex-1 bg-gradient-to-b from-white/[0.18] to-white/[0.06]" />
                          )}
                        </div>

                        <div className={`mb-3 flex-1 rounded-xl border p-4 ${style.card}`}>
                          <div className="flex items-start justify-between gap-3">
                            <div>
                              <div className="flex items-center gap-2">
                                <NodeIcon size={14} className={style.text} />
                                <p className={`text-sm font-medium ${style.text}`}>{node.title}</p>
                              </div>
                              <p className="mt-1 text-xs text-zinc-400">{node.subtitle}</p>
                            </div>
                            {node.timestamp && (
                              <span className="font-mono text-[10px] text-zinc-500">{formatDate(node.timestamp)}</span>
                            )}
                          </div>

                          {node.error && (
                            <div className="mt-3 rounded-lg border border-red-500/20 bg-red-500/8 p-3">
                              <p className="mb-1 text-[10px] uppercase tracking-wider text-red-300">Erro</p>
                              <p className="whitespace-pre-wrap font-mono text-xs text-red-200">{node.error}</p>
                            </div>
                          )}

                          {node.payload && Object.keys(node.payload).length > 0 && (
                            <div className="mt-3 rounded-lg border border-white/[0.08] bg-black/20 p-3">
                              <p className="mb-2 text-[10px] uppercase tracking-wider text-zinc-500">Payload enviado</p>
                              <pre className="max-h-52 overflow-auto whitespace-pre-wrap font-mono text-xs text-zinc-300">
                                {prettyJson(node.payload)}
                              </pre>
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
                <div className="rounded-lg bg-white/[0.02] p-3">
                  <p className="mb-2 text-[10px] uppercase tracking-wider text-zinc-500">Entrada resolvida</p>
                  <pre className="max-h-56 overflow-auto whitespace-pre-wrap font-mono text-xs text-zinc-300">
                    {prettyJson(execution.input_payload)}
                  </pre>
                </div>

                <div className="rounded-lg bg-white/[0.02] p-3">
                  <p className="mb-2 text-[10px] uppercase tracking-wider text-zinc-500">Saida estruturada</p>
                  <pre className="max-h-56 overflow-auto whitespace-pre-wrap font-mono text-xs text-zinc-300">
                    {prettyJson(execution.output_payload)}
                  </pre>
                </div>
              </div>

              <div className="flex flex-wrap items-center gap-4 text-xs">
                <div className={`flex items-center gap-1 ${execution.destination_sent ? 'text-emerald-400' : 'text-zinc-500'}`}>
                  {execution.destination_sent ? <CheckCircle2 size={12} /> : <XCircle size={12} />}
                  Destino {execution.destination_sent ? 'enviado' : 'nao enviado'}
                </div>
                {execution.fallback_used && (
                  <div className="flex items-center gap-1 text-amber-400">
                    <ShieldAlert size={12} />
                    Fallback utilizado
                  </div>
                )}
              </div>

              {execution.error_message && (
                <div className="rounded-lg border border-red-500/[0.1] bg-red-500/[0.06] p-3">
                  <p className="mb-1 text-[10px] uppercase tracking-wider text-red-400">Erro final</p>
                  <p className="whitespace-pre-wrap font-mono text-xs text-red-300">{execution.error_message}</p>
                </div>
              )}

              {execution.output && (
                <div className="rounded-lg bg-white/[0.02] p-3">
                  <p className="mb-2 text-[10px] uppercase tracking-wider text-zinc-500">Saida bruta do script</p>
                  <pre className="max-h-48 overflow-auto whitespace-pre-wrap font-mono text-xs text-zinc-300">
                    {prettyJson(execution.output)}
                  </pre>
                </div>
              )}

              {execution.logs && execution.logs.length > 0 && (
                <div className="rounded-lg bg-white/[0.02] p-3">
                  <p className="mb-2 text-[10px] uppercase tracking-wider text-zinc-500">
                    Logs tecnicos ({execution.logs.length})
                  </p>
                  <div className="max-h-52 space-y-1 overflow-auto font-mono text-[11px]">
                    {execution.logs.map((log, idx) => (
                      <div
                        key={idx}
                        className={`flex items-start gap-2 ${
                          log.level === 'error'
                            ? 'text-red-400'
                            : log.level === 'warn'
                              ? 'text-amber-400'
                              : 'text-zinc-400'
                        }`}
                      >
                        <span className="shrink-0 text-zinc-600">{new Date(log.timestamp).toLocaleTimeString('pt-BR')}</span>
                        <span className="w-10 shrink-0 text-[9px] font-bold uppercase">{log.level}</span>
                        <span>{log.message}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}
