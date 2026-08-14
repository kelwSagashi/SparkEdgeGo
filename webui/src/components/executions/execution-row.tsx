import { useState } from 'react';
import {
  CheckCircle2,
  ChevronDown,
  Clock,
  Loader2,
  ShieldAlert,
  XCircle,
} from 'lucide-react';
import { AnimatePresence, motion } from 'framer-motion';
import { ExecutionFlow } from './execution-flow';

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

const statusMap: Record<ExecutionStatus, { icon: React.ElementType; color: string; bg: string; label: string }> = {
  queued: { icon: Clock, color: 'text-zinc-400', bg: 'bg-zinc-500/10', label: 'Na fila' },
  running: { icon: Loader2, color: 'text-blue-400', bg: 'bg-blue-500/10', label: 'Executando' },
  success: { icon: CheckCircle2, color: 'text-emerald-400', bg: 'bg-emerald-500/10', label: 'Sucesso' },
  failed: { icon: XCircle, color: 'text-red-400', bg: 'bg-red-500/10', label: 'Falha' },
  timeout: { icon: Clock, color: 'text-amber-400', bg: 'bg-amber-500/10', label: 'Timeout' },
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

              <ExecutionFlow execution={execution} />

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
