import { useEffect, useState } from 'react';
import { api } from '@/server/server.service';
import {
  Clock, Loader2, Search
} from 'lucide-react';
import { ExecutionRow, type InstanceExecution } from '@/components/executions/execution-row';

const statusLabels: Record<string, string> = {
  all: 'Todos',
  success: 'Sucesso',
  failed: 'Falha',
  running: 'Executando',
  queued: 'Na fila',
  timeout: 'Timeout'
};

export default function ExecutionHistoryPage() {
  const [executions, setExecutions] = useState<InstanceExecution[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<InstanceExecution['status'] | 'all'>('all');

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        // We'll use the executions list endpoint
        const res = await api.listExecutions();
        setExecutions(res.data?.data || []);
      } catch (e) {
        console.error('Failed to load executions', e);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const filtered = executions.filter(e => {
    const matchesSearch = !search || e.instance_id.includes(search) || e.id.includes(search);
    const matchesStatus = statusFilter === 'all' || e.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <main className="grow px-8 py-6 w-full max-w-[1200px] mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-white tracking-tight">Histórico de Execuções</h1>
        <p className="text-sm text-zinc-500 mt-1">Visualize todas as execuções de instâncias e seus resultados.</p>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={14} />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar por ID..."
            className="w-full pl-9 pr-3 py-2 bg-white/[0.04] border border-white/[0.1] rounded-lg text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] transition-colors"
          />
        </div>
        <div className="flex items-center gap-1.5 bg-white/[0.04] border border-white/[0.1] rounded-lg p-0.5">
          {(['all', 'success', 'failed', 'running', 'queued', 'timeout'] as const).map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                statusFilter === s ? 'bg-white/[0.1] text-white' : 'text-zinc-500 hover:text-zinc-300'
              }`}
            >
              {s === 'all' ? 'Todos' : statusLabels[s]}
            </button>
          ))}
        </div>
      </div>

      {/* List */}
      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-16 h-16 rounded-2xl bg-white/[0.04] border border-white/[0.08] flex items-center justify-center mb-4">
            <Clock size={24} className="text-zinc-600" />
          </div>
          <h3 className="text-sm font-medium text-white mb-1">Nenhuma execução encontrada</h3>
          <p className="text-xs text-zinc-500 max-w-sm">
            Sem execuções registradas. Execute uma instância para ver o histórico aqui.
          </p>
        </div>
      ) : (
        <div className="bg-white/[0.02] border border-white/[0.08] rounded-xl overflow-hidden">
          {filtered.map(exec => (
            <ExecutionRow key={exec.id} execution={exec} />
          ))}
        </div>
      )}
    </main>
  );
}

