import { useEffect, useMemo, useState } from 'react';
import {
  Activity,
  AlertTriangle,
  CheckCircle2,
  Clock3,
  FileJson2,
  GitBranch,
  Loader2,
  PlayCircle,
  RadioTower,
  ShieldAlert,
  TerminalSquare,
  Timer,
  XCircle,
} from 'lucide-react';
import type { ExecutionStatus, InstanceExecution } from './execution-row';

type FlowNodeStatus = 'success' | 'failed' | 'running' | 'pending' | 'warning';

type FlowNode = {
  id: string;
  lane: 'core' | 'destination';
  title: string;
  subtitle: string;
  status: FlowNodeStatus;
  icon: React.ElementType;
  timestamp?: string;
  payload?: unknown;
  error?: string | null;
  meta?: Array<{ label: string; value: string }>;
};

const nodeStatusStyles: Record<FlowNodeStatus, { dot: string; card: string; text: string; badge: string }> = {
  success: {
    dot: 'bg-emerald-400 shadow-[0_0_18px_rgba(52,211,153,0.45)]',
    card: 'border-emerald-500/20 bg-emerald-500/8',
    text: 'text-emerald-300',
    badge: 'bg-emerald-500/15 text-emerald-300',
  },
  failed: {
    dot: 'bg-red-400 shadow-[0_0_18px_rgba(248,113,113,0.45)]',
    card: 'border-red-500/20 bg-red-500/8',
    text: 'text-red-300',
    badge: 'bg-red-500/15 text-red-300',
  },
  running: {
    dot: 'bg-blue-400 shadow-[0_0_18px_rgba(96,165,250,0.45)]',
    card: 'border-blue-500/20 bg-blue-500/8',
    text: 'text-blue-300',
    badge: 'bg-blue-500/15 text-blue-300',
  },
  pending: {
    dot: 'bg-zinc-500',
    card: 'border-white/[0.08] bg-white/[0.03]',
    text: 'text-zinc-200',
    badge: 'bg-white/[0.06] text-zinc-300',
  },
  warning: {
    dot: 'bg-amber-400 shadow-[0_0_18px_rgba(251,191,36,0.4)]',
    card: 'border-amber-500/20 bg-amber-500/8',
    text: 'text-amber-200',
    badge: 'bg-amber-500/15 text-amber-200',
  },
};

const executionStatusLabel: Record<ExecutionStatus, string> = {
  queued: 'Na fila',
  running: 'Executando',
  success: 'Sucesso',
  failed: 'Falha',
  timeout: 'Timeout',
};

