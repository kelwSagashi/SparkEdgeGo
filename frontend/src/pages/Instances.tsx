import { useEffect, useState } from 'react';
import { useInstancesStore } from '@/stores/instances-store';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import {
  Plus, Play, Pause, Trash2, Settings, Activity, Clock,
  Search, Zap, ExternalLink
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import type { InstanceReturningValues } from "@/types/db";
import { cn } from '@/lib/utils';

const statusConfig: Record<InstanceReturningValues['status'], { label: string; color: string; dotColor: string }> = {
  idle:    { label: 'Idle',    color: 'text-zinc-400', dotColor: 'bg-zinc-400' },
  running: { label: 'Running', color: 'text-emerald-400', dotColor: 'bg-emerald-400' },
  paused:  { label: 'Paused',  color: 'text-amber-400', dotColor: 'bg-amber-400' },
  error:   { label: 'Error',   color: 'text-red-400', dotColor: 'bg-red-400' },
};

function StatusBadge({ status }: { status: InstanceReturningValues['status'] }) {
  const cfg = statusConfig[status];
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${cfg.color} bg-white/5 border border-white/10`}>
      <span className={`w-1.5 h-1.5 rounded-full ${cfg.dotColor} ${status === 'running' ? 'animate-pulse' : ''}`} />
      {cfg.label}
    </span>
  );
}

function InstanceCard({ instance, onTrigger, onDelete, onUpdateActive }: {
  instance: any;
  onTrigger: (id: string) => void;
  onDelete: (id: string) => void;
  onUpdateActive: (id: string, active: boolean) => void;
}) {
  const navigate = useNavigate();

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -12 }}
      transition={{ duration: 0.2 }}
      className="group relative flex flex-col bg-white/[0.03] hover:bg-white/[0.06] border border-white/[0.08] hover:border-white/[0.15] rounded-xl p-5 transition-all duration-300"
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1 min-w-0">
          <div 
            className="flex items-center gap-2 mb-2 cursor-pointer group/title"
            onClick={() => navigate(`/instances/${instance.id}`)}
          >
            <h3 className="text-sm font-medium text-white group-hover/title:text-emerald-400 transition-colors">
              {instance.name}
            </h3>
            <ExternalLink size={12} className="text-zinc-600 opacity-0 group-hover/title:opacity-100 transition-all" />
          </div>
          <p className="text-xs text-zinc-500 mt-0.5 truncate">
            {instance.description || 'Sem descrição'}
          </p>
        </div>
        <StatusBadge status={instance.status} />
      </div>

      {/* Info Row */}
      <div className="flex items-center gap-4 text-[11px] text-zinc-500">
        <div className="flex items-center gap-1">
          <Zap size={11} />
          <span className="capitalize">{instance.trigger_type}</span>
          {instance.trigger_type === 'interval' && instance.trigger_config?.interval_seconds && (
            <span className="text-zinc-600">({instance.trigger_config.interval_seconds}s)</span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <Clock size={11} />
          <span>{new Date(instance.created_at).toLocaleDateString('pt-BR')}</span>
        </div>
      </div>

      {/* Footer / Actions */}
      <div className="flex justify-between items-center pt-4 mt-auto border-t border-white/[0.06]">
        <div className="flex gap-1">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onUpdateActive(instance.id, !instance.active);
            }}
            className={cn(
              "p-2 rounded-lg transition-colors border",
              !instance.active
                ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20 hover:bg-emerald-500/20"
                : "bg-amber-500/10 text-amber-500 border-amber-500/20 hover:bg-amber-500/20"
            )}
            title={instance.active ? "Desativar (Pausar agendamento)" : "Ativar (Retomar agendamento)"}
          >
            {instance.active ? <Pause size={14} /> : <Play size={14} />}
          </button>
          
          <button
            onClick={(e) => {
              e.stopPropagation();
              onTrigger(instance.id);
            }}
            className="p-2 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 text-white transition-colors"
            title="Executar forçadamente agora"
          >
            <Zap size={14} />
          </button>
        </div>

        <div className="flex gap-1">
          <Button 
            size="sm" variant="ghost"
            className="h-8 text-xs text-secondary hover:text-primary"
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/instances/${instance.id}`)
            }}
          >
            Editar
          </Button>
          <Button 
            size="sm" variant="ghost"
            className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs"
            onClick={(e) => {
              e.stopPropagation();
              onDelete(instance.id);
            }}
          >
            Excluir
          </Button>
        </div>
      </div>

      {/* Running glow */}
      {instance.status === 'running' && (
        <div className="absolute inset-0 rounded-xl ring-1 ring-emerald-500/20 pointer-events-none" />
      )}
    </motion.div>
  );
}

