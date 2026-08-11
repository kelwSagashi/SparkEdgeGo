import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { AlertCircle, ArrowLeft, Clock, Loader2, RefreshCw, Send, ShieldAlert } from 'lucide-react';
import { api } from '@/server/server.service';
import { ExecutionRow, type InstanceExecution } from '@/components/executions/execution-row';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

export default function InstanceLogsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [executions, setExecutions] = useState<InstanceExecution[]>([]);
  const [instance, setInstance] = useState<any>(null);
  const [destinations, setDestinations] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dispatchPayload, setDispatchPayload] = useState('{\n  "value": 42\n}');
  const [eventName, setEventName] = useState('');
  const [dispatching, setDispatching] = useState(false);
  const [dispatchMessage, setDispatchMessage] = useState<string | null>(null);

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
        setDestinations(resInstance.data.data.destinations || []);
        setEventName(
          resInstance.data.data.instance?.trigger_config?.event_name || ''
        );
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

  const breakerCards = useMemo(
    () =>
      destinations
        .filter((item) => item?.breaker_state)
        .map((item) => ({
          id: item.destination?.id,
          operation: item.destination?.resource_operation_id,
          failures: item.breaker_state?.consecutive_failures ?? 0,
          openedUntil: item.breaker_state?.opened_until,
        })),
    [destinations],
  );

  const handleDispatch = useCallback(async () => {
    if (!instance) return;
    try {
      setDispatching(true);
      setDispatchMessage(null);
      const parsedPayload = dispatchPayload.trim() ? JSON.parse(dispatchPayload) : {};
      if (instance.trigger_type === 'event') {
        const response = await api.dispatchEvent(eventName, parsedPayload);
        const matched = (response.data?.data || []).find((item: any) => item.instance_id === instance.id);
        setDispatchMessage(
          matched
            ? `Evento disparado para a instancia. Status: ${matched.status}`
            : 'Evento enviado, mas nenhuma execucao dessa instancia foi retornada.',
        );
      } else if (instance.trigger_type === 'state_change') {
        const response = await api.dispatchStateChange(parsedPayload);
        const matched = (response.data?.data || []).find((item: any) => item.instance_id === instance.id);
        setDispatchMessage(
          matched
            ? `Mudanca de estado processada. Status: ${matched.status}`
            : 'Payload enviado, mas nenhuma execucao dessa instancia foi retornada.',
        );
      }
      await loadData();
    } catch (dispatchError: any) {
      setDispatchMessage(dispatchError?.message || 'Falha ao disparar o gatilho.');
    } finally {
      setDispatching(false);
    }
  }, [dispatchPayload, eventName, instance, loadData]);

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

      {(instance?.trigger_type === 'event' || instance?.trigger_type === 'state_change') && (
        <Card className="mb-6 border-white/[0.08] bg-white/[0.02] p-5">
          <div className="mb-4 flex items-center gap-3">
            <div className="rounded-lg bg-white/[0.05] p-2 text-zinc-300">
              <Send size={16} />
            </div>
            <div>
              <h3 className="text-sm font-medium text-white">Disparo Assistido</h3>
              <p className="text-xs text-zinc-500">
                Teste manual do gatilho {instance.trigger_type} diretamente pela UI.
              </p>
            </div>
          </div>

          {instance?.trigger_type === 'event' && (
            <div className="mb-3">
              <label className="mb-1 block text-xs uppercase tracking-wider text-zinc-500">
                Nome do evento
              </label>
              <input
                value={eventName}
                onChange={(e) => setEventName(e.target.value)}
                className="w-full rounded-lg border border-white/[0.1] bg-white/[0.03] px-3 py-2 text-sm text-white outline-none focus:border-white/[0.2]"
                placeholder="sensor.updated"
              />
            </div>
          )}

          <div className="mb-3">
            <label className="mb-1 block text-xs uppercase tracking-wider text-zinc-500">
              Payload JSON
            </label>
            <Textarea
              value={dispatchPayload}
              onChange={(e) => setDispatchPayload(e.target.value)}
              className="min-h-32 font-mono text-xs text-white"
            />
          </div>

          <div className="flex items-center gap-3">
            <Button onClick={handleDispatch} disabled={dispatching} className="gap-2">
              {dispatching ? <Loader2 size={14} className="animate-spin" /> : <Send size={14} />}
              Disparar
            </Button>
            {dispatchMessage && <p className="text-xs text-zinc-400">{dispatchMessage}</p>}
          </div>
        </Card>
      )}

      <div className="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
        {breakerCards.length === 0 ? (
          <div className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4 md:col-span-3">
            <p className="text-xs uppercase tracking-wider text-zinc-500">Circuit Breaker</p>
            <p className="mt-1 text-sm text-zinc-400">Nenhum destino com estado persistido de circuit breaker no momento.</p>
          </div>
        ) : (
          breakerCards.map((breaker) => (
            <div key={breaker.id} className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-4">
              <div className="mb-2 flex items-center gap-2">
                <ShieldAlert size={14} className={breaker.openedUntil ? 'text-amber-400' : 'text-zinc-400'} />
                <p className="text-xs uppercase tracking-wider text-zinc-500">Destino {breaker.id?.slice(0, 8)}</p>
              </div>
              <p className="text-sm font-medium text-white">{breaker.operation || 'Sem operacao'}</p>
              <p className="mt-2 text-xs text-zinc-400">Falhas consecutivas: {breaker.failures}</p>
              <p className="text-xs text-zinc-400">
                {breaker.openedUntil
                  ? `Aberto ate ${new Date(breaker.openedUntil).toLocaleString('pt-BR')}`
                  : 'Circuito fechado'}
              </p>
            </div>
          ))
        )}
      </div>

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