const destinationStatusLabel: Record<string, string> = {
  success: 'Enviado com sucesso',
  failed: 'Falha no envio',
  fallback: 'Enviado para fallback',
  skipped: 'Destino ignorado',
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

function nodeStatusFromExecution(status: ExecutionStatus): FlowNodeStatus {
  switch (status) {
    case 'success':
      return 'success';
    case 'failed':
      return 'failed';
    case 'running':
      return 'running';
    case 'timeout':
      return 'warning';
    default:
      return 'pending';
  }
}

function nodeStatusFromDestination(status?: string, usedFallback?: boolean): FlowNodeStatus {
  if (usedFallback) return 'warning';
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
}

function inferLogNodeStatus(logs: InstanceExecution['logs']): FlowNodeStatus {
  if (!logs || logs.length === 0) return 'pending';
  if (logs.some((log) => log.level === 'error')) return 'failed';
  if (logs.some((log) => log.level === 'warn')) return 'warning';
  return 'success';
}

function summarizeLogs(logs: InstanceExecution['logs']) {
  if (!logs || logs.length === 0) return [];
  return logs.slice(-8).map((log) => ({
    label: `${new Date(log.timestamp).toLocaleTimeString('pt-BR')} [${log.level.toUpperCase()}]`,
    value: log.message,
  }));
}

function buildNodes(execution: InstanceExecution): FlowNode[] {
  const coreNodes: FlowNode[] = [
    {
      id: `${execution.id}-trigger`,
      lane: 'core',
      title: 'Trigger',
      subtitle: `Disparo por ${execution.trigger_type}`,
      status: execution.status === 'queued' ? 'pending' : 'success',
      icon: GitBranch,
      timestamp: execution.started_at,
      meta: [
        { label: 'Tipo', value: execution.trigger_type },
        { label: 'Inicio', value: formatDate(execution.started_at) },
      ],
    },
    {
      id: `${execution.id}-input`,
      lane: 'core',
      title: 'Entrada Resolvida',
      subtitle: execution.input_payload ? 'Payload usado como entrada do script' : 'Sem payload de entrada persistido',
      status: execution.input_payload ? 'success' : 'pending',
      icon: FileJson2,
      payload: execution.input_payload,
      meta: [
        {
          label: 'Campos',
          value:
            execution.input_payload && typeof execution.input_payload === 'object'
              ? String(Object.keys(execution.input_payload).length)
              : '0',
        },
      ],
    },
    {
      id: `${execution.id}-script`,
      lane: 'core',
      title: 'Execucao do Script',
      subtitle: execution.error_message ? 'Script terminou com erro' : 'Script executado pelo Sparkit',
      status: nodeStatusFromExecution(execution.status),
      icon: execution.status === 'running' ? Loader2 : PlayCircle,
      error: execution.error_message,
      meta: [
        { label: 'Status', value: executionStatusLabel[execution.status] || execution.status },
        { label: 'Duracao', value: execution.duration_ms ? `${execution.duration_ms} ms` : '--' },
      ],
    },
    {
      id: `${execution.id}-output`,
      lane: 'core',
      title: 'Saida Estruturada',
      subtitle: execution.output_payload ? 'Objeto normalizado pronto para destinos' : 'Saida estruturada indisponivel',
      status: execution.output_payload ? 'success' : execution.status === 'failed' ? 'failed' : 'pending',
      icon: Activity,
      payload: execution.output_payload ?? execution.output,
      meta: [
        {
          label: 'Campos',
          value:
            execution.output_payload && typeof execution.output_payload === 'object'
              ? String(Object.keys(execution.output_payload).length)
              : '0',
        },
      ],
    },
  ];

  const destinationNodes: FlowNode[] = (execution.destination_details || []).map((detail, index) => {
    const labelParts = [detail.server_name, detail.resource_name, detail.operation_name].filter(Boolean);
    const title = labelParts.length > 0 ? labelParts.join(' / ') : `Destino ${detail.destination_id.slice(0, 8)}`;

    return {
      id: `${execution.id}-destination-${index}`,
      lane: 'destination',
      title,
      subtitle: destinationStatusLabel[detail.status] || detail.status,
      status: nodeStatusFromDestination(detail.status, detail.used_fallback),
      icon: detail.used_fallback ? ShieldAlert : RadioTower,
      payload: detail.payload,
      error: detail.error,
      timestamp: detail.timestamp,
      meta: [
        { label: 'Operacao', value: detail.operation_name || detail.resource_operation_id || '--' },
        { label: 'Servidor', value: detail.server_name || '--' },
        { label: 'Timestamp', value: formatDate(detail.timestamp) },
      ],
    };
  });

  const logNodes: FlowNode[] = execution.logs && execution.logs.length > 0
    ? [
        {
          id: `${execution.id}-logs`,
          lane: 'core',
          title: 'Logs Tecnicos',
          subtitle: `${execution.logs.length} eventos de execucao capturados`,
          status: inferLogNodeStatus(execution.logs),
          icon: TerminalSquare,
          meta: summarizeLogs(execution.logs),
        },
      ]
    : [];

  const completionNode: FlowNode = {
    id: `${execution.id}-completion`,
    lane: 'core',
    title: 'Conclusao',
    subtitle: execution.fallback_used
      ? 'Execucao concluida com fallback'
      : execution.destination_sent
        ? 'Execucao concluida com envio aos destinos'
        : 'Execucao concluida sem envio',
    status:
      execution.status === 'failed'
        ? 'failed'
        : execution.fallback_used
          ? 'warning'
          : execution.status === 'running'
            ? 'running'
            : execution.status === 'timeout'
              ? 'warning'
              : 'success',
    icon: execution.status === 'failed' ? XCircle : execution.status === 'timeout' ? Timer : CheckCircle2,
    timestamp: execution.finished_at,
    meta: [
      { label: 'Finalizado', value: formatDate(execution.finished_at) },
      { label: 'Destino enviado', value: execution.destination_sent ? 'Sim' : 'Nao' },
      { label: 'Fallback', value: execution.fallback_used ? 'Sim' : 'Nao' },
    ],
  };

  return [...coreNodes, ...destinationNodes, ...logNodes, completionNode];
}

export function ExecutionFlow({ execution }: { execution: InstanceExecution }) {
  const nodes = useMemo(() => buildNodes(execution), [execution]);
  const coreNodes = useMemo(() => nodes.filter((node) => node.lane === 'core'), [nodes]);
  const destinationNodes = useMemo(() => nodes.filter((node) => node.lane === 'destination'), [nodes]);
  const [selectedNodeId, setSelectedNodeId] = useState<string>(coreNodes[0]?.id ?? '');

  useEffect(() => {
    if (!nodes.some((node) => node.id === selectedNodeId)) {
      setSelectedNodeId(coreNodes[0]?.id ?? destinationNodes[0]?.id ?? '');
    }
  }, [coreNodes, destinationNodes, nodes, selectedNodeId]);

  const selectedNode = nodes.find((node) => node.id === selectedNodeId) ?? nodes[0];
  const SelectedIcon = selectedNode?.icon;

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
        <div className="mb-4">
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-zinc-500">Pipeline da Execucao</p>
          <p className="mt-1 text-sm text-zinc-400">
            Nos conectados com entrada, saida, destinos e logs tecnicos da execucao.
          </p>
        </div>

        <div className="overflow-x-auto pb-2">
          <div className="flex min-w-max items-start gap-3">
            {coreNodes.map((node, index) => {
              const style = nodeStatusStyles[node.status];
              const NodeIcon = node.icon;
              const isSelected = selectedNodeId === node.id;

              return (
                <div key={node.id} className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={() => setSelectedNodeId(node.id)}
                    className={`group relative w-60 rounded-2xl border p-4 text-left transition-all ${style.card} ${
                      isSelected ? 'ring-1 ring-white/25 shadow-[0_18px_50px_rgba(0,0,0,0.28)]' : 'hover:-translate-y-0.5 hover:border-white/[0.14]'
                    }`}
                  >
                    <div className="mb-3 flex items-center justify-between gap-3">
                      <div className="flex items-center gap-2">
                        <div className={`h-2.5 w-2.5 rounded-full ${style.dot}`} />
                        <span className={`rounded-full px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.18em] ${style.badge}`}>
                          {node.status}
                        </span>
                      </div>
                      <NodeIcon size={16} className={`${style.text} ${node.status === 'running' ? 'animate-spin' : ''}`} />
                    </div>

                    <p className={`text-sm font-semibold ${style.text}`}>{node.title}</p>
                    <p className="mt-1 min-h-10 text-xs leading-relaxed text-zinc-400">{node.subtitle}</p>

                    {node.timestamp && (
                      <div className="mt-3 flex items-center gap-2 text-[11px] text-zinc-500">
                        <Clock3 size={12} />
                        <span>{formatDate(node.timestamp)}</span>
                      </div>
                    )}
                  </button>

                  {index < coreNodes.length - 1 && (
                    <div className="flex min-w-16 items-center justify-center">
                      <div className="h-px w-16 bg-gradient-to-r from-white/[0.2] via-white/[0.45] to-white/[0.12]" />
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {destinationNodes.length > 0 && (
          <div className="mt-6 rounded-2xl border border-white/[0.06] bg-black/10 p-4">
            <div className="mb-4 flex items-center gap-2">
              <RadioTower size={15} className="text-zinc-400" />
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-zinc-500">Ramificacoes de Destino</p>
            </div>

            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              {destinationNodes.map((node) => {
                const style = nodeStatusStyles[node.status];
                const NodeIcon = node.icon;
                const isSelected = selectedNodeId === node.id;

                return (
                  <button
                    key={node.id}
                    type="button"
                    onClick={() => setSelectedNodeId(node.id)}
                    className={`relative rounded-2xl border p-4 text-left transition-all ${style.card} ${
                      isSelected ? 'ring-1 ring-white/25 shadow-[0_18px_50px_rgba(0,0,0,0.28)]' : 'hover:-translate-y-0.5 hover:border-white/[0.14]'
                    }`}
                  >
                    <div className="mb-3 flex items-center justify-between gap-3">
                      <div className="flex items-center gap-2">
                        <div className={`h-2.5 w-2.5 rounded-full ${style.dot}`} />
                        <span className={`rounded-full px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.18em] ${style.badge}`}>
                          {node.status}
                        </span>
                      </div>
                      <NodeIcon size={16} className={style.text} />
                    </div>

                    <p className={`text-sm font-semibold ${style.text}`}>{node.title}</p>
                    <p className="mt-1 text-xs leading-relaxed text-zinc-400">{node.subtitle}</p>
                  </button>
                );
              })}
            </div>
          </div>
        )}
      </div>

      {selectedNode && (
        <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.15fr_0.85fr]">
          <div className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
            <div className="mb-3 flex items-center gap-2">
              {SelectedIcon ? (
                <SelectedIcon
                  size={16}
                  className={`${nodeStatusStyles[selectedNode.status].text} ${selectedNode.status === 'running' ? 'animate-spin' : ''}`}
                />
              ) : null}
              <div>
                <p className="text-sm font-semibold text-white">{selectedNode.title}</p>
                <p className="text-xs text-zinc-500">{selectedNode.subtitle}</p>
              </div>
            </div>

            {selectedNode.payload !== undefined && (
              <div className="rounded-xl border border-white/[0.08] bg-black/20 p-3">
                <p className="mb-2 text-[10px] uppercase tracking-[0.2em] text-zinc-500">Payload</p>
                <pre className="max-h-[28rem] overflow-auto whitespace-pre-wrap font-mono text-xs leading-relaxed text-zinc-300">
                  {prettyJson(selectedNode.payload)}
                </pre>
              </div>
            )}

            {selectedNode.error && (
              <div className="mt-3 rounded-xl border border-red-500/20 bg-red-500/8 p-3">
                <div className="mb-2 flex items-center gap-2">
                  <AlertTriangle size={14} className="text-red-300" />
                  <p className="text-[10px] font-semibold uppercase tracking-[0.2em] text-red-300">Erro</p>
                </div>
                <pre className="whitespace-pre-wrap font-mono text-xs leading-relaxed text-red-200">{selectedNode.error}</pre>
              </div>
            )}

            {selectedNode.payload === undefined && !selectedNode.error && (
              <div className="rounded-xl border border-dashed border-white/[0.08] bg-white/[0.02] p-6 text-center text-sm text-zinc-500">
                Este no nao possui payload detalhado salvo. Selecione outro passo para inspecionar os dados.
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
              <p className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-zinc-500">Metadados do No</p>
              <div className="space-y-2">
                {(selectedNode.meta || []).map((item) => (
                  <div key={`${selectedNode.id}-${item.label}`} className="rounded-lg bg-white/[0.03] px-3 py-2">
                    <p className="text-[10px] uppercase tracking-[0.16em] text-zinc-500">{item.label}</p>
                    <p className="mt-1 break-words text-sm text-zinc-200">{item.value}</p>
                  </div>
                ))}
                {!selectedNode.meta?.length && (
                  <div className="rounded-lg bg-white/[0.03] px-3 py-4 text-sm text-zinc-500">
                    Nenhum metadado adicional disponivel.
                  </div>
                )}
              </div>
            </div>

            <div className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
              <p className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-zinc-500">Resumo da Execucao</p>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div className="rounded-lg bg-white/[0.03] p-3">
                  <p className="text-[10px] uppercase tracking-[0.16em] text-zinc-500">Status</p>
                  <p className="mt-1 text-zinc-200">{executionStatusLabel[execution.status] || execution.status}</p>
                </div>
                <div className="rounded-lg bg-white/[0.03] p-3">
                  <p className="text-[10px] uppercase tracking-[0.16em] text-zinc-500">Duracao</p>
                  <p className="mt-1 text-zinc-200">
                    {execution.duration_ms
                      ? execution.duration_ms >= 1000
                        ? `${(execution.duration_ms / 1000).toFixed(1)}s`
                        : `${execution.duration_ms}ms`
                      : '--'}
                  </p>
                </div>
                <div className="rounded-lg bg-white/[0.03] p-3">
                  <p className="text-[10px] uppercase tracking-[0.16em] text-zinc-500">Fallback</p>
                  <p className="mt-1 text-zinc-200">{execution.fallback_used ? 'Ativado' : 'Nao usado'}</p>
                </div>
                <div className="rounded-lg bg-white/[0.03] p-3">
                  <p className="text-[10px] uppercase tracking-[0.16em] text-zinc-500">Destinos</p>
                  <p className="mt-1 text-zinc-200">{execution.destination_details?.length || 0}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
