import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '@/server/server.service';
import {
  Clock, Loader2, ArrowLeft, RefreshCw, AlertCircle
} from 'lucide-react';
import { ExecutionRow, type InstanceExecution } from '@/components/executions/execution-row';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { cn } from '@/lib/utils';

export default function InstanceLogsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [executions, setExecutions] = useState<InstanceExecution[]>([]);
  const [instance, setInstance] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    if (!id) return;
    try {
      setLoading(true);
      setError(null);
      
      const [resInstance, resExecutions] = await Promise.all([
        api.getInstanceById(id),
        api.listInstanceExecutions(id)
      ]);

      console.log(resInstance)
      if (resInstance.data?.data) {
        setInstance(resInstance.data.data.instance);
      }
      
      setExecutions(resExecutions.data?.data || []);
    } catch (e: any) {
      console.error('Failed to load instance logs', e);
      setError(e.message || 'Erro ao carregar logs da instância.');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleBack = () => navigate('/instances');

  if (loading && !instance) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="animate-spin text-zinc-500" size={32} />
      </div>
    );
  }

  return (
    <main className="grow px-8 py-6 w-full max-w-[1200px] mx-auto pb-20">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <button
            onClick={handleBack}
            className="p-2 rounded-lg hover:bg-white/5 text-zinc-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={20} />
          </button>
          <div>
            <h1 className="text-2xl font-semibold text-white tracking-tight">
              {instance?.name || 'Carregando...'}
            </h1>
            <p className="text-sm text-zinc-500 mt-1">
              Logs e histórico de execuções da instância.
            </p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={loadData}
            disabled={loading}
            className="gap-2 text-secondary"
          >
            <RefreshCw size={14} className={cn(loading ? 'animate-spin' : '')} />
            Recarregar
          </Button>
          <Button
            variant="default"
            size="sm"
            onClick={() => navigate(`/instances/${id}/edit`)}
          >
            Editar Instância
          </Button>
        </div>
      </div>

      {error && (
        <Card className="p-4 bg-destructive/10 border-destructive/20 text-destructive flex gap-3 mb-6">
          <AlertCircle size={20} />
          <p className="text-sm">{error}</p>
        </Card>
      )}

      {/* Stats Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        {[
          { label: 'Total Execuções', value: executions.length },
          { label: 'Sucesso', value: executions.filter(e => e.status === 'success').length, color: 'text-emerald-400' },
          { label: 'Falha', value: executions.filter(e => e.status === 'failed').length, color: 'text-red-400' },
          { label: 'Status Atual', value: instance?.status || '—', capitalize: true },
        ].map(stat => (
          <div key={stat.label} className="bg-white/[0.02] border border-white/[0.08] rounded-xl p-4">
            <p className="text-[10px] text-zinc-500 uppercase tracking-wider mb-1">{stat.label}</p>
            <p className={`text-xl font-semibold ${stat.color || 'text-white'} ${stat.capitalize ? 'capitalize' : ''}`}>
              {stat.value}
            </p>
          </div>
        ))}
      </div>

      {/* List */}
      <div className="bg-white/[0.02] border border-white/[0.08] rounded-xl overflow-hidden">
        <div className="px-5 py-4 border-b border-white/[0.08] bg-white/[0.01]">
          <h3 className="text-sm font-medium text-white">Últimas Execuções</h3>
        </div>
        
        {executions.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center text-zinc-500">
            <Clock size={40} className="mb-3 opacity-20" />
            <p className="text-sm">Nenhuma execução registrada para esta instância.</p>
          </div>
        ) : (
          <div className="divide-y divide-white/[0.04]">
            {executions.map(exec => (
              <ExecutionRow key={exec.id} execution={exec} />
            ))}
          </div>
        )}
      </div>
    </main>
  );
}