export default function InstancesPage() {
  const { instances, loading, fetchAll, triggerInstance, deleteInstance, updateInstance } = useInstancesStore();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<InstanceReturningValues['status'] | 'all'>('all');

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const filtered = instances.filter(i => {
    const matchesSearch = !search ||
      i.name.toLowerCase().includes(search.toLowerCase()) ||
      i.tags?.some(t => t.toLowerCase().includes(search.toLowerCase()));
    const matchesStatus = statusFilter === 'all' || i.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const stats = {
    total: instances.length,
    running: instances.filter(i => i.status === 'running').length,
    idle: instances.filter(i => i.status === 'idle').length,
    error: instances.filter(i => i.status === 'error').length,
  };

  return (
    <main className="grow px-8 py-6 w-full max-w-[1400px] mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Instâncias</h1>
          <p className="text-sm text-zinc-500 mt-1">Gerencie e monitore suas instâncias de coleta de dados.</p>
        </div>
        <Button
          onClick={() => navigate('/instances/new')}
          className="bg-violet-600 hover:bg-violet-700 text-white gap-2"
        >
          Nova Instância
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        {[
          { label: 'Total', value: stats.total, icon: Activity, color: 'text-white' },
          { label: 'Running', value: stats.running, icon: Play, color: 'text-emerald-400' },
          { label: 'Idle', value: stats.idle, icon: Pause, color: 'text-zinc-400' },
          { label: 'Errors', value: stats.error, icon: Zap, color: 'text-red-400' },
        ].map((s) => (
          <div key={s.label} className="bg-white/[0.03] border border-white/[0.08] rounded-xl p-4 flex items-center gap-3">
            <div className={`p-2 rounded-lg bg-white/[0.05] ${s.color}`}>
              <s.icon size={16} />
            </div>
            <div>
              <p className="text-xl font-bold text-white">{s.value}</p>
              <p className="text-[11px] text-zinc-500 uppercase tracking-wider">{s.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={14} />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar instâncias..."
            className="w-full pl-9 pr-3 py-2 bg-white/[0.04] border border-white/[0.1] rounded-lg text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] transition-colors"
          />
        </div>
        <div className="flex items-center gap-1.5 bg-white/[0.04] border border-white/[0.1] rounded-lg p-0.5">
          {(['all', 'running', 'idle', 'paused', 'error'] as const).map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                statusFilter === s
                  ? 'bg-white/[0.1] text-white'
                  : 'text-zinc-500 hover:text-zinc-300'
              }`}
            >
              {s === 'all' ? 'Todos' : statusConfig[s].label}
            </button>
          ))}
        </div>
      </div>

      {/* Grid */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-16 h-16 rounded-2xl bg-white/[0.04] border border-white/[0.08] flex items-center justify-center mb-4">
            <Activity size={24} className="text-zinc-600" />
          </div>
          <h3 className="text-sm font-medium text-white mb-1">
            {search || statusFilter !== 'all' ? 'Nenhuma instância encontrada' : 'Nenhuma instância criada'}
          </h3>
          <p className="text-xs text-zinc-500 mb-4 max-w-sm">
            {search || statusFilter !== 'all'
              ? 'Tente ajustar os filtros de busca.'
              : 'Crie sua primeira instância para começar a coletar dados.'}
          </p>
          {!search && statusFilter === 'all' && (
            <Button
              onClick={() => navigate('/instances/new')}
              variant="outline"
              size="sm"
              className="gap-2 border-white/10 text-white hover:bg-white/5"
            >
              <Plus size={14} />
              Criar instância
            </Button>
          )}
        </div>
      ) : (
        <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          <AnimatePresence mode="popLayout">
            {filtered.map((instance) => (
              <InstanceCard
                key={instance.id}
                instance={instance}
                onTrigger={triggerInstance}
                onDelete={deleteInstance}
                onUpdateActive={(id, active) => updateInstance(id, { active })}
              />
            ))}
          </AnimatePresence>
        </motion.div>
      )}
    </main>
  );
}

