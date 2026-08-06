import { useState } from 'react';
import {
  Clock, CheckCircle2, XCircle, Loader2, Timer,
  ChevronDown
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

// Types duplicated here or imported from a common place. 
// For now, I'll use the one from the service types if possible, 
// but ExecutionHistory uses a local type. Let's see.

export type ExecutionStatus = 'queued' | 'running' | 'success' | 'failed' | 'timeout';

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
  output?: any;
  logs?: Array<{
    timestamp: string;
    level: 'info' | 'warn' | 'error';
    message: string;
  }>;
}

const statusMap: Record<ExecutionStatus, { icon: React.ElementType; color: string; bg: string; label: string }> = {
  queued:  { icon: Clock,        color: 'text-zinc-400',   bg: 'bg-zinc-500/10',   label: 'Na fila' },
  running: { icon: Loader2,      color: 'text-blue-400',   bg: 'bg-blue-500/10',   label: 'Executando' },
  success: { icon: CheckCircle2, color: 'text-emerald-400', bg: 'bg-emerald-500/10', label: 'Sucesso' },
  failed:  { icon: XCircle,      color: 'text-red-400',    bg: 'bg-red-500/10',    label: 'Falha' },
  timeout: { icon: Timer,        color: 'text-amber-400',  bg: 'bg-amber-500/10',  label: 'Timeout' },
};

export function ExecutionRow({ execution }: { execution: InstanceExecution }) {
  const [expanded, setExpanded] = useState(false);
  const cfg = statusMap[execution.status] || statusMap.queued;
  const Icon = cfg.icon;

  const duration = execution.duration_ms
    ? execution.duration_ms >= 1000
      ? `${(execution.duration_ms / 1000).toFixed(1)}s`
      : `${execution.duration_ms}ms`
    : '—';

  return (
    <motion.div layout className="border-b border-white/[0.04] last:border-0 overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-4 px-5 py-4 text-left hover:bg-white/[0.02] transition-colors"
      >
        <div className={`p-2 rounded-lg ${cfg.bg}`}>
          <Icon size={14} className={`${cfg.color} ${execution.status === 'running' ? 'animate-spin' : ''}`} />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-xs font-mono text-zinc-500">{execution.id.slice(0, 8)}</span>
            <span className={`text-xs font-medium ${cfg.color}`}>{cfg.label}</span>
          </div>
          <p className="text-[11px] text-zinc-600 mt-0.5 truncate">
            Iniciado em: {execution.started_at ? new Date(execution.started_at).toLocaleString('pt-BR') : '—'}
          </p>
        </div>

        <div className="flex items-center gap-6 text-[11px] text-zinc-500">
          <div className="flex items-center gap-1">
            <Clock size={11} />
            <span>{duration}</span>
          </div>
          <span className="capitalize px-2 py-0.5 rounded-md bg-white/[0.04] text-zinc-500">{execution.trigger_type}</span>
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
            <div className="px-5 pb-5 pt-1 space-y-3">
              {/* Timestamps */}
              <div className="grid grid-cols-3 gap-3">
                {[
                  ['Iniciado', execution.started_at],
                  ['Finalizado', execution.finished_at],
                  ['Duração', duration],
                ].map(([label, val]) => (
                  <div key={label as string} className="bg-white/[0.02] rounded-lg p-3">
                    <p className="text-[10px] text-zinc-600 uppercase tracking-wider">{label}</p>
                    <p className="text-xs text-white font-mono mt-1">{val && label !== 'Duração' ? new Date(val as string).toLocaleString('pt-BR') : val || '—'}</p>
                  </div>
                ))}
              </div>

              {/* Delivery status */}
              <div className="flex items-center gap-4 text-xs">
                <div className={`flex items-center gap-1 ${execution.destination_sent ? 'text-emerald-400' : 'text-zinc-500'}`}>
                  {execution.destination_sent ? <CheckCircle2 size={12} /> : <XCircle size={12} />}
                  Destino {execution.destination_sent ? 'enviado' : 'não enviado'}
                </div>
                {execution.fallback_used && (
                  <div className="flex items-center gap-1 text-amber-400">
                    <Clock size={12} />
                    Fallback utilizado
                  </div>
                )}
              </div>

              {/* Error */}
              {execution.error_message && (
                <div className="bg-red-500/[0.06] border border-red-500/[0.1] rounded-lg p-3">
                  <p className="text-[10px] text-red-400 uppercase tracking-wider mb-1">Erro</p>
                  <p className="text-xs text-red-300 font-mono">{execution.error_message}</p>
                </div>
              )}

              {/* Output */}
              {execution.output && (
                <div className="bg-white/[0.02] rounded-lg p-3">
                  <p className="text-[10px] text-zinc-500 uppercase tracking-wider mb-2">Output</p>
                  <pre className="text-xs text-zinc-300 font-mono whitespace-pre-wrap max-h-48 overflow-auto">
                    {typeof execution.output === 'string' ? execution.output : JSON.stringify(execution.output, null, 2)}
                  </pre>
                </div>
              )}

              {/* Logs */}
              {execution.logs && execution.logs.length > 0 && (
                <div className="bg-white/[0.02] rounded-lg p-3">
                  <p className="text-[10px] text-zinc-500 uppercase tracking-wider mb-2">Logs ({execution.logs.length})</p>
                  <div className="space-y-1 max-h-48 overflow-auto font-mono text-[11px]">
                    {execution.logs.map((log, idx) => (
                      <div key={idx} className={`flex items-start gap-2 ${
                        log.level === 'error' ? 'text-red-400' : log.level === 'warn' ? 'text-amber-400' : 'text-zinc-400'
                      }`}>
                        <span className="text-zinc-600 shrink-0">{new Date(log.timestamp).toLocaleTimeString('pt-BR')}</span>
                        <span className="uppercase text-[9px] font-bold w-10 shrink-0">{log.level}</span>
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

