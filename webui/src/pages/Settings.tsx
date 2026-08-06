import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import { User, Mail, Lock, Save, Key, Fingerprint, Globe, RefreshCcw, ShieldCheck, ShieldAlert, AlertTriangle, Settings2, ChevronRight } from 'lucide-react';
import { systemService } from '@/rest-api-client/system.service';
import type { SystemIdentity, MqttConfig } from '@/rest-api-client/system.service';
import { toast } from 'sonner';

export default function SettingsPage() {
  const navigate = useNavigate();
  const { user, generateNewApiKey } = useAuthStore();
  const [name, setName] = useState(user?.first_name ?? '');
  const [saving, setSaving] = useState(false);

  // System Settings State
  const [identity, setIdentity] = useState<SystemIdentity | null>(null);
  const [mqttConfig, setMqttConfig] = useState<MqttConfig | null>(null);
  const [loadingSystem, setLoadingSystem] = useState(true);
  const [resettingPk, setResettingPk] = useState(false);

  const inputClasses = "w-full px-4 py-3 bg-input border border-border rounded-xl text-sm text-primary placeholder:text-secondary focus:outline-none focus:border-primary focus:bg-input transition-all";
  const labelClasses = "block text-xs font-medium text-secondary uppercase tracking-wider mb-2";

  useEffect(() => {
    loadSystemSettings();
  }, []);

  const loadSystemSettings = async () => {
    try {
      // Identity loading removed - managed in Cloud Settings
    } finally {
      setLoadingSystem(false);
    }
  };


  return (
    <main className="grow px-8 py-6 w-full max-w-[800px] mx-auto pb-20">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold text-white tracking-tight">Configurações</h1>
        <p className="text-sm text-zinc-500 mt-1">Gerencie seu perfil e as configurações de infraestrutura do sistema.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-1 gap-6">
        <div className="space-y-6">
          {/* Profile section */}
          <div className="bg-foreground border border-border rounded-xl p-6">
            <div className="flex items-center gap-4 mb-6 pb-6 border-b border-border">
              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-fuchsia-500 flex items-center justify-center">
                <span className="text-white font-bold text-base">
                  {(user?.first_name ?? user?.email ?? '').slice(0, 2).toUpperCase()}
                </span>
              </div>
              <div>
                <h2 className="text-base font-semibold text-white">{user?.first_name ?? 'Usuário'}</h2>
                <p className="text-xs text-zinc-500">{user?.email}</p>
              </div>
            </div>

            <div className="space-y-5">
              <div>
                <label className={labelClasses}>
                  <User size={11} className="inline mr-1" /> Nome
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  placeholder="Seu nome"
                  className={inputClasses}
                />
              </div>

              <Button
                disabled={saving}
                className="w-full gap-2 bg-white text-zinc-900 hover:bg-zinc-200 font-medium h-11"
              >
                <Save size={14} />
                {saving ? 'Salvando...' : 'Salvar perfil'}
              </Button>
            </div>
          </div>
        </div>

        <div className="space-y-6">

          {/* API Key section */}
          <div className="bg-foreground border border-white/[0.08] rounded-xl p-6">
            <h3 className="text-sm font-semibold text-white mb-4 flex items-center gap-2">
              <Key size={14} className="text-amber-400" /> Segurança da API
            </h3>
            <div className="space-y-4">
              <div>
                <label className={labelClasses}>Chave de Acesso Local</label>
                <div className="px-4 py-3 bg-input border border-white/[0.06] rounded-xl text-xs font-mono text-zinc-500 truncate">
                  {user?.api_key}
                </div>
              </div>
              <p className="text-[10px] text-zinc-600 leading-relaxed italic">
                * Resetar a chave de API invalidará imediatamente integrações externas que utilizem este token.
              </p>
              <Button
                onClick={() => generateNewApiKey(user?.id ?? '')}
                variant="outline" 
                className="w-full gap-2 border-white/10 text-white hover:bg-white/5 h-11"
              >
                <RefreshCcw size={14} /> Gerar Nova Chave
              </Button>
            </div>
          </div>
        </div>

        {/* Advanced Settings card */}
        <div className="space-y-6 mt-6">
          <button
            onClick={() => navigate('/settings/advanced')}
            className="w-full bg-amber-500/[0.05] border border-amber-500/20 hover:border-amber-500/40 hover:bg-amber-500/[0.08] rounded-xl p-5 text-left transition-all group"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-xl bg-amber-500/10 flex items-center justify-center shrink-0">
                  <Settings2 size={16} className="text-amber-400" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-semibold text-white">Configurações Avançadas</p>
                    <span className="inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-wider text-amber-400 bg-amber-500/10 border border-amber-500/20 px-1.5 py-0.5 rounded-full">
                      <AlertTriangle size={9} />
                      Avançado
                    </span>
                  </div>
                  <p className="text-xs text-zinc-500 mt-0.5">
                    Spark Cloud URL, MQTT broker, banco de dados, JWT secret e porta do servidor.
                  </p>
                </div>
              </div>
              <ChevronRight size={16} className="text-zinc-600 group-hover:text-amber-400 transition-colors shrink-0 ml-2" />
            </div>
          </button>
        </div>
      </div>
    </main>
  );
}
