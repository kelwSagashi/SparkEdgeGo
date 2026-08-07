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
  const [dbFile, setDbFile] = useState('');
  const [jwtSecret, setJwtSecret] = useState('');
  const [serverPort, setServerPort] = useState('');

  const normalizePortValue = (value?: string | number | null) =>
    String(value ?? '').replace(/^:/, '');

  const loadConfig = async () => {
    setLoading(true);
    try {
      const res = await cloudService.getConfig();
      const cfg = res.data.data;
      setConfig(cfg);
      setCloudUrl(cfg.cloud.url);
      setMqttUrl(cfg.cloud.mqtt_url);
      setDbFile(cfg.db.file);
      setJwtSecret('');
      setServerPort(normalizePortValue(cfg.server.port));
    } catch (err: any) {
      toast.error(`Não foi possível carregar as configurações: ${err?.message ?? 'Erro desconhecido'}`);
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

      if (cloudUrl !== config?.cloud.url || mqttUrl !== config?.cloud.mqtt_url) {
        updates.cloud = {};
        if (cloudUrl !== config?.cloud.url) updates.cloud.url = cloudUrl;
        if (mqttUrl !== config?.cloud.mqtt_url) updates.cloud.mqtt_url = mqttUrl;
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

      if (Object.keys(updates).length === 0) {
        toast.info('Nenhuma alteração detectada.');
        return;
      }

      const res = await cloudService.updateConfig(updates);
      toast.success(res.data.data.message || 'Configurações salvas. Reinicie o serviço para aplicar as mudanças.');
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
          <span className="text-sm">Carregando configurações...</span>
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
          <h1 className="text-2xl font-semibold text-white tracking-tight">Configurações Avançadas</h1>
        </div>
        <p className="text-sm text-zinc-500 mt-1 ml-11">
          Configurações de infraestrutura do SparkEdge.
        </p>
      </div>

      <div className="flex items-start gap-3 bg-amber-500/[0.08] border border-amber-500/30 rounded-2xl px-5 py-4 mb-8 animate-in slide-in-from-top-2">
        <AlertTriangle size={18} className="text-amber-400 shrink-0 mt-0.5" />
        <div>
          <p className="text-sm font-semibold text-amber-300 mb-1">Atenção: Configurações de Infraestrutura</p>
          <p className="text-xs text-amber-400/80 leading-relaxed">
            Alterar essas configurações pode interromper a conexão com o Spark Cloud e o funcionamento do SparkEdge.
            Após salvar, é necessário reiniciar o serviço para aplicar as mudanças.
            As configurações são salvas no arquivo <code className="bg-amber-500/10 px-1 rounded font-mono text-amber-300">spark-edge.config.yml</code>.
          </p>
        </div>
      </div>

      <div className="flex items-center gap-2 bg-white/[0.02] border border-white/[0.06] rounded-xl px-4 py-3 mb-6">
        <Info size={13} className="text-zinc-500 shrink-0" />
        <p className="text-xs text-zinc-500 font-mono flex-1 truncate">
          Arquivo: <span className="text-zinc-400">{config?.config_file ?? 'spark-edge.config.yml'}</span>
          <span className="mx-2 text-zinc-700">·</span>
          Prioridade: <span className="text-zinc-400">yml → .env → padrões</span>
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
            description="URLs de conexão com o Spark Cloud e broker MQTT"
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
            title="Autenticação"
            description="Chave secreta para assinar tokens JWT locais"
          />
          <div>
            <label className={labelCls}>
              JWT Secret
              {config?.auth.is_default && (
                <span className="ml-2 normal-case text-[10px] font-bold text-amber-400 bg-amber-500/10 px-1.5 py-0.5 rounded-full tracking-normal border border-amber-500/20">
                  Usando valor padrão
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
              Mínimo de 8 caracteres. Deixe em branco para manter o valor atual.
            </p>
          </div>
        </div>

        <div className={sectionCls}>
          <SectionHeader
            icon={<Server size={16} className="text-emerald-400" />}
            title="Servidor"
            description="Configurações do servidor HTTP local"
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

        <div className="pt-2">
          {saved && (
            <div className="flex items-center gap-2 mb-4 text-emerald-400 text-sm animate-in fade-in">
              <CheckCircle2 size={16} />
              <span>Configurações salvas. Reinicie o serviço para aplicar.</span>
            </div>
          )}

          <Button
            type="submit"
            disabled={saving}
            className="w-full h-12 gap-2 bg-amber-500 text-zinc-900 hover:bg-amber-400 font-semibold transition-all active:scale-[0.98] shadow-lg shadow-amber-500/20"
          >
            {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
            {saving ? 'Salvando...' : 'Salvar Configurações'}
          </Button>
        </div>
      </form>
    </main>
  );
}
