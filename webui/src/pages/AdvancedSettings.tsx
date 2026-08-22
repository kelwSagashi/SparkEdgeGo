import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { cloudService, type EdgeConfig } from '@/rest-api-client/cloud.service';
import {
  AlertTriangle,
  CheckCircle2,
  Cloud,
  Database,
  Info,
  Loader2,
  Lock,
  RefreshCw,
  Save,
  Server,
  Settings2,
} from 'lucide-react';
import { toast } from 'sonner';

const inputCls =
  'w-full px-4 py-3 bg-white/[0.04] border border-white/[0.1] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.25] focus:bg-white/[0.06] transition-all font-mono';
const labelCls = 'block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2';
const sectionCls = 'bg-white/[0.02] border border-white/[0.06] rounded-2xl p-6';

interface SectionHeaderProps {
  icon: React.ReactNode;
  title: string;
  description: string;
}

function SectionHeader({ icon, title, description }: SectionHeaderProps) {
  return (
    <div className="flex items-start gap-3 mb-5 pb-4 border-b border-white/[0.05]">
      <div className="w-9 h-9 rounded-xl bg-white/[0.04] flex items-center justify-center shrink-0 mt-0.5">
        {icon}
      </div>
      <div>
        <p className="text-sm font-semibold text-white">{title}</p>
        <p className="text-xs text-zinc-500 mt-0.5">{description}</p>
      </div>
    </div>
  );
}

