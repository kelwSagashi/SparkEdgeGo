import { useEffect, useState } from 'react';
import {
  FileText, HardDrive, RefreshCw, Clock, CheckCircle2,
  XCircle, AlertTriangle, Trash2, Search,
  Table as TableIcon, Database, ArrowRight, ShieldAlert, Activity, Layers3
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion } from 'framer-motion';
import { toast } from 'sonner';
import { useFallbackStore } from '@/stores/fallback-store';

const statusConfig: Record<string, { icon: React.ElementType; color: string; bg: string; label: string }> = {
  pending: { icon: Clock,         color: 'text-amber-400',   bg: 'bg-amber-500/10',   label: 'Pendente' },
  sending: { icon: RefreshCw,     color: 'text-blue-400',    bg: 'bg-blue-500/10',    label: 'Enviando' },
  sent:    { icon: CheckCircle2,  color: 'text-emerald-400', bg: 'bg-emerald-500/10', label: 'Enviado' },
  failed:  { icon: XCircle,       color: 'text-red-400',     bg: 'bg-red-500/10',     label: 'Falhado' },
};

function severityFromPercent(value: number) {
  if (value >= 85) {
    return {
      label: 'critical',
      card: 'border-red-500/30 bg-red-500/[0.06]',
      text: 'text-red-300',
      badge: 'bg-red-500/15 text-red-200 border border-red-500/30',
    };
  }
  if (value >= 60) {
    return {
      label: 'warning',
      card: 'border-amber-500/30 bg-amber-500/[0.05]',
      text: 'text-amber-200',
      badge: 'bg-amber-500/15 text-amber-100 border border-amber-500/30',
    };
  }
  return {
    label: 'normal',
    card: 'border-emerald-500/20 bg-emerald-500/[0.04]',
    text: 'text-emerald-200',
    badge: 'bg-emerald-500/15 text-emerald-100 border border-emerald-500/20',
  };
}

function severityFromAgeSeconds(value: number) {
  if (value >= 1800) {
    return {
      label: 'critical',
      card: 'border-red-500/30 bg-red-500/[0.06]',
      text: 'text-red-300',
      badge: 'bg-red-500/15 text-red-200 border border-red-500/30',
    };
  }
  if (value >= 600) {
    return {
      label: 'warning',
      card: 'border-amber-500/30 bg-amber-500/[0.05]',
      text: 'text-amber-200',
      badge: 'bg-amber-500/15 text-amber-100 border border-amber-500/30',
    };
  }
  return {
    label: 'normal',
    card: 'border-emerald-500/20 bg-emerald-500/[0.04]',
    text: 'text-emerald-200',
    badge: 'bg-emerald-500/15 text-emerald-100 border border-emerald-500/20',
  };
}

function formatDuration(seconds: number) {
  if (!seconds || seconds <= 0) return '-';
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return minutes > 0 ? `${hours}h ${minutes}min` : `${hours}h`;
}

