import { useState, useEffect, useCallback, useRef } from 'react';
import { cloudService, type CloudStatus } from '@/rest-api-client/cloud.service';
import { Button } from '@/components/ui/button';
import {
  Wifi, WifiOff, Loader2, Unplug, RefreshCw, Mail, Lock, Zap, CheckCircle2,
  AlertCircle, PlugZap, Building2, MapPin, Tag, ArrowRight, Settings2,
  Navigation, MousePointer2, Trash2, Key
} from 'lucide-react';

// Leaflet imports
import { MapContainer, TileLayer, Marker, useMapEvents, useMap } from 'react-leaflet';
import L from 'leaflet';

// Fix for Leaflet default marker icon in Vite/React
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
  shadowSize: [41, 41]
});

L.Marker.prototype.options.icon = DefaultIcon;

const inputCls =
  'w-full px-4 py-3 bg-white/[0.04] border border-white/[0.1] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] focus:bg-white/[0.06] transition-all';
const labelCls = 'block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2';

type Step = 'loading' | 'onboarding' | 'connection' | 'connected';

/** Map helper to handle clicks */
function LocationMarker({ position, setPosition }: { position: L.LatLng | null, setPosition: (p: L.LatLng) => void }) {
  useMapEvents({
    click(e) {
      setPosition(e.latlng);
    },
  });

  return position === null ? null : (
    <Marker position={position} />
  );
}