export default function AdvancedSettingsPage() {
  const [config, setConfig] = useState<EdgeConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const [cloudUrl, setCloudUrl] = useState('');
  const [mqttUrl, setMqttUrl] = useState('');
  const [cloudSyncToken, setCloudSyncToken] = useState('');
  const [dbFile, setDbFile] = useState('');
  const [jwtSecret, setJwtSecret] = useState('');
  const [serverPort, setServerPort] = useState('');
  const [intermittentPendingAgeSeconds, setIntermittentPendingAgeSeconds] = useState('120');
  const [degradedPendingAgeSeconds, setDegradedPendingAgeSeconds] = useState('600');
  const [intermittentCloudSyncQueueDepth, setIntermittentCloudSyncQueueDepth] = useState('5');
  const [degradedCloudSyncQueueDepth, setDegradedCloudSyncQueueDepth] = useState('25');
  const [degradedMqttQueueDepth, setDegradedMqttQueueDepth] = useState('10');
  const [heartbeatHealthySeconds, setHeartbeatHealthySeconds] = useState('30');
  const [heartbeatDegradedSeconds, setHeartbeatDegradedSeconds] = useState('90');
  const [statsHealthySeconds, setStatsHealthySeconds] = useState('120');
  const [statsDegradedSeconds, setStatsDegradedSeconds] = useState('300');
  const [mqttQueueMaxItems, setMqttQueueMaxItems] = useState('1000');
  const [mqttQueueMaxAgeHours, setMqttQueueMaxAgeHours] = useState('336');
  const [cloudSyncSentRetentionHours, setCloudSyncSentRetentionHours] = useState('168');
  const [cloudSyncFailedRetentionHours, setCloudSyncFailedRetentionHours] = useState('720');
  const [cloudSyncKeepSentItems, setCloudSyncKeepSentItems] = useState('1000');
  const [cloudSyncKeepFailedItems, setCloudSyncKeepFailedItems] = useState('1000');
  const [localFallbackSentRetentionHours, setLocalFallbackSentRetentionHours] = useState('168');
  const [localFallbackFailedRetentionHours, setLocalFallbackFailedRetentionHours] = useState('720');
  const [localFallbackKeepSentItems, setLocalFallbackKeepSentItems] = useState('1000');
  const [localFallbackKeepFailedItems, setLocalFallbackKeepFailedItems] = useState('1000');

  const normalizePortValue = (value?: string | number | null) => String(value ?? '').replace(/^:/, '');

  const loadConfig = async () => {
    setLoading(true);
    try {
      const res = await cloudService.getConfig();
      const cfg = res.data;
      if (!cfg || !cfg.cloud || !cfg.db || !cfg.auth || !cfg.server || !cfg.connectivity || !cfg.retention) {
        throw new Error('Resposta de configuracao invalida.');
      }
      setConfig(cfg);
      setCloudUrl(cfg.cloud.url);
      setMqttUrl(cfg.cloud.mqtt_url);
      setCloudSyncToken('');
      setDbFile(cfg.db.file);
      setJwtSecret('');
      setServerPort(normalizePortValue(cfg.server.port));
      setIntermittentPendingAgeSeconds(String(cfg.connectivity.intermittent_pending_age_seconds));
      setDegradedPendingAgeSeconds(String(cfg.connectivity.degraded_pending_age_seconds));
      setIntermittentCloudSyncQueueDepth(String(cfg.connectivity.intermittent_cloud_sync_queue_depth));
      setDegradedCloudSyncQueueDepth(String(cfg.connectivity.degraded_cloud_sync_queue_depth));
      setDegradedMqttQueueDepth(String(cfg.connectivity.degraded_mqtt_queue_depth));
      setHeartbeatHealthySeconds(String(cfg.connectivity.heartbeat_healthy_seconds));
      setHeartbeatDegradedSeconds(String(cfg.connectivity.heartbeat_degraded_seconds));
      setStatsHealthySeconds(String(cfg.connectivity.stats_healthy_seconds));
      setStatsDegradedSeconds(String(cfg.connectivity.stats_degraded_seconds));
      setMqttQueueMaxItems(String(cfg.retention.mqtt_queue_max_items));
      setMqttQueueMaxAgeHours(String(cfg.retention.mqtt_queue_max_age_hours));
      setCloudSyncSentRetentionHours(String(cfg.retention.cloud_sync_sent_retention_hours));
      setCloudSyncFailedRetentionHours(String(cfg.retention.cloud_sync_failed_retention_hours));
      setCloudSyncKeepSentItems(String(cfg.retention.cloud_sync_keep_sent_items));
      setCloudSyncKeepFailedItems(String(cfg.retention.cloud_sync_keep_failed_items));
      setLocalFallbackSentRetentionHours(String(cfg.retention.local_fallback_sent_retention_hours));
      setLocalFallbackFailedRetentionHours(String(cfg.retention.local_fallback_failed_retention_hours));
      setLocalFallbackKeepSentItems(String(cfg.retention.local_fallback_keep_sent_items));
      setLocalFallbackKeepFailedItems(String(cfg.retention.local_fallback_keep_failed_items));
    } catch (err: any) {
      toast.error(`Nao foi possivel carregar as configuracoes: ${err?.message ?? 'Erro desconhecido'}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadConfig();
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setSaved(false);

    try {
      const updates: Record<string, any> = {};

      if (
        cloudUrl !== config?.cloud.url ||
        mqttUrl !== config?.cloud.mqtt_url ||
        cloudSyncToken.trim() !== ''
      ) {
        updates.cloud = {};
        if (cloudUrl !== config?.cloud.url) updates.cloud.url = cloudUrl;
        if (mqttUrl !== config?.cloud.mqtt_url) updates.cloud.mqtt_url = mqttUrl;
        if (cloudSyncToken.trim() !== '') updates.cloud.sync_token = cloudSyncToken.trim();
      }

      if (dbFile !== config?.db.file) {
        updates.db = { file: dbFile };
      }

      if (jwtSecret.trim() !== '') {
        updates.auth = { jwt_secret: jwtSecret };
      }

      const normalizedCurrentPort = normalizePortValue(config?.server.port);
      const normalizedNextPort = normalizePortValue(serverPort);
      if (normalizedNextPort !== normalizedCurrentPort) {
        updates.server = { port: normalizedNextPort };
      }

      const connectivityUpdates: Record<string, number> = {};
      if (Number(intermittentPendingAgeSeconds) !== config?.connectivity.intermittent_pending_age_seconds) {
        connectivityUpdates.intermittent_pending_age_seconds = Number(intermittentPendingAgeSeconds);
      }
      if (Number(degradedPendingAgeSeconds) !== config?.connectivity.degraded_pending_age_seconds) {
        connectivityUpdates.degraded_pending_age_seconds = Number(degradedPendingAgeSeconds);
      }
      if (Number(intermittentCloudSyncQueueDepth) !== config?.connectivity.intermittent_cloud_sync_queue_depth) {
        connectivityUpdates.intermittent_cloud_sync_queue_depth = Number(intermittentCloudSyncQueueDepth);
      }
      if (Number(degradedCloudSyncQueueDepth) !== config?.connectivity.degraded_cloud_sync_queue_depth) {
        connectivityUpdates.degraded_cloud_sync_queue_depth = Number(degradedCloudSyncQueueDepth);
      }
      if (Number(degradedMqttQueueDepth) !== config?.connectivity.degraded_mqtt_queue_depth) {
        connectivityUpdates.degraded_mqtt_queue_depth = Number(degradedMqttQueueDepth);
      }
      if (Number(heartbeatHealthySeconds) !== config?.connectivity.heartbeat_healthy_seconds) {
        connectivityUpdates.heartbeat_healthy_seconds = Number(heartbeatHealthySeconds);
      }
      if (Number(heartbeatDegradedSeconds) !== config?.connectivity.heartbeat_degraded_seconds) {
        connectivityUpdates.heartbeat_degraded_seconds = Number(heartbeatDegradedSeconds);
      }
      if (Number(statsHealthySeconds) !== config?.connectivity.stats_healthy_seconds) {
        connectivityUpdates.stats_healthy_seconds = Number(statsHealthySeconds);
      }
      if (Number(statsDegradedSeconds) !== config?.connectivity.stats_degraded_seconds) {
        connectivityUpdates.stats_degraded_seconds = Number(statsDegradedSeconds);
      }
      if (Object.keys(connectivityUpdates).length > 0) {
        updates.connectivity = connectivityUpdates;
      }

      const retentionUpdates: Record<string, number> = {};
      if (Number(mqttQueueMaxItems) !== config?.retention.mqtt_queue_max_items) {
        retentionUpdates.mqtt_queue_max_items = Number(mqttQueueMaxItems);
      }
      if (Number(mqttQueueMaxAgeHours) !== config?.retention.mqtt_queue_max_age_hours) {
        retentionUpdates.mqtt_queue_max_age_hours = Number(mqttQueueMaxAgeHours);
      }
      if (Number(cloudSyncSentRetentionHours) !== config?.retention.cloud_sync_sent_retention_hours) {
        retentionUpdates.cloud_sync_sent_retention_hours = Number(cloudSyncSentRetentionHours);
      }
      if (Number(cloudSyncFailedRetentionHours) !== config?.retention.cloud_sync_failed_retention_hours) {
        retentionUpdates.cloud_sync_failed_retention_hours = Number(cloudSyncFailedRetentionHours);
      }
      if (Number(cloudSyncKeepSentItems) !== config?.retention.cloud_sync_keep_sent_items) {
        retentionUpdates.cloud_sync_keep_sent_items = Number(cloudSyncKeepSentItems);
      }
      if (Number(cloudSyncKeepFailedItems) !== config?.retention.cloud_sync_keep_failed_items) {
        retentionUpdates.cloud_sync_keep_failed_items = Number(cloudSyncKeepFailedItems);
      }
      if (Number(localFallbackSentRetentionHours) !== config?.retention.local_fallback_sent_retention_hours) {
        retentionUpdates.local_fallback_sent_retention_hours = Number(localFallbackSentRetentionHours);
      }
      if (Number(localFallbackFailedRetentionHours) !== config?.retention.local_fallback_failed_retention_hours) {
        retentionUpdates.local_fallback_failed_retention_hours = Number(localFallbackFailedRetentionHours);
      }
      if (Number(localFallbackKeepSentItems) !== config?.retention.local_fallback_keep_sent_items) {
        retentionUpdates.local_fallback_keep_sent_items = Number(localFallbackKeepSentItems);
      }
      if (Number(localFallbackKeepFailedItems) !== config?.retention.local_fallback_keep_failed_items) {
        retentionUpdates.local_fallback_keep_failed_items = Number(localFallbackKeepFailedItems);
      }
      if (Object.keys(retentionUpdates).length > 0) {
        updates.retention = retentionUpdates;
      }

      if (Object.keys(updates).length === 0) {
        toast.info('Nenhuma alteracao detectada.');
        return;
      }

      const res = await cloudService.updateConfig(updates);
      toast.success(res.data.message || 'Configuracoes salvas. Reinicie o servico para aplicar as mudancas.');
      setSaved(true);
      setJwtSecret('');
      await loadConfig();
    } catch (err: any) {
      toast.error(`Falha ao salvar: ${err?.message ?? 'Erro desconhecido'}`);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <main className="grow px-8 py-6 w-full max-w-[700px] mx-auto">
        <div className="flex items-center gap-2 text-zinc-500 mt-24 justify-center">
          <Loader2 size={18} className="animate-spin" />
          <span className="text-sm">Carregando configuracoes...</span>
        </div>
      </main>
    );
  }

  return (
    <main className="grow px-8 py-6 w-full max-w-[700px] mx-auto pb-24 animate-in fade-in duration-300">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center shadow-lg shadow-amber-500/20">
            <Settings2 size={14} className="text-white" />
          </div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Configuracoes Avancadas</h1>
        </div>
        <p className="text-sm text-zinc-500 mt-1 ml-11">
          Configuracoes de infraestrutura do SparkEdge.
        </p>
      </div>

      <div className="flex items-start gap-3 bg-amber-500/[0.08] border border-amber-500/30 rounded-2xl px-5 py-4 mb-8 animate-in slide-in-from-top-2">
        <AlertTriangle size={18} className="text-amber-400 shrink-0 mt-0.5" />
        <div>
          <p className="text-sm font-semibold text-amber-300 mb-1">Atencao: Configuracoes de Infraestrutura</p>
          <p className="text-xs text-amber-400/80 leading-relaxed">
            Alterar essas configuracoes pode interromper a conexao com o Spark Cloud e o funcionamento do SparkEdge.
            Apos salvar, e necessario reiniciar o servico para aplicar as mudancas.
            As configuracoes sao salvas no arquivo <code className="bg-amber-500/10 px-1 rounded font-mono text-amber-300">config.yml</code>.
          </p>
        </div>
      </div>

      <div className="flex items-center gap-2 bg-white/[0.02] border border-white/[0.06] rounded-xl px-4 py-3 mb-6">
        <Info size={13} className="text-zinc-500 shrink-0" />
        <p className="text-xs text-zinc-500 font-mono flex-1 truncate">
          Arquivo: <span className="text-zinc-400">{config?.config_file ?? 'config.yml'}</span>
          <span className="mx-2 text-zinc-700">|</span>
          Prioridade: <span className="text-zinc-400">yml -&gt; .env -&gt; padroes</span>
        </p>
        <button
          onClick={() => void loadConfig()}
          className="text-zinc-600 hover:text-zinc-300 transition-colors ml-2 shrink-0"
          title="Recarregar"
        >
          <RefreshCw size={13} />
        </button>
      </div>

      <form onSubmit={handleSave} className="space-y-5">
        <div className={sectionCls}>
          <SectionHeader
            icon={<Cloud size={16} className="text-cyan-400" />}
            title="Cloud Integration"
            description="URLs de conexao com o Spark Cloud e broker MQTT"
          />
          <div className="space-y-4">
            <div>
              <label className={labelCls}>Spark Cloud URL</label>
              <input
                type="url"
                value={cloudUrl}
                onChange={(e) => setCloudUrl(e.target.value)}
                placeholder="https://spark-cloud.com"
                className={inputCls}
                required
              />
            </div>

            <div>
              <label className={labelCls}>MQTT Broker URL</label>
              <input
                type="text"
                value={mqttUrl}
                onChange={(e) => setMqttUrl(e.target.value)}
                placeholder="mqtt://localhost:1883"
                className={inputCls}
                required
              />
            </div>

            <div>
              <label className={labelCls}>Cloud Sync Token</label>
              <input
                type="password"
                value={cloudSyncToken}
                onChange={(e) => setCloudSyncToken(e.target.value)}
                placeholder={config?.cloud.sync_token ? `Atual: ${config.cloud.sync_token}` : 'Token para sincronizar eventos com o Spark Cloud'}
                className={inputCls}
                autoComplete="off"
              />
              <p className="mt-1.5 text-[10px] text-zinc-600">
                Deixe em branco para manter o token atual. Esse campo e opcional e funciona como sobreposicao manual para autenticar a fila HTTP de sincronizacao com o Spark Cloud.
              </p>
            </div>
          </div>
        </div>

        <div className={sectionCls}>
          <SectionHeader
            icon={<Database size={16} className="text-violet-400" />}
            title="Banco de Dados"
            description="Caminho do arquivo SQLite local"
          />
          <div>
            <label className={labelCls}>Caminho do arquivo DB</label>
            <input
              type="text"
              value={dbFile}
              onChange={(e) => setDbFile(e.target.value)}
              placeholder="sparkedge.db"
              className={inputCls}
              required
            />
          </div>
        </div>

        <div className={sectionCls}>
          <SectionHeader
            icon={<Lock size={16} className="text-rose-400" />}
            title="Autenticacao"
            description="Chave secreta para assinar tokens JWT locais"
          />
          <div>
            <label className={labelCls}>
              JWT Secret
              {config?.auth.is_default && (
                <span className="ml-2 normal-case text-[10px] font-bold text-amber-400 bg-amber-500/10 px-1.5 py-0.5 rounded-full tracking-normal border border-amber-500/20">
                  Usando valor padrao
                </span>
              )}
            </label>
            <input
              type="password"
              value={jwtSecret}
              onChange={(e) => setJwtSecret(e.target.value)}
              placeholder={config?.auth.jwt_secret ? `Atual: ${config.auth.jwt_secret}` : 'Nova chave secreta...'}
              className={inputCls}
              minLength={8}
              autoComplete="new-password"
            />
            <p className="mt-1.5 text-[10px] text-zinc-600">
              Minimo de 8 caracteres. Deixe em branco para manter o valor atual.
            </p>
          </div>
        </div>

        <div className={sectionCls}>
          <SectionHeader
            icon={<Server size={16} className="text-emerald-400" />}
            title="Servidor"
            description="Configuracoes do servidor HTTP local"
          />
          <div>
            <label className={labelCls}>Porta HTTP</label>
            <input
              type="number"
              value={serverPort}
              onChange={(e) => setServerPort(e.target.value)}
              placeholder="3009"
              className={inputCls}
              min={1}
              max={65535}
              required
            />
          </div>
        </div>

        <div className={sectionCls}>
          <SectionHeader
            icon={<RefreshCw size={16} className="text-sky-400" />}
            title="Conectividade"
            description="Politicas para baixa conectividade, degradacao e cadencia de sincronizacao"
          />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className={labelCls}>Fila Intermittent (cloud sync)</label>
              <input type="number" value={intermittentCloudSyncQueueDepth} onChange={(e) => setIntermittentCloudSyncQueueDepth(e.target.value)} className={inputCls} min={1} />
            </div>
            <div>
              <label className={labelCls}>Fila Degraded (cloud sync)</label>
              <input type="number" value={degradedCloudSyncQueueDepth} onChange={(e) => setDegradedCloudSyncQueueDepth(e.target.value)} className={inputCls} min={1} />
            </div>
            <div>
              <label className={labelCls}>Idade Intermittent (s)</label>
              <input type="number" value={intermittentPendingAgeSeconds} onChange={(e) => setIntermittentPendingAgeSeconds(e.target.value)} className={inputCls} min={1} />
            </div>
            <div>
              <label className={labelCls}>Idade Degraded (s)</label>
              <input type="number" value={degradedPendingAgeSeconds} onChange={(e) => setDegradedPendingAgeSeconds(e.target.value)} className={inputCls} min={1} />
            </div>
            <div>
              <label className={labelCls}>Fila Degraded (MQTT)</label>
              <input type="number" value={degradedMqttQueueDepth} onChange={(e) => setDegradedMqttQueueDepth(e.target.value)} className={inputCls} min={1} />
            </div>
            <div />
            <div>
              <label className={labelCls}>Heartbeat Healthy (s)</label>
              <input type="number" value={heartbeatHealthySeconds} onChange={(e) => setHeartbeatHealthySeconds(e.target.value)} className={inputCls} min={5} />
            </div>
            <div>
              <label className={labelCls}>Heartbeat Degraded (s)</label>
              <input type="number" value={heartbeatDegradedSeconds} onChange={(e) => setHeartbeatDegradedSeconds(e.target.value)} className={inputCls} min={5} />
            </div>
            <div>
              <label className={labelCls}>Stats Healthy (s)</label>
              <input type="number" value={statsHealthySeconds} onChange={(e) => setStatsHealthySeconds(e.target.value)} className={inputCls} min={10} />
            </div>
            <div>
              <label className={labelCls}>Stats Degraded (s)</label>
              <input type="number" value={statsDegradedSeconds} onChange={(e) => setStatsDegradedSeconds(e.target.value)} className={inputCls} min={10} />
            </div>
          </div>
          <p className="mt-3 text-[10px] text-zinc-600 leading-relaxed">
            Quando o Edge entrar em modo <span className="text-zinc-400">intermittent</span> ou <span className="text-zinc-400">degraded</span>,
            ele reduz heartbeat e stats automaticamente e passa a sinalizar isso no snapshot operacional local e no Spark Cloud.
          </p>
        </div>

        <div className={sectionCls}>
          <SectionHeader
            icon={<Database size={16} className="text-fuchsia-400" />}
            title="Retencao Local"
            description="Limites e janelas de limpeza automatica para filas locais e historicos terminais"
          />
          <div className="space-y-5">
            <div>
              <p className="text-xs font-semibold text-white mb-3">MQTT Queue</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className={labelCls}>Max itens MQTT</label>
                  <input type="number" value={mqttQueueMaxItems} onChange={(e) => setMqttQueueMaxItems(e.target.value)} className={inputCls} min={100} />
                </div>
                <div>
                  <label className={labelCls}>Max idade MQTT (h)</label>
                  <input type="number" value={mqttQueueMaxAgeHours} onChange={(e) => setMqttQueueMaxAgeHours(e.target.value)} className={inputCls} min={1} />
                </div>
              </div>
            </div>

            <div>
              <p className="text-xs font-semibold text-white mb-3">Cloud Sync History</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className={labelCls}>Sent retention (h)</label>
                  <input type="number" value={cloudSyncSentRetentionHours} onChange={(e) => setCloudSyncSentRetentionHours(e.target.value)} className={inputCls} min={1} />
                </div>
                <div>
                  <label className={labelCls}>Failed retention (h)</label>
                  <input type="number" value={cloudSyncFailedRetentionHours} onChange={(e) => setCloudSyncFailedRetentionHours(e.target.value)} className={inputCls} min={1} />
                </div>
                <div>
                  <label className={labelCls}>Keep sent items</label>
                  <input type="number" value={cloudSyncKeepSentItems} onChange={(e) => setCloudSyncKeepSentItems(e.target.value)} className={inputCls} min={100} />
                </div>
                <div>
                  <label className={labelCls}>Keep failed items</label>
                  <input type="number" value={cloudSyncKeepFailedItems} onChange={(e) => setCloudSyncKeepFailedItems(e.target.value)} className={inputCls} min={100} />
                </div>
              </div>
            </div>

            <div>
              <p className="text-xs font-semibold text-white mb-3">Local Fallback History</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className={labelCls}>Sent retention (h)</label>
                  <input type="number" value={localFallbackSentRetentionHours} onChange={(e) => setLocalFallbackSentRetentionHours(e.target.value)} className={inputCls} min={1} />
                </div>
                <div>
                  <label className={labelCls}>Failed retention (h)</label>
                  <input type="number" value={localFallbackFailedRetentionHours} onChange={(e) => setLocalFallbackFailedRetentionHours(e.target.value)} className={inputCls} min={1} />
                </div>
                <div>
                  <label className={labelCls}>Keep sent items</label>
                  <input type="number" value={localFallbackKeepSentItems} onChange={(e) => setLocalFallbackKeepSentItems(e.target.value)} className={inputCls} min={100} />
                </div>
                <div>
                  <label className={labelCls}>Keep failed items</label>
                  <input type="number" value={localFallbackKeepFailedItems} onChange={(e) => setLocalFallbackKeepFailedItems(e.target.value)} className={inputCls} min={100} />
                </div>
              </div>
            </div>
          </div>
          <p className="mt-3 text-[10px] text-zinc-600 leading-relaxed">
            Essa secao controla limpeza automatica do historico local. Ela nao descarta silenciosamente itens pendentes de cloud sync ou fallback; atua sobre historico terminal e limite da fila MQTT.
          </p>
        </div>

        <div className="pt-2">
          {saved && (
            <div className="flex items-center gap-2 mb-4 text-emerald-400 text-sm animate-in fade-in">
              <CheckCircle2 size={16} />
              <span>Configuracoes salvas. Reinicie o servico para aplicar.</span>
            </div>
          )}

          <Button
            type="submit"
            disabled={saving}
            className="w-full h-12 gap-2 bg-amber-500 text-zinc-900 hover:bg-amber-400 font-semibold transition-all active:scale-[0.98] shadow-lg shadow-amber-500/20"
          >
            {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            {saving ? 'Salvando...' : 'Salvar Configuracoes'}
          </Button>
        </div>
      </form>
    </main>
  );
}
