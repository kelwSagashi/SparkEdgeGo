import { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { AlertCircle, ArrowLeft, Clock, Loader2, RefreshCw } from 'lucide-react';
import { api } from '@/server/server.service';
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
        api.listInstanceExecutions(id),
      ]);

      if (resInstance.data?.data) {
        setInstance(resInstance.data.data.instance);
      }

      setExecutions(resExecutions.data?.data || []);
    } catch (e: any) {
      console.error('Failed to load instance logs', e);
      setError(e.message || 'Erro ao carregar logs da instancia.');
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
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="animate-spin text-zinc-500" size={32} />
      </div>
    );
  }

  return (
    <main className="mx-auto w-full max-w-[1200px] grow px-8 py-6 pb-20">
      <div className="mb-8 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={handleBack}
            className="rounded-lg p-2 text-zinc-400 transition-colors hover:bg-white/5 hover:text-white"
          >
            <ArrowLeft size={20} />
          </button>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-white">
              {instance?.name || 'Carregando...'}
            </h1>
            <p className="mt-1 text-sm text-zinc-500">
              Logs e historico de execucoes da instancia.
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
            Editar Instancia
          </Button>
        </div>
      </div>

      {error && (
        <Card className="mb-6 flex gap-3 border-destructive/20 bg-destructive/10 p-4 text-destructive">
          <AlertCircle size={20} />
          <p className="text-sm">{error}</p>
        </Card>
      )}

      <div className="mb-8 grid grid-cols-1 gap-4 md:grid-cols-4">
        {[
          { label: 'Total Execucoes', value: executions.length },
          { label: 'Sucesso', value: executions.filter((item) => item.status === 'success').length, color: 'text-emerald-400' },
          { label: 'Falha', value: executions.filter((item) => item.status === 'failed').length, color: 'text-red-400' },
          { label: 'Status Atual', value: instance?.status || '--', capitalize: true },
        ].map((stat) => (
          <div key={stat.label} className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
            <p className="mb-1 text-[10px] uppercase tracking-wider text-zinc-500">{stat.label}</p>
            <p className={`text-xl font-semibold ${stat.color || 'text-white'} ${stat.capitalize ? 'capitalize' : ''}`}>
              {stat.value}
            </p>
          </div>
        ))}
      </div>

      <div className="overflow-hidden rounded-xl border border-white/[0.08] bg-white/[0.02]">
        <div className="border-b border-white/[0.08] bg-white/[0.01] px-5 py-4">
          <h3 className="text-sm font-medium text-white">Ultimas Execucoes</h3>
        </div>

        {executions.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center text-zinc-500">
            <Clock size={40} className="mb-3 opacity-20" />
            <p className="text-sm">Nenhuma execucao registrada para esta instancia.</p>
          </div>
        ) : (
          <div className="divide-y divide-white/[0.04]">
            {executions.map((execution) => (
              <ExecutionRow key={execution.id} execution={execution} />
            ))}
          </div>
        )}
      </div>
    </main>
  );
}
