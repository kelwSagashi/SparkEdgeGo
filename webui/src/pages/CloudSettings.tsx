import { useState, useEffect, useCallback, useRef } from 'react';
import { cloudService, type CloudStatus } from '@/rest-api-client/cloud.service';
import { cloudSyncService, type CloudSyncItem, type CloudSyncStats } from '@/rest-api-client/cloud-sync.service';
import { Button } from '@/components/ui/button';
import {
  Wifi,
  Loader2,
  Unplug,
  RefreshCw,
  Mail,
  Lock,
  Zap,
  CheckCircle2,
  AlertCircle,
  PlugZap,
  Building2,
  MapPin,
  Tag,
  ArrowRight,
  Settings2,
  Navigation,
  MousePointer2,
  Trash2,
  Key,
  Database,
  Send,
  Clock3,
  Copy,
  Search,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import { toast } from 'sonner';

import { MapContainer, TileLayer, Marker, useMapEvents, useMap } from 'react-leaflet';
import L from 'leaflet';

import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png';

const DefaultIcon = L.icon({
  iconUrl: markerIcon,
  iconRetinaUrl: markerIcon2x,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  tooltipAnchor: [16, -28],
  shadowSize: [41, 41],
});

L.Marker.prototype.options.icon = DefaultIcon;

const inputCls =
  'w-full px-4 py-3 bg-white/[0.04] border border-white/[0.1] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] focus:bg-white/[0.06] transition-all';
const labelCls = 'block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2';

type Step = 'loading' | 'onboarding' | 'connection' | 'connected';

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

function formatDuration(seconds?: number | null) {
  if (!seconds || seconds <= 0) return '-';
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min`;
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return minutes > 0 ? `${hours}h ${minutes}min` : `${hours}h`;
}

function LocationMarker({ position, setPosition }: { position: L.LatLng | null; setPosition: (p: L.LatLng) => void }) {
  useMapEvents({
    click(e) {
      setPosition(e.latlng);
    },
  });

  return position === null ? null : <Marker position={position} />;
}

function MapCenter({ position }: { position: L.LatLng | null }) {
  const map = useMap();
  useEffect(() => {
    if (position) {
      map.setView(position, map.getZoom());
    }
  }, [position, map]);
  return null;
}

export default function CloudSettingsPage() {
  const [step, setStep] = useState<Step>('loading');
  const [manualStepOverride, setManualStepOverride] = useState<Step | null>(null);
  const [status, setStatus] = useState<CloudStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [token, setToken] = useState('');
  const [useToken, setUseToken] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [syncStats, setSyncStats] = useState<CloudSyncStats | null>(null);
  const [syncItems, setSyncItems] = useState<CloudSyncItem[]>([]);
  const [syncLoading, setSyncLoading] = useState(false);
  const [syncStatusFilter, setSyncStatusFilter] = useState<'all' | 'pending' | 'failed' | 'sent'>('all');
  const [syncTypeFilter, setSyncTypeFilter] = useState('all');
  const [syncSearch, setSyncSearch] = useState('');
  const [syncPageSize, setSyncPageSize] = useState(5);
  const [syncPage, setSyncPage] = useState(1);
  const [expandedPayloads, setExpandedPayloads] = useState<Record<string, boolean>>({});
  const [itemActionLoading, setItemActionLoading] = useState<Record<string, 'retry' | 'delete' | null>>({});

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [location, setLocation] = useState<L.LatLng | null>(null);
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const unwrapCloudSyncArray = (value: unknown): CloudSyncItem[] => {
    if (Array.isArray(value)) {
      return value;
    }
    if (value && typeof value === 'object' && Array.isArray((value as { data?: unknown }).data)) {
      return (value as { data: CloudSyncItem[] }).data;
    }
    return [];
  };

  const eventTypeOf = (item: Partial<CloudSyncItem> | null | undefined) =>
    String(item?.event_type ?? '').trim();

  const statusOf = (item: Partial<CloudSyncItem> | null | undefined) =>
    String(item?.status ?? '').trim();

  const unwrapCloudSyncObject = (value: unknown): CloudSyncStats | null => {
    if (value && typeof value === 'object') {
      if ('data' in (value as Record<string, unknown>)) {
        const data = (value as { data?: unknown }).data;
        return data && typeof data === 'object' ? (data as CloudSyncStats) : null;
      }
      return value as CloudSyncStats;
    }
    return null;
  };

  const fetchStatus = useCallback(async () => {
    try {
      const s = await cloudService.getStatus();
      const statusData = s.data;
      setStatus(statusData);

      if (statusData.connected) {
        setStep('connected');
        setManualStepOverride(null);
        return;
      }

      const onb = await cloudService.getOnboarding();
      if (onb.data.complete) {
        if (manualStepOverride === 'onboarding') {
          setStep('onboarding');
        } else {
          setStep('connection');
        }
        return;
      }

      setStep('onboarding');
      setManualStepOverride(null);
      if (onb.data.data) {
        setName(onb.data.data.name || '');
        setDescription(onb.data.data.description || '');
        if (onb.data.data.lat && onb.data.data.lng) {
          setLocation(new L.LatLng(Number(onb.data.data.lat), Number(onb.data.data.lng)));
        }
        setTags(onb.data.data.tags || []);
      }
    } catch {
      setStep('onboarding');
    }
  }, [manualStepOverride]);

  const fetchSync = useCallback(async () => {
    setSyncLoading(true);
    try {
      const [statsRes, listRes] = await Promise.all([
        cloudSyncService.stats(),
        cloudSyncService.list(),
      ]);
      setSyncStats(unwrapCloudSyncObject(statsRes));
      setSyncItems(unwrapCloudSyncArray(listRes));
    } catch {
      // Keep cloud setup flow usable even if sync panel fails.
    } finally {
      setSyncLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchStatus();
    void fetchSync();

    pollingRef.current = setInterval(async () => {
      try {
        const [statusRes, statsRes, listRes] = await Promise.all([
          cloudService.getStatus(),
          cloudSyncService.stats(),
          cloudSyncService.list(),
        ]);
        setStatus(statusRes.data);
        setSyncStats(unwrapCloudSyncObject(statsRes));
        setSyncItems(unwrapCloudSyncArray(listRes));
      } catch {
        // Ignore background polling errors.
      }
    }, 8000);

    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current);
    };
  }, [fetchStatus, fetchSync]);

  const handleSaveOnboarding = async (e: React.FormEvent) => {
    e.preventDefault();

    setError(null);
    setActionLoading(true);
    try {
      await cloudService.saveOnboarding({
        name,
        description,
        lat: location ? String(location.lat) : undefined,
        lng: location ? String(location.lng) : undefined,
        tags,
      });
      setManualStepOverride(null);
      setStep('connection');
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao salvar dados de onboarding.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleLocateMe = () => {
    if (!navigator.geolocation) {
      setError('Geolocalizacao nao suportada pelo navegador.');
      return;
    }

    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocation(new L.LatLng(pos.coords.latitude, pos.coords.longitude));
      },
      () => {
        setError('Nao foi possivel obter sua localizacao. Permita o acesso ao GPS.');
      }
    );
  };

  const handlePair = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setActionLoading(true);
    try {
      await cloudService.pair(token);
      setToken('');
      await fetchStatus();
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao vincular dispositivo.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setActionLoading(true);
    try {
      await cloudService.connect({ email, password });
      setPassword('');
      await fetchStatus();
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao conectar ao Spark Cloud. Verifique suas credenciais.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDisconnect = async () => {
    setError(null);
    setActionLoading(true);
    try {
      await cloudService.disconnect();
      await fetchStatus();
    } catch (err: any) {
      setError(err?.message ?? 'Erro ao desconectar.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRemove = async () => {
    if (!confirm('AVISO: isso vai remover completamente a identidade deste Edge e desconectar do Spark Cloud. Deseja continuar?')) {
      return;
    }

    setError(null);
    setActionLoading(true);
    try {
      await cloudService.remove();
      setName('');
      setDescription('');
      setLocation(null);
      setTags([]);
      setEmail('');
      setPassword('');
      setStep('onboarding');
      await fetchStatus();
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao remover conexao.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleReconnect = async () => {
    setError(null);
    setActionLoading(true);
    try {
      await cloudService.reconnect();
      await fetchStatus();
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao reconectar.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleFlushSync = async () => {
    setError(null);
    setActionLoading(true);
    try {
      const result = await cloudSyncService.flush();
      await fetchSync();
      toast.success(`Fila sincronizada: ${result.sent ?? 0} enviados, ${result.failed ?? 0} falhas.`);
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao sincronizar a fila local com o Spark Cloud.');
    } finally {
      setActionLoading(false);
    }
  };

  const formatDateTime = (value?: string | null) => {
    if (!value) return '-';
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return value;
    return parsed.toLocaleString('pt-BR');
  };

  const formatPayloadText = (payload?: Record<string, unknown>) => {
    if (!payload) return '-';
    try {
      return JSON.stringify(payload, null, 2);
    } catch {
      return '[payload nao serializavel]';
    }
  };

  const formatPayloadPreview = (payload?: Record<string, unknown>) => {
    const serialized = formatPayloadText(payload);
    return serialized.length > 220 ? `${serialized.slice(0, 220)}...` : serialized;
  };

  const togglePayloadExpanded = (itemID: string) => {
    setExpandedPayloads((current) => ({
      ...current,
      [itemID]: !current[itemID],
    }));
  };

  const handleCopyPayload = async (item: CloudSyncItem) => {
    try {
      await navigator.clipboard.writeText(formatPayloadText(item.payload));
      toast.success(`Payload do evento ${item.event_type} copiado.`);
    } catch {
      toast.error('Nao foi possivel copiar o payload.');
    }
  };

  const setItemLoading = (id: string, action: 'retry' | 'delete' | null) => {
    setItemActionLoading((current) => ({ ...current, [id]: action }));
  };

  const handleRetryItem = async (item: CloudSyncItem) => {
    setError(null);
    setItemLoading(item.id, 'retry');
    try {
      const result = await cloudSyncService.retry(item.id);
      await fetchSync();
      if (result.sent) {
        toast.success(`Evento ${item.event_type} enviado com sucesso.`);
      } else if (result.skipped) {
        toast.warning(result.message ?? 'Sincronizacao cloud nao configurada.');
      } else {
        toast.error(result.last_error ?? 'Falha ao reenviar evento.');
      }
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao reenviar item da fila cloud.');
    } finally {
      setItemLoading(item.id, null);
    }
  };

  const handleDeleteItem = async (item: CloudSyncItem) => {
    if (!confirm(`Remover o evento ${item.event_type} da fila local?`)) {
      return;
    }
    setError(null);
    setItemLoading(item.id, 'delete');
    try {
      await cloudSyncService.remove(item.id);
      await fetchSync();
      toast.success(`Evento ${item.event_type} removido da fila.`);
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao remover item da fila cloud.');
    } finally {
      setItemLoading(item.id, null);
    }
  };

  const extractPayloadSummary = (payload?: Record<string, unknown>) => {
    if (!payload) {
      return [] as Array<{ label: string; value: string }>;
    }

    const preferredKeys = [
      ['message_id', 'Message ID'],
      ['edge_id', 'Edge ID'],
      ['execution_id', 'Execution'],
      ['instance_id', 'Instancia'],
      ['command_id', 'Comando'],
      ['type', 'Tipo'],
      ['status', 'Status payload'],
      ['topic', 'Topico'],
    ] as const;

    return preferredKeys
      .map(([key, label]) => {
        const value = payload[key];
        if (value === undefined || value === null || value === '') {
          return null;
        }
        if (typeof value === 'object') {
          try {
            return { label, value: JSON.stringify(value) };
          } catch {
            return { label, value: '[objeto]' };
          }
        }
        return { label, value: String(value) };
      })
      .filter((item): item is { label: string; value: string } => item !== null);
  };

  const describeEvent = (item: CloudSyncItem) => {
    const payload = item.payload ?? {};
    const messageID = typeof payload.message_id === 'string' ? payload.message_id : null;
    const edgeID = typeof payload.edge_id === 'string' ? payload.edge_id : null;
    const executionID =
      typeof payload.execution_id === 'string'
        ? payload.execution_id
        : typeof payload.instance_execution_id === 'string'
          ? payload.instance_execution_id
          : null;
    const instanceID = typeof payload.instance_id === 'string' ? payload.instance_id : null;
    const topic = typeof payload.topic === 'string' ? payload.topic : null;
    const state = typeof payload.state === 'string' ? payload.state : null;
    const commandType = typeof payload.command_type === 'string' ? payload.command_type : null;

    const lines: string[] = [];
    if (messageID) lines.push(`Mensagem ${messageID}`);
    if (executionID) lines.push(`Execucao ${executionID}`);
    if (instanceID) lines.push(`Instancia ${instanceID}`);
    if (edgeID) lines.push(`Edge ${edgeID}`);
    if (topic) lines.push(`Topico ${topic}`);
    if (state) lines.push(`Estado ${state}`);
    if (commandType) lines.push(`Comando ${commandType}`);

    if (lines.length > 0) {
      return lines;
    }

    switch (item.event_type) {
      case 'instance_execution':
        return ['Execucao de instancia aguardando sincronizacao com o Spark Cloud.'];
      case 'paired':
        return ['Pareamento local registrado para envio posterior ao Spark Cloud.'];
      case 'registered':
        return ['Registro de edge aguardando confirmacao na nuvem.'];
      case 'mqtt_disconnected':
        return ['Desconexao MQTT capturada localmente e pendente de sincronizacao.'];
      case 'mqtt_reconnected':
        return ['Reconexao MQTT capturada localmente e pendente de sincronizacao.'];
      default:
        return ['Evento local aguardando envio assistido ao Spark Cloud.'];
    }
  };

  const safeSyncItems = Array.isArray(syncItems) ? syncItems : [];
  const syncEventTypes = Array.from(
    new Set(safeSyncItems.map((item) => String(item.event_type ?? '').trim()).filter(Boolean)),
  ).sort();

  const filteredSyncItems = safeSyncItems.filter((item) => {
    if (syncStatusFilter !== 'all' && statusOf(item) !== syncStatusFilter) {
      return false;
    }
    if (syncTypeFilter !== 'all' && eventTypeOf(item) !== syncTypeFilter) {
      return false;
    }
    if (syncSearch.trim() !== '') {
      const haystack = [
        item.id,
        eventTypeOf(item),
        statusOf(item),
        item.last_error ?? '',
        formatPayloadText(item.payload),
      ]
        .join(' ')
        .toLowerCase();
      if (!haystack.includes(syncSearch.trim().toLowerCase())) {
        return false;
      }
    }
    return true;
  });

  useEffect(() => {
    setSyncPage(1);
  }, [syncStatusFilter, syncTypeFilter, syncSearch, syncPageSize]);

  const totalSyncPages = Math.max(1, Math.ceil(filteredSyncItems.length / syncPageSize));
  const currentSyncPage = Math.min(syncPage, totalSyncPages);
  const pagedSyncItems = filteredSyncItems.slice(
    (currentSyncPage - 1) * syncPageSize,
    currentSyncPage * syncPageSize,
  );
  const connectivityInfo = status?.connectivity ?? syncStats?.connectivity;
  const connectivityMode = String(connectivityInfo?.mode || 'unknown').toLowerCase();
  const cloudPressurePct = Math.max(
    Number(syncStats?.usage?.pending_total_pct_of_failed_window ?? 0),
    Number(syncStats?.usage?.sent_pct_of_sent_window ?? 0),
  );
  const mqttUsagePct = Number(syncStats?.mqtt_queue?.usage_pct ?? 0);
  const oldestCloudPendingAgeSeconds = Number(syncStats?.oldest_pending_age_seconds ?? 0);
  const oldestMqttPendingAgeSeconds = Number(syncStats?.mqtt_queue?.oldest_pending_age_seconds ?? 0);
  const cloudPressureTone = severityFromPercent(cloudPressurePct);
  const mqttPressureTone = severityFromPercent(mqttUsagePct);
  const oldestCloudPendingTone = severityFromAgeSeconds(oldestCloudPendingAgeSeconds);
  const oldestMqttPendingTone = severityFromAgeSeconds(oldestMqttPendingAgeSeconds);
  const offlineSignals = [
    !status?.mqtt.connected ? 'broker mqtt offline' : null,
    connectivityMode === 'degraded' || connectivityMode === 'offline' ? `modo ${connectivityMode}` : null,
    (syncStats?.pending ?? 0) > 0 ? `${syncStats?.pending ?? 0} evento(s) aguardando cloud sync` : null,
    (syncStats?.failed ?? 0) > 0 ? `${syncStats?.failed ?? 0} evento(s) com falha` : null,
    (syncStats?.mqtt_queue?.total ?? 0) > 0 ? `${syncStats?.mqtt_queue?.total ?? 0} item(ns) na fila mqtt local` : null,
  ].filter((value): value is string => Boolean(value));
  const operationalPosture = (syncStats?.failed ?? 0) > 0 || connectivityMode === 'degraded' || connectivityMode === 'offline'
    ? {
        title: 'Operacao em contingencia',
        description: 'O Edge esta preservando ou reprocessando eventos localmente. Vale acompanhar backlog, fila MQTT e causas de degradacao antes de considerar a conexao estabilizada.',
        tone: 'border-red-500/25 bg-red-500/[0.08] text-red-100',
      }
    : (syncStats?.pending ?? 0) > 0 || (syncStats?.mqtt_queue?.total ?? 0) > 0
      ? {
          title: 'Sincronizacao em recuperacao',
          description: 'Existe backlog controlado. O fluxo assistido esta mantendo a entrega eventual, mas ainda ha itens em fila local.',
          tone: 'border-amber-500/25 bg-amber-500/[0.08] text-amber-100',
        }
      : {
          title: 'Fluxo online saudavel',
          description: 'Cloud sync, MQTT e filas locais estao em um estado baixo de pressao operacional.',
          tone: 'border-emerald-500/25 bg-emerald-500/[0.08] text-emerald-100',
        };
  const quickQueueLanes = [
    { key: 'all', label: 'Todos', count: safeSyncItems.length, onClick: () => { setSyncStatusFilter('all'); setSyncTypeFilter('all'); } },
    { key: 'pending', label: 'Pendentes', count: safeSyncItems.filter((item) => item.status === 'pending').length, onClick: () => setSyncStatusFilter('pending') },
    { key: 'failed', label: 'Falhas', count: safeSyncItems.filter((item) => item.status === 'failed').length, onClick: () => setSyncStatusFilter('failed') },
    { key: 'mqtt', label: 'MQTT', count: safeSyncItems.filter((item) => String(item.event_type ?? '').toLowerCase().includes('mqtt')).length, onClick: () => setSyncTypeFilter('mqtt_disconnected') },
    { key: 'exec', label: 'Execucoes', count: safeSyncItems.filter((item) => item.event_type === 'instance_execution').length, onClick: () => setSyncTypeFilter('instance_execution') },
  ];

  const addTag = () => {
    const value = newTag.trim();
    if (value && !tags.includes(value)) {
      setTags([...tags, value]);
      setNewTag('');
    }
  };

  const removeTag = (tag: string) => setTags(tags.filter((item) => item !== tag));

  const MqttBadge = ({ connected }: { connected: boolean }) => (
    <span
      className={`inline-flex items-center gap-1.5 text-[11px] font-medium px-2.5 py-0.5 rounded-full ${
        connected
          ? 'bg-emerald-500/15 text-emerald-400 ring-1 ring-emerald-500/30'
          : 'bg-zinc-700/50 text-zinc-400 ring-1 ring-white/10'
      }`}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'}`} />
      {connected ? 'MQTT online' : 'MQTT offline'}
    </span>
  );

  const ConnectivityBadge = () => {
    const tone =
      connectivityMode === 'healthy'
        ? 'bg-emerald-500/15 text-emerald-300 ring-1 ring-emerald-500/30'
        : connectivityMode === 'intermittent'
          ? 'bg-amber-500/15 text-amber-300 ring-1 ring-amber-500/30'
          : connectivityMode === 'degraded' || connectivityMode === 'offline'
            ? 'bg-red-500/15 text-red-300 ring-1 ring-red-500/30'
            : 'bg-zinc-700/50 text-zinc-300 ring-1 ring-white/10';

    return (
      <span className={`inline-flex items-center gap-1.5 text-[11px] font-medium px-2.5 py-0.5 rounded-full ${tone}`}>
        <span className="w-1.5 h-1.5 rounded-full bg-current" />
        {connectivityMode === 'unknown' ? 'Conectividade desconhecida' : `Modo ${connectivityMode}`}
      </span>
    );
  };

  if (step === 'loading') {
    return (
      <main className="grow px-8 py-6 w-full max-w-[600px] mx-auto">
        <div className="flex items-center gap-2 text-zinc-500 mt-24 justify-center">
          <Loader2 size={18} className="animate-spin" />
          <span className="text-sm">Verificando status...</span>
        </div>
      </main>
    );
  }

  return (
    <main className="grow px-8 py-6 w-full max-w-[600px] mx-auto animate-in fade-in duration-500">
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-cyan-500 flex items-center justify-center shadow-lg shadow-emerald-500/20">
            <Zap size={14} className="text-white" />
          </div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Spark Cloud</h1>
        </div>
        <p className="text-sm text-zinc-500 mt-1 ml-11">
          Gerencie a conexao e a identidade deste Edge na nuvem.
        </p>
      </div>

      {error && (
        <div className="flex items-start gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3 mb-6 text-sm text-red-400 animate-in slide-in-from-top-2">
          <AlertCircle size={16} className="mt-0.5 shrink-0" />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto hover:text-white">x</button>
        </div>
      )}

      {step === 'onboarding' && (
        <div className="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-6 backdrop-blur-md shadow-xl">
          <div className="flex items-center gap-3 mb-6 pb-5 border-b border-white/[0.06]">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/10 flex items-center justify-center">
              <Settings2 size={18} className="text-cyan-400" />
            </div>
            <div>
              <p className="text-sm font-semibold text-white">Passo 1: Configuracao Local</p>
              <p className="text-xs text-zinc-500">Identifique este Edge antes de registra-lo</p>
            </div>
          </div>

          <form onSubmit={handleSaveOnboarding} className="space-y-6">
            <div>
              <label className={labelCls}><Building2 size={11} className="inline mr-1" /> Nome do Edge</label>
              <input
                type="text"
                placeholder="Ex: Edge Laboratorio 01"
                className={inputCls}
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                disabled={actionLoading}
              />
            </div>

            <div>
              <label className={labelCls}>Descricao (Opcional)</label>
              <textarea
                placeholder="Uma breve descricao sobre a finalidade deste dispositivo..."
                className={`${inputCls} min-h-[80px] py-3 resize-none`}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={actionLoading}
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <label className={labelCls}><MapPin size={11} className="inline mr-1" /> Localizacao</label>
                <div className="flex items-center gap-3">
                  {location ? (
                    <button
                      type="button"
                      onClick={() => setLocation(null)}
                      className="text-[10px] uppercase font-bold text-rose-400 hover:text-rose-300 flex items-center gap-1 transition-colors"
                    >
                      <Trash2 size={10} />
                      Limpar
                    </button>
                  ) : null}
                  <button
                    type="button"
                    onClick={handleLocateMe}
                    className="text-[10px] uppercase font-bold text-cyan-500 hover:text-cyan-400 flex items-center gap-1 transition-colors"
                  >
                    <Navigation size={10} />
                    Usar meu GPS
                  </button>
                </div>
              </div>

              <div className="relative h-[240px] rounded-xl overflow-hidden border border-white/10 group">
                <MapContainer
                  center={[-23.5505, -46.6333]}
                  zoom={13}
                  style={{ height: '100%', width: '100%', filter: 'grayscale(100%) invert(100%) contrast(90%)' }}
                  zoomControl={false}
                >
                  <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
                  <LocationMarker position={location} setPosition={setLocation} />
                  <MapCenter position={location} />
                </MapContainer>

                {!location && (
                  <div className="absolute inset-0 bg-black/40 backdrop-blur-[2px] flex flex-col items-center justify-center p-6 text-center pointer-events-none transition-opacity group-hover:opacity-0">
                    <MousePointer2 size={24} className="text-white/40 mb-2 animate-bounce" />
                    <p className="text-xs text-white/60 font-medium">Clique no mapa para marcar a posicao do Edge</p>
                    <p className="mt-1 text-[11px] text-white/40">Opcional: voce pode pular agora e definir depois no Edge ou no Spark Cloud.</p>
                  </div>
                )}

                {location && (
                  <div className="absolute bottom-3 left-3 right-3 bg-zinc-900/90 backdrop-blur-md px-3 py-2 rounded-lg border border-white/10 flex items-center justify-between shadow-2xl">
                    <span className="text-[10px] font-mono text-zinc-400">
                      {location.lat.toFixed(5)}, {location.lng.toFixed(5)}
                    </span>
                    <span className="text-[10px] font-bold text-emerald-400 uppercase tracking-widest">Pin Definido</span>
                  </div>
                )}
              </div>
            </div>

            <div>
              <label className={labelCls}><Tag size={11} className="inline mr-1" /> Tags</label>
              <div className="flex gap-2 mb-3">
                <input
                  type="text"
                  placeholder="producao, sensores, piape..."
                  className={inputCls}
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
                  disabled={actionLoading}
                />
                <Button type="button" variant="outline" className="border-white/10" onClick={addTag}>Add</Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {tags.map((tag) => (
                  <span key={tag} className="flex items-center gap-1.5 px-2 py-1 bg-white/5 border border-white/10 rounded-lg text-xs text-zinc-300 animate-in zoom-in-90">
                    {tag}
                    <button type="button" onClick={() => removeTag(tag)} className="hover:text-red-400">x</button>
                  </span>
                ))}
              </div>
            </div>

            <Button
              type="submit"
              className="w-full gap-2 bg-white text-zinc-900 hover:bg-zinc-100 font-medium h-11 transition-all active:scale-95"
              disabled={actionLoading || !name}
            >
              {actionLoading ? <Loader2 size={15} className="animate-spin" /> : <ArrowRight size={15} />}
              Proximo Passo: Conectar Cloud
            </Button>
          </form>
        </div>
      )}

      {step === 'connection' && (
        <div className="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-6 backdrop-blur-md shadow-xl animate-in slide-in-from-right-4 duration-300">
          <div className="flex items-center gap-3 mb-6 pb-5 border-b border-white/[0.06]">
            <div className="w-10 h-10 rounded-xl bg-purple-500/10 flex items-center justify-center">
              <PlugZap size={18} className="text-purple-400" />
            </div>
            <div>
              <p className="text-sm font-semibold text-white">Passo 2: Vincular ao Spark</p>
              <p className="text-xs text-zinc-500">Escolha como conectar "{name}" ao Cloud</p>
            </div>
          </div>

          <div className="flex p-1 bg-white/5 rounded-xl border border-white/5 mb-6">
            <button
              onClick={() => setUseToken(true)}
              className={`flex-1 py-1.5 text-[10px] font-bold uppercase tracking-wider rounded-lg transition-all ${useToken ? 'bg-white/10 text-white shadow-lg' : 'text-zinc-500 hover:text-zinc-300'}`}
            >
              Via Token
            </button>
            <button
              onClick={() => setUseToken(false)}
              className={`flex-1 py-1.5 text-[10px] font-bold uppercase tracking-wider rounded-lg transition-all ${!useToken ? 'bg-white/10 text-white shadow-lg' : 'text-zinc-500 hover:text-zinc-300'}`}
            >
              Via Email/Senha
            </button>
          </div>

          {useToken ? (
            <form onSubmit={handlePair} className="space-y-5">
              <div>
                <label className={labelCls}><Key size={11} className="inline mr-1" /> Token de Pareamento</label>
                <input
                  type="text"
                  placeholder="Cole o token gerado no dashboard"
                  className={`${inputCls} font-mono uppercase tracking-widest text-center`}
                  value={token}
                  onChange={(e) => setToken(e.target.value.toUpperCase())}
                  required
                  disabled={actionLoading}
                />
                <p className="mt-2 text-[10px] text-zinc-500 text-center leading-relaxed">
                  Para obter um token, acesse o dashboard do Spark Cloud, selecione uma Unit e clique em "Conectar Edge".
                </p>
              </div>

              <div className="flex gap-3 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  className="flex-1 border-white/5 text-zinc-400 hover:text-white"
                  onClick={() => {
                    setManualStepOverride('onboarding');
                    setStep('onboarding');
                  }}
                  disabled={actionLoading}
                >
                  Voltar
                </Button>
                <Button
                  type="submit"
                  className="flex-[2] gap-2 bg-purple-500 text-white hover:bg-purple-600 font-medium h-11 shadow-lg shadow-purple-500/20 active:scale-95 transition-all"
                  disabled={actionLoading || !token}
                >
                  {actionLoading ? <Loader2 size={15} className="animate-spin" /> : <Wifi size={15} />}
                  Vincular via Token
                </Button>
              </div>
            </form>
          ) : (
            <form onSubmit={handleConnect} className="space-y-5">
              <div>
                <label className={labelCls}><Mail size={11} className="inline mr-1" /> Email da conta Spark</label>
                <input
                  type="email"
                  placeholder="voce@exemplo.com"
                  className={inputCls}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div>
                <label className={labelCls}><Lock size={11} className="inline mr-1" /> Senha</label>
                <input
                  type="password"
                  placeholder="........"
                  className={inputCls}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="flex gap-3 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  className="flex-1 border-white/5 text-zinc-400 hover:text-white"
                  onClick={() => {
                    setManualStepOverride('onboarding');
                    setStep('onboarding');
                  }}
                  disabled={actionLoading}
                >
                  Voltar
                </Button>
                <Button
                  type="submit"
                  className="flex-[2] gap-2 bg-emerald-500 text-white hover:bg-emerald-600 font-medium h-11 shadow-lg shadow-emerald-500/20 active:scale-95 transition-all"
                  disabled={actionLoading || !email || !password}
                >
                  {actionLoading ? <Loader2 size={15} className="animate-spin" /> : <Wifi size={15} />}
                  Vincular Agora
                </Button>
              </div>
            </form>
          )}

          <button
            type="button"
            onClick={handleRemove}
            disabled={actionLoading}
            className="w-full mt-6 text-[10px] uppercase font-bold text-zinc-600 hover:text-red-400/70 transition-colors flex items-center justify-center gap-1.5"
          >
            <Trash2 size={10} />
            Resetar Configuracoes e Voltar ao Inicio
          </button>
        </div>
      )}

      {step === 'connected' && status && (
        <div className="space-y-6 animate-in zoom-in-95 duration-300">
          <div className={`rounded-2xl border px-5 py-4 ${operationalPosture.tone}`}>
            <div className="flex items-start justify-between gap-4">
              <div className="space-y-2">
                <p className="text-sm font-semibold">{operationalPosture.title}</p>
                <p className="text-xs leading-relaxed opacity-90">{operationalPosture.description}</p>
              </div>
              <div className="text-right">
                <p className="text-[10px] uppercase tracking-widest opacity-70">Backlog cloud</p>
                <p className="mt-1 text-sm font-semibold">{syncStats?.pending ?? 0} pendente(s)</p>
              </div>
            </div>
            {offlineSignals.length > 0 && (
              <div className="mt-4 flex flex-wrap gap-2">
                {offlineSignals.map((signal) => (
                  <span
                    key={signal}
                    className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest bg-black/20 border border-white/10"
                  >
                    {signal}
                  </span>
                ))}
              </div>
            )}
          </div>

          <div className="bg-emerald-500/[0.06] border border-emerald-500/20 rounded-2xl p-6 shadow-xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-32 h-32 bg-emerald-500/5 blur-3xl -mr-16 -mt-16 group-hover:bg-emerald-500/10 transition-colors duration-500" />

            <div className="flex items-start justify-between relative">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-2xl bg-emerald-500/20 flex items-center justify-center shadow-inner">
                  <CheckCircle2 size={24} className="text-emerald-400" />
                </div>
                <div>
                  <p className="text-base font-semibold text-white">Edge Provisionado</p>
                  <div className="flex items-center gap-2 mt-0.5">
                    <p className="text-xs text-emerald-400/80 font-medium">{status.edge_name || 'Edge s/ nome'}</p>
                    <span className="w-1 h-1 rounded-full bg-zinc-700" />
                    <p className="text-[10px] text-zinc-500 font-mono tracking-tighter uppercase">{status.edge_id?.substring(0, 8)}...</p>
                  </div>
                </div>
              </div>
              <div className="flex flex-col items-end gap-2">
                <MqttBadge connected={status.mqtt.connected} />
                <ConnectivityBadge />
              </div>
            </div>

            <div className="mt-8 pt-5 border-t border-white/[0.04]">
              <p className="text-[10px] uppercase tracking-widest text-zinc-500 font-bold mb-2">Detalhes da Identidade</p>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-white/[0.03] rounded-xl p-3 border border-white/[0.03]">
                  <p className="text-[9px] uppercase text-zinc-600 mb-1">Status de Conexao</p>
                  <p className={`text-xs font-medium ${status.mqtt.connected ? 'text-emerald-400' : 'text-zinc-500'}`}>
                    {status.mqtt.connected ? 'Ativo e Recebendo' : 'Aguardando Broker'}
                  </p>
                </div>
                <div className="bg-white/[0.03] rounded-xl p-3 border border-white/[0.03]">
                  <p className="text-[9px] uppercase text-zinc-600 mb-1">Protocolo</p>
                  <p className="text-xs font-medium text-zinc-400">MQTT over TLS</p>
                </div>
                <div className="bg-white/[0.03] rounded-xl p-3 border border-white/[0.03]">
                  <p className="text-[9px] uppercase text-zinc-600 mb-1">Heartbeat Atual</p>
                  <p className="text-xs font-medium text-zinc-300">
                    {connectivityInfo?.heartbeat_interval_seconds ? `${connectivityInfo.heartbeat_interval_seconds}s` : '-'}
                  </p>
                </div>
                <div className="bg-white/[0.03] rounded-xl p-3 border border-white/[0.03]">
                  <p className="text-[9px] uppercase text-zinc-600 mb-1">Stats Atual</p>
                  <p className="text-xs font-medium text-zinc-300">
                    {connectivityInfo?.stats_interval_seconds ? `${connectivityInfo.stats_interval_seconds}s` : '-'}
                  </p>
                </div>
              </div>
              {Array.isArray(connectivityInfo?.reasons) && connectivityInfo.reasons.length > 0 && (
                <div className="mt-3 flex flex-wrap gap-2">
                  {connectivityInfo.reasons.map((reason) => (
                    <span
                      key={reason}
                      className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest bg-amber-500/10 text-amber-200 border border-amber-500/20"
                    >
                      {reason}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="flex flex-col gap-3">
            <div className="flex gap-3">
              {status.mqtt.connected ? (
                <Button
                  variant="outline"
                  className="flex-1 h-11 gap-2 border-red-500/10 bg-red-500/[0.02] text-red-500/60 hover:bg-red-500/10 hover:text-red-400 transition-all active:scale-95"
                  disabled={actionLoading}
                  onClick={handleDisconnect}
                >
                  {actionLoading ? <Loader2 size={14} className="animate-spin" /> : <Unplug size={14} />}
                  Desconectar MQTT
                </Button>
              ) : (
                <Button
                  variant="outline"
                  className="flex-1 h-11 gap-2 border-emerald-500/10 bg-emerald-500/[0.02] text-emerald-500/60 hover:bg-emerald-500/10 hover:text-emerald-400 transition-all active:scale-95"
                  disabled={actionLoading}
                  onClick={handleReconnect}
                >
                  {actionLoading ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
                  Conectar MQTT
                </Button>
              )}
            </div>

            <Button
              variant="outline"
              className="w-full h-11 gap-2 border-red-500/10 bg-red-500/[0.05] text-red-500 hover:bg-red-500/20 hover:text-red-400 transition-all active:scale-95 border-dashed"
              disabled={actionLoading}
              onClick={handleRemove}
            >
              {actionLoading ? <Loader2 size={14} className="animate-spin" /> : <Trash2 size={14} />}
              Remover Identidade e Desvincular do Cloud
            </Button>
          </div>

          <div className="p-4 bg-zinc-900/50 border border-white/[0.03] rounded-xl">
            <p className="text-xs text-zinc-500 leading-relaxed italic">
              "Desconectar vai interromper o trafego de dados, mas nao vai apagar o registro deste dispositivo.
              Voce pode reconectar a qualquer momento usando as credenciais ja armazenadas."
            </p>
          </div>

          <div className="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-6 backdrop-blur-md shadow-xl">
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-5 pb-5 border-b border-white/[0.06]">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-cyan-500/10 flex items-center justify-center">
                  <Database size={18} className="text-cyan-400" />
                </div>
                <div>
                  <p className="text-sm font-semibold text-white">Fila de Sincronizacao</p>
                  <p className="text-xs text-zinc-500">Eventos locais aguardando envio assistido ao Spark Cloud</p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  className="h-10 gap-2 border-white/10"
                  disabled={syncLoading}
                  onClick={() => void fetchSync()}
                >
                  {syncLoading ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
                  Atualizar
                </Button>
                <Button
                  className="h-10 gap-2 bg-cyan-500 text-zinc-950 hover:bg-cyan-400"
                  disabled={actionLoading}
                  onClick={handleFlushSync}
                >
                  {actionLoading ? <Loader2 size={14} className="animate-spin" /> : <Send size={14} />}
                  Sincronizar Agora
                </Button>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-5">
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Configurado</p>
                <p className={`mt-1 text-sm font-medium ${syncStats?.configured ? 'text-emerald-400' : 'text-amber-400'}`}>
                  {syncStats?.configured ? 'Sim' : 'Nao'}
                </p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Pendentes</p>
                <p className="mt-1 text-sm font-medium text-white">{syncStats?.pending ?? 0}</p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Falhas</p>
                <p className="mt-1 text-sm font-medium text-amber-300">{syncStats?.failed ?? 0}</p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Enviados</p>
                <p className="mt-1 text-sm font-medium text-cyan-300">{syncStats?.sent ?? 0}</p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Prontos</p>
                <p className="mt-1 text-sm font-medium text-sky-300">{syncStats?.ready ?? 0}</p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-5">
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Retencao cloud sync</p>
                <p className="mt-1 text-xs font-medium text-zinc-200">
                  Sent: {syncStats?.retention?.sent_retention_hours ?? '-'}h | Failed: {syncStats?.retention?.failed_retention_hours ?? '-'}h
                </p>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Keep sent {syncStats?.retention?.keep_sent_items ?? '-'} | keep failed {syncStats?.retention?.keep_failed_items ?? '-'}
                </p>
              </div>
              <div className={`rounded-xl border p-3 ${cloudPressureTone.card}`}>
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Pressao da fila cloud</p>
                <div className="mt-1 flex items-center justify-between gap-3">
                  <p className={`text-xs font-medium ${cloudPressureTone.text}`}>
                    Pico: {cloudPressurePct}%
                  </p>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${cloudPressureTone.badge}`}>
                    {cloudPressureTone.label}
                  </span>
                </div>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Failed window: {syncStats?.usage?.pending_total_pct_of_failed_window ?? 0}% | Sent window: {syncStats?.usage?.sent_pct_of_sent_window ?? 0}%
                </p>
              </div>
              <div className={`rounded-xl border p-3 ${mqttPressureTone.card}`}>
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Fila MQTT local</p>
                <div className="mt-1 flex items-center justify-between gap-3">
                  <p className={`text-xs font-medium ${mqttPressureTone.text}`}>
                    {syncStats?.mqtt_queue?.total ?? 0} item(ns) | {mqttUsagePct}% do limite
                  </p>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${mqttPressureTone.badge}`}>
                    {mqttPressureTone.label}
                  </span>
                </div>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Limite {syncStats?.mqtt_queue?.retention?.max_items ?? '-'} | idade maxima {syncStats?.mqtt_queue?.retention?.max_age_hours ?? '-'}h
                </p>
              </div>
              <div className={`rounded-xl border p-3 ${oldestCloudPendingTone.card}`}>
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Backlog mais antigo</p>
                <div className="mt-1 flex items-center justify-between gap-3">
                  <p className={`text-xs font-medium ${oldestCloudPendingTone.text}`}>
                    {oldestCloudPendingAgeSeconds > 0 ? `${oldestCloudPendingAgeSeconds}s` : '-'}
                  </p>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${oldestCloudPendingTone.badge}`}>
                    {oldestCloudPendingTone.label}
                  </span>
                </div>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Mede ha quanto tempo o evento mais antigo aguarda sincronizacao.
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-5">
              <div className={`rounded-xl border p-3 ${oldestMqttPendingTone.card}`}>
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Item MQTT mais antigo</p>
                <div className="mt-1 flex items-center justify-between gap-3">
                  <p className={`text-xs font-medium ${oldestMqttPendingTone.text}`}>
                    {formatDuration(oldestMqttPendingAgeSeconds)}
                  </p>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${oldestMqttPendingTone.badge}`}>
                    {oldestMqttPendingTone.label}
                  </span>
                </div>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Quanto tempo a fila MQTT local segura o item mais antigo.
                </p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Eventos MQTT</p>
                <p className="mt-1 text-sm font-medium text-zinc-100">
                  {safeSyncItems.filter((item) => eventTypeOf(item).toLowerCase().includes('mqtt')).length}
                </p>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Mudancas de estado do broker preservadas no fluxo de sync.
                </p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Execucoes em fila</p>
                <p className="mt-1 text-sm font-medium text-zinc-100">
                  {safeSyncItems.filter((item) => eventTypeOf(item) === 'instance_execution').length}
                </p>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Resultado de instancias aguardando sincronizacao com o Spark Cloud.
                </p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Falhas acumuladas</p>
                <p className="mt-1 text-sm font-medium text-amber-200">
                  {safeSyncItems.filter((item) => statusOf(item) === 'failed').length}
                </p>
                <p className="mt-2 text-[11px] text-zinc-500">
                  Itens que merecem revisao manual antes de drenar a fila.
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-5">
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Edge ID</p>
                <p className="mt-1 text-xs font-mono text-zinc-300 break-all">{syncStats?.edge_id ?? '-'}</p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Endpoint Cloud</p>
                <p className="mt-1 text-xs font-mono text-zinc-300 break-all">{syncStats?.base_url ?? '-'}</p>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500">Idade Mais Antiga</p>
                <p className="mt-1 text-xs font-medium text-zinc-300">
                  {syncStats?.oldest_pending_age_seconds ? `${syncStats.oldest_pending_age_seconds}s` : '-'}
                </p>
              </div>
            </div>

            <div className="mb-5 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-3">
              <div className="flex flex-wrap gap-2">
                {quickQueueLanes.map((lane) => (
                  <button
                    key={lane.key}
                    type="button"
                    onClick={lane.onClick}
                    className="inline-flex items-center gap-2 rounded-full border border-white/[0.08] bg-black/20 px-3 py-2 text-[11px] font-semibold text-zinc-200 hover:bg-white/[0.05] transition-colors"
                  >
                    <span>{lane.label}</span>
                    <span className="rounded-full bg-white/[0.06] px-2 py-0.5 text-[10px] text-zinc-400">{lane.count}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-5">
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3 md:col-span-2">
                <label className="block text-[10px] uppercase tracking-widest text-zinc-500 mb-2">Busca</label>
                <div className="relative">
                  <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
                  <input
                    value={syncSearch}
                    onChange={(e) => setSyncSearch(e.target.value)}
                    placeholder="Buscar por id, tipo, status ou conteudo do payload"
                    className="w-full bg-zinc-950/70 border border-white/[0.08] rounded-lg pl-9 pr-3 py-2 text-sm text-zinc-200 placeholder:text-zinc-500"
                  />
                </div>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <label className="block text-[10px] uppercase tracking-widest text-zinc-500 mb-2">Filtro de status</label>
                <select
                  value={syncStatusFilter}
                  onChange={(e) => setSyncStatusFilter(e.target.value as 'all' | 'pending' | 'failed' | 'sent')}
                  className="w-full bg-zinc-950/70 border border-white/[0.08] rounded-lg px-3 py-2 text-sm text-zinc-200"
                >
                  <option value="all">Todos</option>
                  <option value="pending">Pending</option>
                  <option value="failed">Failed</option>
                  <option value="sent">Sent</option>
                </select>
              </div>
              <div className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-3">
                <label className="block text-[10px] uppercase tracking-widest text-zinc-500 mb-2">Filtro de tipo</label>
                <select
                  value={syncTypeFilter}
                  onChange={(e) => setSyncTypeFilter(e.target.value)}
                  className="w-full bg-zinc-950/70 border border-white/[0.08] rounded-lg px-3 py-2 text-sm text-zinc-200"
                >
                  <option value="all">Todos</option>
                  {syncEventTypes.map((eventType) => (
                    <option key={eventType} value={eventType}>
                      {eventType}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-3 mb-4 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-3">
              <div className="text-xs text-zinc-500">
                Mostrando <span className="text-zinc-200">{pagedSyncItems.length}</span> de{' '}
                <span className="text-zinc-200">{filteredSyncItems.length}</span> eventos filtrados.
              </div>
              <div className="flex items-center gap-2">
                <label className="text-[10px] uppercase tracking-widest text-zinc-500">Por pagina</label>
                <select
                  value={syncPageSize}
                  onChange={(e) => setSyncPageSize(Number(e.target.value))}
                  className="bg-zinc-950/70 border border-white/[0.08] rounded-lg px-3 py-2 text-sm text-zinc-200"
                >
                  {[5, 10, 20, 50].map((size) => (
                    <option key={size} value={size}>
                      {size}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="space-y-3">
              {filteredSyncItems.length === 0 && (
                <div className="rounded-xl border border-dashed border-white/[0.08] bg-white/[0.02] px-4 py-6 text-center text-sm text-zinc-500">
                  Nenhum evento encontrado para os filtros atuais.
                </div>
              )}

              {pagedSyncItems.map((item) => (
                <div key={item.id} className="rounded-xl border border-white/[0.06] bg-zinc-950/50 p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="text-sm font-medium text-white">{item.event_type}</p>
                      <p className="mt-1 text-[11px] font-mono text-zinc-500">{item.id}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => void handleCopyPayload(item)}
                        className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest bg-white/[0.04] text-zinc-300 border border-white/[0.08] hover:bg-white/[0.08] transition-colors"
                      >
                        <Copy size={10} />
                        Copiar
                      </button>
                      <button
                        type="button"
                        onClick={() => void handleRetryItem(item)}
                        disabled={item.status === 'sent' || itemActionLoading[item.id] !== undefined && itemActionLoading[item.id] !== null}
                        className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest bg-cyan-500/10 text-cyan-200 border border-cyan-500/20 hover:bg-cyan-500/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {itemActionLoading[item.id] === 'retry' ? <Loader2 size={10} className="animate-spin" /> : <RefreshCw size={10} />}
                        Reenviar
                      </button>
                      <button
                        type="button"
                        onClick={() => void handleDeleteItem(item)}
                        disabled={itemActionLoading[item.id] !== undefined && itemActionLoading[item.id] !== null}
                        className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest bg-red-500/10 text-red-200 border border-red-500/20 hover:bg-red-500/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {itemActionLoading[item.id] === 'delete' ? <Loader2 size={10} className="animate-spin" /> : <Trash2 size={10} />}
                        Remover
                      </button>
                      <span className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-widest ${
                        item.status === 'sent'
                          ? 'bg-emerald-500/10 text-emerald-300 border border-emerald-500/20'
                          : item.status === 'failed'
                            ? 'bg-amber-500/10 text-amber-300 border border-amber-500/20'
                            : 'bg-cyan-500/10 text-cyan-300 border border-cyan-500/20'
                      }`}>
                        <Clock3 size={10} />
                        {item.status}
                      </span>
                    </div>
                  </div>

                  <div className="mt-3 grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
                    <div className="rounded-lg bg-white/[0.02] border border-white/[0.04] px-3 py-2">
                      <p className="text-zinc-500">Tentativas</p>
                      <p className="mt-1 text-zinc-200">{item.attempts}</p>
                    </div>
                    <div className="rounded-lg bg-white/[0.02] border border-white/[0.04] px-3 py-2">
                      <p className="text-zinc-500">Proximo retry</p>
                      <p className="mt-1 text-zinc-200">{formatDateTime(item.next_retry_at)}</p>
                    </div>
                    <div className="rounded-lg bg-white/[0.02] border border-white/[0.04] px-3 py-2">
                      <p className="text-zinc-500">Ultimo envio</p>
                      <p className="mt-1 text-zinc-200">{formatDateTime(item.sent_at)}</p>
                    </div>
                  </div>

                  <div className="mt-3 rounded-lg border border-white/[0.04] bg-white/[0.015] px-3 py-3">
                    <p className="text-[10px] uppercase tracking-widest text-zinc-500 mb-2">Leitura do evento</p>
                    <div className="space-y-1">
                      {describeEvent(item).map((line, index) => (
                        <p key={`${item.id}-description-${index}`} className="text-xs text-zinc-300">
                          {line}
                        </p>
                      ))}
                    </div>
                  </div>

                  {extractPayloadSummary(item.payload).length > 0 && (
                    <div className="mt-3 flex flex-wrap gap-2">
                      {extractPayloadSummary(item.payload).map((entry) => (
                        <div
                          key={`${item.id}-${entry.label}`}
                          className="rounded-full border border-cyan-500/15 bg-cyan-500/5 px-3 py-1.5 text-[11px] text-cyan-100"
                        >
                          <span className="text-cyan-300/70">{entry.label}:</span>{' '}
                          <span className="font-mono break-all">{entry.value}</span>
                        </div>
                      ))}
                    </div>
                  )}

                  <div className="mt-3 rounded-lg bg-white/[0.02] border border-white/[0.04] px-3 py-3">
                    <div className="flex items-center justify-between gap-3 mb-2">
                      <p className="text-zinc-500 text-xs uppercase tracking-widest">Payload</p>
                      <button
                        type="button"
                        onClick={() => togglePayloadExpanded(item.id)}
                        className="text-xs text-cyan-300 hover:text-cyan-200 transition-colors"
                      >
                        {expandedPayloads[item.id] ? 'Recolher' : 'Expandir'}
                      </button>
                    </div>
                    <pre className="text-[11px] text-zinc-300 whitespace-pre-wrap break-words font-mono">
                      {expandedPayloads[item.id]
                        ? formatPayloadText(item.payload)
                        : formatPayloadPreview(item.payload)}
                    </pre>
                  </div>

                  {item.last_error && (
                    <div className="mt-3 rounded-lg border border-red-500/20 bg-red-500/5 px-3 py-2 text-xs text-red-300">
                      {item.last_error}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {filteredSyncItems.length > 0 && (
              <div className="mt-4 flex flex-col md:flex-row md:items-center md:justify-between gap-3 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-3">
                <div className="text-xs text-zinc-500">
                  Pagina <span className="text-zinc-200">{currentSyncPage}</span> de{' '}
                  <span className="text-zinc-200">{totalSyncPages}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    className="h-9 gap-2 border-white/10"
                    disabled={currentSyncPage <= 1}
                    onClick={() => setSyncPage((current) => Math.max(1, current - 1))}
                  >
                    <ChevronLeft size={14} />
                    Anterior
                  </Button>
                  <Button
                    variant="outline"
                    className="h-9 gap-2 border-white/10"
                    disabled={currentSyncPage >= totalSyncPages}
                    onClick={() => setSyncPage((current) => Math.min(totalSyncPages, current + 1))}
                  >
                    Proxima
                    <ChevronRight size={14} />
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </main>
  );
}