export default function LocalDataPage() {
  const { items, stats: fallbackStats, loading, refreshing, fetchItems, flushQueue, retryItem, deleteItem } = useFallbackStore();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string | 'all'>('all');

  useEffect(() => {
    fetchItems();
  }, []);

  const handleFlush = async () => {
    const sent = await flushQueue();
    if (sent > 0) {
      toast.success(`${sent} itens reenviados com sucesso`);
    } else {
      toast.info('Nenhum item pendente para reenvio');
    }
  };

  const handleRetry = async (id: string) => {
    const success = await retryItem(id);
    if (success) {
      toast.success('Item reenviado com sucesso');
    } else {
      toast.error('Falha ao reenviar item');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Tem certeza que deseja excluir este item?')) return;
    await deleteItem(id);
    toast.success('Item excluído');
  };

  const filtered = items.filter(i => {
    const term = search.toLowerCase();
    const matchesSearch = !search || 
      i.instance_id.toLowerCase().includes(term) || 
      i.id.toLowerCase().includes(term) ||
      (i.execution_id?.toLowerCase().includes(term) ?? false);
      
    const matchesStatus = statusFilter === 'all' || i.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const stats = {
    total: items.length,
    pending: items.filter(i => i.status === 'pending').length,
    sent: items.filter(i => i.status === 'sent').length,
    failed: items.filter(i => i.status === 'failed').length,
  };

  const sentRetentionHours = fallbackStats?.retention?.sent_retention_hours ?? 0;
  const failedRetentionHours = fallbackStats?.retention?.failed_retention_hours ?? 0;
  const keepSentItems = fallbackStats?.retention?.keep_sent_items ?? 0;
  const keepFailedItems = fallbackStats?.retention?.keep_failed_items ?? 0;
  const sentUsagePct = fallbackStats?.usage?.sent_pct_of_sent_window ?? 0;
  const failedUsagePct = fallbackStats?.usage?.failed_pct_of_failed_window ?? 0;
  const oldestPendingAgeSeconds = fallbackStats?.oldest_pending_age_seconds ?? 0;
  const historyPressurePct = Math.max(sentUsagePct, failedUsagePct);
  const historyPressureTone = severityFromPercent(historyPressurePct);
  const pendingAgeTone = severityFromAgeSeconds(oldestPendingAgeSeconds);
  const retryPressurePct = stats.total > 0 ? Math.round((items.filter((item) => item.retry_count > 0).length / stats.total) * 100) : 0;
  const retryPressureTone = severityFromPercent(retryPressurePct);
  const unresolvedPct = stats.total > 0 ? Math.round(((stats.pending + stats.failed) / stats.total) * 100) : 0;
  const unresolvedTone = severityFromPercent(unresolvedPct);
  const destinationCoverage = items.filter((item) => item.destination_id).length;
  const topPendingInstances = Array.from(
    items
      .filter((item) => item.status === 'pending' || item.status === 'failed')
      .reduce((map, item) => {
        map.set(item.instance_id, (map.get(item.instance_id) ?? 0) + 1);
        return map;
      }, new Map<string, number>()),
  )
    .sort((a, b) => b[1] - a[1])
    .slice(0, 3);

  const localGuidance = stats.failed > 0
    ? {
        title: 'Existem itens com falha terminal',
        description: 'Revise os erros mais recentes e priorize os destinos com tentativas esgotadas antes que o backlog cresca.',
        tone: 'border-red-500/30 bg-red-500/[0.06] text-red-100',
        icon: ShieldAlert,
      }
    : stats.pending > 0
      ? {
          title: 'Fila local aguardando reenvio',
          description: 'O Edge esta preservando dados offline. Vale monitorar a idade do item mais antigo e forcar um flush quando a conectividade estabilizar.',
          tone: 'border-amber-500/30 bg-amber-500/[0.05] text-amber-100',
          icon: Activity,
        }
      : {
          title: 'Fallback local sob controle',
          description: 'Nao ha backlog operacional pendente. O historico local esta funcionando mais como trilha de auditoria do que como fila de contingencia.',
          tone: 'border-emerald-500/20 bg-emerald-500/[0.05] text-emerald-100',
          icon: CheckCircle2,
        };
  const GuidanceIcon = localGuidance.icon;

  return (
    <main className="grow px-8 py-6 w-full max-w-[1200px] mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 border border-white/[0.08] flex items-center justify-center shadow-lg shadow-cyan-500/10">
            <HardDrive size={20} className="text-cyan-400" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold text-white tracking-tight">Dados Locais</h1>
            <p className="text-sm text-zinc-500">Traceabilidade e Gerenciamento de Fallback.</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="sm"
            className="text-zinc-400 hover:text-white"
            onClick={fetchItems}
            disabled={loading}
          >
            <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="gap-2 border-cyan-500/30 text-cyan-400 hover:bg-cyan-500/10 hover:border-cyan-500/50 transition-all font-medium"
            onClick={handleFlush}
            disabled={refreshing || stats.pending === 0}
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
            Reenviar todos os pendentes
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {[
          { label: 'Total Registrado', value: stats.total, color: 'text-zinc-100', bg: 'bg-zinc-500/10', Icon: HardDrive },
          { label: 'Aguardando Fila', value: stats.pending, color: 'text-amber-400', bg: 'bg-amber-500/10', Icon: Clock },
          { label: 'Recuperados', value: stats.sent, color: 'text-emerald-400', bg: 'bg-emerald-500/10', Icon: CheckCircle2 },
          { label: 'Tentativas Exaustas', value: stats.failed, color: 'text-red-400', bg: 'bg-red-500/10', Icon: XCircle },
        ].map(({ label, value, color, bg, Icon }) => (
          <div key={label} className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4 flex items-center gap-4 hover:bg-white/[0.04] transition-colors group">
            <div className={`w-11 h-11 rounded-xl ${bg} ${color} flex items-center justify-center group-hover:scale-110 transition-transform`}>
              <Icon size={20} />
            </div>
            <div>
              <p className="text-2xl font-bold text-white leading-none mb-1">{value}</p>
              <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">{label}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4">
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Retencao fallback</p>
          <p className="mt-2 text-sm text-white font-medium">
            Sent: {sentRetentionHours}h | Failed: {failedRetentionHours}h
          </p>
          <p className="mt-2 text-xs text-zinc-500">
            Historico terminal mantido localmente antes da limpeza automatica.
          </p>
        </div>
        <div className={`border rounded-2xl p-4 ${historyPressureTone.card}`}>
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Ocupacao historico</p>
          <div className="mt-2 flex items-center justify-between gap-3">
            <p className={`text-sm font-medium ${historyPressureTone.text}`}>
              {historyPressurePct}% de pressao
            </p>
            <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${historyPressureTone.badge}`}>
              {historyPressureTone.label}
            </span>
          </div>
          <p className="mt-2 text-sm text-white font-medium">
            Sent {sentUsagePct}% de {keepSentItems} | Failed {failedUsagePct}% de {keepFailedItems}
          </p>
          <p className="mt-2 text-xs text-zinc-500">
            Indica pressao sobre as janelas de retencao do armazenamento local.
          </p>
        </div>
        <div className={`border rounded-2xl p-4 ${pendingAgeTone.card}`}>
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Fila pendente mais antiga</p>
          <div className="mt-2 flex items-center justify-between gap-3">
            <p className={`text-sm font-medium ${pendingAgeTone.text}`}>
              {oldestPendingAgeSeconds > 0 ? `${oldestPendingAgeSeconds}s` : '-'}
            </p>
            <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${pendingAgeTone.badge}`}>
              {pendingAgeTone.label}
            </span>
          </div>
          <p className="mt-2 text-xs text-zinc-500">
            Quanto mais alto esse valor, maior o risco de backlog operacional.
          </p>
        </div>
      </div>

      <div className={`mb-8 rounded-2xl border px-5 py-4 ${localGuidance.tone}`}>
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <GuidanceIcon size={16} />
              <p className="text-sm font-semibold">{localGuidance.title}</p>
            </div>
            <p className="text-xs leading-relaxed opacity-90">{localGuidance.description}</p>
          </div>
          <div className="text-right">
            <p className="text-[10px] uppercase tracking-widest opacity-70">Idade do backlog</p>
            <p className="mt-1 text-sm font-semibold">{formatDuration(oldestPendingAgeSeconds)}</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <div className={`border rounded-2xl p-4 ${retryPressureTone.card}`}>
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Pressao de retries</p>
          <div className="mt-2 flex items-center justify-between gap-3">
            <p className={`text-sm font-medium ${retryPressureTone.text}`}>{retryPressurePct}%</p>
            <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${retryPressureTone.badge}`}>
              {retryPressureTone.label}
            </span>
          </div>
          <p className="mt-2 text-xs text-zinc-500">
            {items.filter((item) => item.retry_count > 0).length} item(ns) ja exigiram reenvio manual ou automatico.
          </p>
        </div>
        <div className={`border rounded-2xl p-4 ${unresolvedTone.card}`}>
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Itens nao resolvidos</p>
          <div className="mt-2 flex items-center justify-between gap-3">
            <p className={`text-sm font-medium ${unresolvedTone.text}`}>{unresolvedPct}%</p>
            <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${unresolvedTone.badge}`}>
              {unresolvedTone.label}
            </span>
          </div>
          <p className="mt-2 text-xs text-zinc-500">
            Pendentes + falhos sobre o total persistido localmente.
          </p>
        </div>
        <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4">
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Cobertura de destinos</p>
          <p className="mt-2 text-sm text-white font-medium">
            {destinationCoverage} / {items.length}
          </p>
          <p className="mt-2 text-xs text-zinc-500">
            Itens com `destination_id` conhecido, uteis para diagnostico fino do envio.
          </p>
        </div>
        <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4">
          <p className="text-[10px] text-zinc-500 uppercase tracking-widest font-semibold">Janela operacional</p>
          <p className="mt-2 text-sm text-white font-medium">
            Sent {sentRetentionHours}h | Failed {failedRetentionHours}h
          </p>
          <p className="mt-2 text-xs text-zinc-500">
            Tempo maximo visivel para auditoria local antes da limpeza automatica.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-[1.2fr,0.8fr] gap-4 mb-8">
        <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4">
          <div className="flex items-center gap-2 mb-3">
            <Layers3 size={15} className="text-cyan-300" />
            <p className="text-sm font-semibold text-white">Instancias com mais backlog local</p>
          </div>
          {topPendingInstances.length === 0 ? (
            <p className="text-sm text-zinc-500">Nenhuma instancia acumula backlog no momento.</p>
          ) : (
            <div className="space-y-2">
              {topPendingInstances.map(([instanceId, count]) => (
                <div key={instanceId} className="flex items-center justify-between rounded-xl border border-white/[0.05] bg-black/20 px-3 py-3">
                  <div>
                    <p className="text-sm text-white font-medium break-all">{instanceId}</p>
                    <p className="text-[11px] text-zinc-500">Itens pendentes ou falhos aguardando resolucao</p>
                  </div>
                  <span className="inline-flex items-center rounded-full px-2.5 py-1 text-[11px] font-bold uppercase tracking-widest bg-amber-500/10 text-amber-200 border border-amber-500/20">
                    {count}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-4">
          <p className="text-sm font-semibold text-white mb-3">Leitura rapida</p>
          <div className="space-y-3 text-xs text-zinc-400">
            <p>
              `pending` indica dados preservados localmente aguardando janela de conectividade.
            </p>
            <p>
              `sending` representa envio em progresso ou retentativa sendo processada.
            </p>
            <p>
              `failed` aponta itens que precisam de revisao porque a politica atual nao conseguiu drenar a fila.
            </p>
          </div>
        </div>
      </div>

      {/* Control Bar */}
      <div className="flex items-center gap-4 mb-6">
        <div className="relative flex-1">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-500" size={16} />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar por ID, Instância ou Execução..."
            className="w-full pl-11 pr-4 py-2.5 bg-white/[0.03] border border-white/[0.08] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-cyan-500/50 focus:bg-white/[0.05] transition-all"
          />
        </div>
        <div className="flex items-center gap-1 bg-white/[0.03] border border-white/[0.08] rounded-xl p-1 shadow-inner">
          {(['all', 'pending', 'sent', 'failed'] as const).map(s => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className={`px-4 py-2 rounded-lg text-xs font-semibold transition-all ${
                statusFilter === s 
                  ? 'bg-white/[0.08] text-white shadow-sm shadow-black/20' 
                  : 'text-zinc-500 hover:text-zinc-300'
              }`}
            >
              {s === 'all' ? 'Todos' : statusConfig[s].label}
            </button>
          ))}
        </div>
      </div>

      {/* Data Table */}
      {loading && items.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 bg-white/[0.01] border border-white/[0.04] rounded-2xl border-dashed">
          <div className="w-12 h-12 border-3 border-cyan-500/20 border-t-cyan-500 rounded-full animate-spin mb-4" />
          <p className="text-sm text-zinc-500 animate-pulse">Carregando armazenamento local...</p>
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center bg-white/[0.02] border border-white/[0.06] rounded-3xl border-dashed">
          <div className="w-20 h-20 rounded-3xl bg-white/[0.03] flex items-center justify-center mb-6 shadow-inner">
            <Database size={32} className="text-zinc-700" />
          </div>
          <h3 className="text-base font-medium text-white mb-2">Nenhum registro encontrado</h3>
          <p className="text-sm text-zinc-500 max-w-xs mx-auto">
            Os dados capturados que não puderem ser enviados aparecerão aqui automaticamente.
          </p>
        </div>
      ) : (
        <div className="overflow-hidden bg-white/[0.01] border border-white/[0.06] rounded-2xl shadow-2xl shadow-black/40">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-white/[0.06] bg-white/[0.02]">
                <th className="px-6 py-4 text-[11px] font-bold text-zinc-500 uppercase tracking-widest">Informações</th>
                <th className="px-6 py-4 text-[11px] font-bold text-zinc-500 uppercase tracking-widest">Estado</th>
                <th className="px-6 py-4 text-[11px] font-bold text-zinc-500 uppercase tracking-widest text-center">Tentativas</th>
                <th className="px-6 py-4 text-[11px] font-bold text-zinc-500 uppercase tracking-widest">Payload</th>
                <th className="px-6 py-4 text-[11px] font-bold text-zinc-500 uppercase tracking-widest text-right">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.04]">
              {filtered.map(item => {
                const cfg = statusConfig[item.status];
                const Icon = cfg.icon;
                return (
                  <motion.tr 
                    layout
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    key={item.id} 
                    className="group hover:bg-white/[0.03] transition-colors"
                  >
                    <td className="px-6 py-5">
                      <div className="flex flex-col gap-1.5">
                        <div className="flex items-center gap-2">
                           <span className="text-xs font-bold text-zinc-300 font-mono">ID: {item.id.slice(0, 8)}</span>
                           {item.execution_id && (
                             <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-blue-500/10 border border-blue-500/20 text-[9px] text-blue-400 font-bold uppercase">
                               Exec: {item.execution_id.slice(0, 6)}
                             </div>
                           )}
                        </div>
                        <div className="flex items-center gap-2 text-[11px] text-zinc-500">
                          <TableIcon size={12} className="text-zinc-600" />
                          <span className="truncate max-w-[150px]">Instância: {item.instance_id}</span>
                        </div>
                        <div className="flex items-center gap-2 text-[10px] text-zinc-600">
                          <Clock size={10} />
                          <span>{new Date(item.created_at).toLocaleString('pt-BR')}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="flex flex-col gap-2">
                         <div className={`flex items-center gap-2 px-2.5 py-1 rounded-full w-fit ${cfg.bg} ${cfg.color}`}>
                           <Icon size={12} className={item.status === 'sending' ? 'animate-spin' : ''} />
                           <span className="text-[10px] font-bold uppercase tracking-wider">{cfg.label}</span>
                         </div>
                         {item.last_error && (
                           <div className="flex items-center gap-1 text-[10px] text-red-400/80 italic max-w-[180px] truncate" title={item.last_error}>
                             <AlertTriangle size={10} />
                             {item.last_error}
                           </div>
                         )}
                      </div>
                    </td>
                    <td className="px-6 py-5 text-center">
                      <div className="flex flex-col items-center gap-1">
                        <span className="text-sm font-bold text-white">{item.retry_count}</span>
                        <span className="text-[9px] text-zinc-600 uppercase font-bold">Vezes</span>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                      <div className="relative group/payload">
                        <div className="w-10 h-10 rounded-lg bg-white/[0.05] border border-white/[0.08] flex items-center justify-center text-zinc-500 group-hover/payload:bg-white/[0.1] group-hover/payload:text-cyan-400 transition-all cursor-pointer">
                          <FileText size={18} />
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-5">
                       <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                         {item.status !== 'sent' && (
                           <Button
                             variant="ghost"
                             size="sm"
                             className="h-8 w-8 p-0 rounded-lg hover:bg-emerald-500/10 hover:text-emerald-400 text-zinc-500"
                             onClick={() => handleRetry(item.id)}
                             title="Tentar reenvio manual"
                           >
                             <ArrowRight size={16} />
                           </Button>
                         )}
                         <Button
                           variant="ghost"
                           size="sm"
                           className="h-8 w-8 p-0 rounded-lg hover:bg-red-500/10 hover:text-red-400 text-zinc-500"
                           onClick={() => handleDelete(item.id)}
                           title="Excluir item permanentemente"
                         >
                           <Trash2 size={16} />
                         </Button>
                       </div>
                    </td>
                  </motion.tr>
                );
              })}
            </tbody>
          </table>
          <div className="px-6 py-4 bg-white/[0.02] border-t border-white/[0.06] flex justify-between items-center">
            <p className="text-[11px] text-zinc-500 italic">
              * Itens com estratégia "Fila Ativa" são reenviados automaticamente ao detectar reconexão.
            </p>
            <p className="text-[11px] text-zinc-400 font-medium">
               Mostrando {filtered.length} de {items.length} registros
            </p>
          </div>
        </div>
      )}
    </main>
  );
}