/** Map helper to center on position */
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
  const [status, setStatus] = useState<CloudStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [token, setToken] = useState('');
  const [useToken, setUseToken] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  // Onboarding state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [location, setLocation] = useState<L.LatLng | null>(null);
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');

  // Connection state
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const s = await cloudService.getStatus();
      const statusData = s.data;
      setStatus(statusData);

      if (statusData.connected) {
        setStep('connected');
      } else {
        // Check onboarding progress
        const onb = await cloudService.getOnboarding();
        if (onb.data.complete) {
          setStep('connection');
        } else {
          setStep('onboarding');
          if (onb.data.data) {
            setName(onb.data.data.name || '');
            setDescription(onb.data.data.description || '');
            if (onb.data.data.lat && onb.data.data.lng) {
              setLocation(new L.LatLng(Number(onb.data.data.lat), Number(onb.data.data.lng)));
            }
            setTags(onb.data.data.tags || []);
          }
        }
      }
    } catch {
      setStep('onboarding');
    }
  }, []);

  useEffect(() => {
    fetchStatus().then(() => {
        if (step === 'loading') setStep('onboarding');
    });
    pollingRef.current = setInterval(async () => {
        try {
            const s = await cloudService.getStatus();
            setStatus(s.data);
        } catch { /* ignore silenty during poll */ }
    }, 8000);
    return () => { if (pollingRef.current) clearInterval(pollingRef.current); };
  }, [fetchStatus, step]);

  const handleSaveOnboarding = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!location) return;
    setError(null);
    setActionLoading(true);
    try {
      await cloudService.saveOnboarding({ 
        name, 
        description,
        lat: String(location.lat), 
        lng: String(location.lng), 
        tags 
      });
      setStep('connection');
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao salvar dados de onboarding.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleLocateMe = () => {
    if (!navigator.geolocation) {
      setError("Geolocalização não suportada pelo navegador.");
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const newPos = new L.LatLng(pos.coords.latitude, pos.coords.longitude);
        setLocation(newPos);
      },
      () => {
        setError("Não foi possível obter sua localização. Permita o acesso ao GPS.");
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
    if (!confirm('AVISO: Isso irá remover completamente a identidade deste Edge e desconectar do Spark Cloud. Deseja continuar?')) return;
    
    setError(null);
    setActionLoading(true);
    try {
      await cloudService.remove();
      // Reset local state
      setName('');
      setLocation(null);
      setTags([]);
      setEmail('');
      setPassword('');
      setStep('onboarding');
      await fetchStatus();
    } catch (err: any) {
      setError(err?.message ?? 'Falha ao remover conexão.');
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

  const addTag = () => {
    if (newTag.trim() && !tags.includes(newTag.trim())) {
      setTags([...tags, newTag.trim()]);
      setNewTag('');
    }
  };

  const removeTag = (t: string) => setTags(tags.filter(tag => tag !== t));

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
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-1">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500 to-cyan-500 flex items-center justify-center shadow-lg shadow-emerald-500/20">
            <Zap size={14} className="text-white" />
          </div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Spark Cloud</h1>
        </div>
        <p className="text-sm text-zinc-500 mt-1 ml-11">
          Gerencie a conexão e identidade deste Edge na nuvem.
        </p>
      </div>

      {/* Error banner */}
      {error && (
        <div className="flex items-start gap-3 bg-red-500/10 border border-red-500/20 rounded-xl px-4 py-3 mb-6 text-sm text-red-400 animate-in slide-in-from-top-2">
          <AlertCircle size={16} className="mt-0.5 shrink-0" />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto hover:text-white">×</button>
        </div>
      )}

      {/* ── ONBOARDING STEP ─────────────────────────────── */}
      {step === 'onboarding' && (
        <div className="bg-white/[0.03] border border-white/[0.08] rounded-2xl p-6 backdrop-blur-md shadow-xl">
          <div className="flex items-center gap-3 mb-6 pb-5 border-b border-white/[0.06]">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/10 flex items-center justify-center">
              <Settings2 size={18} className="text-cyan-400" />
            </div>
            <div>
              <p className="text-sm font-semibold text-white">Passo 1: Configuração Local</p>
              <p className="text-xs text-zinc-500">Identifique este Edge antes de registrá-lo</p>
            </div>
          </div>

          <form onSubmit={handleSaveOnboarding} className="space-y-6">
            <div>
              <label className={labelCls}><Building2 size={11} className="inline mr-1" /> Nome do Edge</label>
              <input
                type="text"
                placeholder="Ex: Edge Laboratório 01"
                className={inputCls}
                value={name}
                onChange={e => setName(e.target.value)}
                required
                disabled={actionLoading}
              />
            </div>

            <div>
              <label className={labelCls}>Descrição (Opcional)</label>
              <textarea
                placeholder="Uma breve descrição sobre a finalidade deste dispositivo..."
                className={`${inputCls} min-h-[80px] py-3 resize-none`}
                value={description}
                onChange={e => setDescription(e.target.value)}
                disabled={actionLoading}
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-2">
                <label className={labelCls}><MapPin size={11} className="inline mr-1" /> Localização</label>
                <button 
                  type="button" 
                  onClick={handleLocateMe}
                  className="text-[10px] uppercase font-bold text-cyan-500 hover:text-cyan-400 flex items-center gap-1 transition-colors"
                >
                  <Navigation size={10} />
                  Usar meu GPS
                </button>
              </div>
              
              <div className="relative h-[240px] rounded-xl overflow-hidden border border-white/10 group">
                <MapContainer 
                  center={[-23.5505, -46.6333]} 
                  zoom={13} 
                  style={{ height: '100%', width: '100%', filter: 'grayscale(100%) invert(100%) contrast(90%)' }}
                  zoomControl={false}
                >
                  <TileLayer
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                  />
                  <LocationMarker position={location} setPosition={setLocation} />
                  <MapCenter position={location} />
                </MapContainer>
                
                {/* Overlay guides */}
                {!location && (
                  <div className="absolute inset-0 bg-black/40 backdrop-blur-[2px] flex flex-col items-center justify-center p-6 text-center pointer-events-none transition-opacity group-hover:opacity-0">
                    <MousePointer2 size={24} className="text-white/40 mb-2 animate-bounce" />
                    <p className="text-xs text-white/60 font-medium">Clique no mapa para marcar a posição do Edge</p>
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
                  placeholder="produção, sensores, piape..."
                  className={inputCls}
                  value={newTag}
                  onChange={e => setNewTag(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && (e.preventDefault(), addTag())}
                  disabled={actionLoading}
                />
                <Button type="button" variant="outline" className="border-white/10" onClick={addTag}>Add</Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {tags.map(t => (
                  <span key={t} className="flex items-center gap-1.5 px-2 py-1 bg-white/5 border border-white/10 rounded-lg text-xs text-zinc-300 animate-in zoom-in-90">
                    {t}
                    <button type="button" onClick={() => removeTag(t)} className="hover:text-red-400">×</button>
                  </span>
                ))}
              </div>
            </div>

            <Button
              type="submit"
              className="w-full gap-2 bg-white text-zinc-900 hover:bg-zinc-100 font-medium h-11 transition-all active:scale-95"
              disabled={actionLoading || !name || !location}
            >
              {actionLoading ? <Loader2 size={15} className="animate-spin" /> : <ArrowRight size={15} />}
              Próximo Passo: Conectar Cloud
            </Button>
          </form>
        </div>
      )}

      {/* ── CONNECTION STEP ─────────────────────────────── */}
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
                  onChange={e => setToken(e.target.value.toUpperCase())}
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
                  onClick={() => setStep('onboarding')}
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
                  onChange={e => setEmail(e.target.value)}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div>
                <label className={labelCls}><Lock size={11} className="inline mr-1" /> Senha</label>
                <input
                  type="password"
                  placeholder="••••••••"
                  className={inputCls}
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="flex gap-3 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  className="flex-1 border-white/5 text-zinc-400 hover:text-white"
                  onClick={() => setStep('onboarding')}
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
            Resetar Configurações e Voltar ao Início
          </button>
        </div>
      )}

      {/* ── CONNECTED STATE ─────────────────────────────── */}
      {step === 'connected' && status && (
        <div className="space-y-6 animate-in zoom-in-95 duration-300">
          {/* Status card */}
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
              <MqttBadge connected={status.mqtt.connected} />
            </div>

            <div className="mt-8 pt-5 border-t border-white/[0.04]">
                <p className="text-[10px] uppercase tracking-widest text-zinc-500 font-bold mb-2">Detalhes da Identidade</p>
                <div className="grid grid-cols-2 gap-4">
                    <div className="bg-white/[0.03] rounded-xl p-3 border border-white/[0.03]">
                        <p className="text-[9px] uppercase text-zinc-600 mb-1">Status de Conexão</p>
                        <p className={`text-xs font-medium ${status.mqtt.connected ? 'text-emerald-400' : 'text-zinc-500'}`}>
                            {status.mqtt.connected ? 'Ativo e Recebendo' : 'Aguardando Broker'}
                        </p>
                    </div>
                    <div className="bg-white/[0.03] rounded-xl p-3 border border-white/[0.03]">
                        <p className="text-[9px] uppercase text-zinc-600 mb-1">Protocolo</p>
                        <p className="text-xs font-medium text-zinc-400">MQTT over TLS</p>
                    </div>
                </div>
            </div>
          </div>

          {/* Actions */}
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
              "Desconectar irá interromper o tráfego de dados, mas não apagará o registro deste dispositivo. 
              Você pode reconectar a qualquer momento usando as credenciais já armazenadas."
            </p>
          </div>
        </div>
      )}
    </main>
  );
}
