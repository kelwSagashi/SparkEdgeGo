import { useEffect, useMemo, useState } from 'react';
import { CheckCircle2, ClipboardPaste, Cloud, Loader2, Sparkles } from 'lucide-react';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { api } from '@/server/server.service';
import { credentialsApi } from '@/rest-api-client/credentials.service';
import { useAuthStore } from '@/stores/auth-store';

type Mode = 'spark' | 'curl';
type ParsedCurl = { baseUrl: string; path: string; privateKey: string; body: Record<string, unknown> };
const defaultBaseUrl = 'https://sparkcloud-api.okelwen.site';
const defaultPath = '/telemetry/ingest/PASTE_PUBLIC_KEY';
const telemetrySchema: Record<string, unknown> = {
  timestamp: 'string', device_id: 'string', serverId: 'string', obs: 'string', data: { type: 'object', additionalProperties: true },
  pv_voltage: 'number', pv_current: 'number', pv_power: 'number', pv_energy_today: 'number',
  battery_voltage: 'number', battery_current: 'number', battery_power: 'number', battery_soc: 'number', battery_soh: 'number', battery_temp: 'number',
  ac_voltage: 'number', ac_current: 'number', ac_power: 'number', ac_freq: 'number', ac_power_factor: 'number', ac_energy_today: 'number', ac_energy_total: 'number',
  grid_voltage: 'number', grid_power: 'number', grid_energy_in: 'number', grid_energy_out: 'number', grid_freq: 'number',
  load_power: 'number', load_energy_today: 'number', load_percent: 'number', load_ac_va: 'number', device_temp: 'number',
};

function parseCurl(source: string): ParsedCurl | null {
  const endpoint = source.match(/(?:POST\s+|POST\s+['"]?)(https?:\/\/[^\s'"\\]+)/i)?.[1] || source.match(/curl(?:\s+-X\s+POST)?\s+['"]?(https?:\/\/[^\s'"\\]+)/i)?.[1];
  if (!endpoint) return null;
  const url = new URL(endpoint);
  const key = source.match(/x-spark-private-key:\s*([^'"\\\s]+)/i)?.[1] || '';
  const body = source.match(/-d\s+(['"])([\s\S]*?)\1(?:\s|$)/)?.[2];
  let parsed: Record<string, unknown> = {};
  if (body) { try { parsed = JSON.parse(body.replace(/\\'/g, "'")); } catch { /* schema falls back to telemetry defaults */ } }
  return { baseUrl: url.origin, path: url.pathname, privateKey: key, body: parsed };
}

function inputSchemaFromCurl(body: Record<string, unknown>) {
  const properties = { ...telemetrySchema } as Record<string, unknown>;
  for (const [key, value] of Object.entries(body)) {
    if (value === null || Array.isArray(value) || typeof value === 'object') continue;
    properties[key] = typeof value === 'number' ? 'number' : typeof value === 'boolean' ? 'boolean' : 'string';
  }
  return properties;
}

export function SparkCloudServerDialog({ open, onOpenChange, onCreated, initialMode = 'spark' }: { open: boolean; onOpenChange: (open: boolean) => void; onCreated: () => void; initialMode?: Mode }) {
  const { user, project } = useAuthStore();
  const [mode, setMode] = useState<Mode>('spark');
  const [baseUrl, setBaseUrl] = useState(defaultBaseUrl);
  const [path, setPath] = useState(defaultPath);
  const [privateKey, setPrivateKey] = useState('PASTE_PRIVATE_KEY');
  const [curl, setCurl] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  useEffect(() => { if (open) { setMode(initialMode); setError(''); } }, [open, initialMode]);
  const canCreate = useMemo(() => Boolean(project?.id && user?.id && baseUrl && path.startsWith('/')), [project?.id, user?.id, baseUrl, path]);
  const importCurl = () => {
    const parsed = parseCurl(curl);
    if (!parsed) { setError('Não foi possível identificar uma URL POST válida no curl.'); return; }
    setBaseUrl(parsed.baseUrl); setPath(parsed.path); setPrivateKey(parsed.privateKey || 'PASTE_PRIVATE_KEY'); setError('');
  };
  const create = async () => {
    if (!canCreate) { setError('Entre no SparkEdge e selecione um projeto antes de criar o servidor.'); return; }
    setSaving(true); setError('');
    try {
      const credentialResponse: any = await credentialsApi.create({ name: 'Spark Cloud Telemetry', auth_type_id: 'api_key', owner_id: user?.id, project_id: project?.id, data: { key: 'x-spark-private-key', value: privateKey, in: 'header' } });
      const credential = credentialResponse?.data;
      if (!credential?.id) throw new Error('A credencial não foi criada.');
      const importedBody = mode === 'curl' ? parseCurl(curl)?.body || {} : {};
      await api.registerServer({
        server: { name: 'Spark Cloud Telemetry', type: 'http', server_type_id: 'http', driver_key: 'api_key', credential_id: credential.id, headers: {}, project_id: project?.id, created_by: user?.id },
        resources: [{ resource: { name: 'Telemetry', type: 'http', config: { baseUrl }, server_id: '' }, operations: [{ name: 'Send telemetry', type: 'http', config: { method: 'POST', path }, input_schema: inputSchemaFromCurl(importedBody), output_schema: { type: 'object' } }] }],
      });
      onCreated(); onOpenChange(false);
    } catch (cause: any) { setError(cause?.response?.data?.message || cause?.message || 'Não foi possível criar o servidor Spark Cloud.'); }
    finally { setSaving(false); }
  };
  return <Dialog open={open} onOpenChange={onOpenChange}><DialogContent className="max-w-xl border-white/10 bg-[#18181b] text-white"><DialogHeader><DialogTitle className="flex items-center gap-2"><Cloud className="h-5 w-5 text-sky-300" />Adicionar Spark Cloud</DialogTitle></DialogHeader><div className="flex gap-2 border-b border-white/10 pb-3"><button onClick={() => setMode('spark')} className={`rounded px-3 py-2 text-sm ${mode === 'spark' ? 'bg-white/10 text-white' : 'text-zinc-400'}`}>Adicionar Spark</button><button onClick={() => setMode('curl')} className={`rounded px-3 py-2 text-sm ${mode === 'curl' ? 'bg-white/10 text-white' : 'text-zinc-400'}`}>Importar curl</button></div>{mode === 'curl' && <div className="space-y-3"><p className="text-sm text-zinc-400">Cole o comando disponibilizado pela unidade no Spark Cloud. A URL, o cabeçalho privado e o schema serão interpretados.</p><Textarea value={curl} onChange={(event) => setCurl(event.target.value)} className="min-h-36 border-white/10 bg-[#101112] font-mono text-xs text-zinc-200" placeholder="curl -X POST 'https://...' ..." /><Button type="button" variant="outline" onClick={importCurl} className="border-white/10"><ClipboardPaste className="mr-2 h-4 w-4" />Ler curl</Button></div>}<div className="space-y-4"><label className="block text-sm text-zinc-300">Base URL<Input value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} className="mt-2 border-white/10 bg-[#101112]" /></label><label className="block text-sm text-zinc-300">Path de telemetria<Input value={path} onChange={(event) => setPath(event.target.value)} className="mt-2 border-white/10 bg-[#101112] font-mono text-xs" /></label><label className="block text-sm text-zinc-300">Chave privada<Input type="password" value={privateKey} onChange={(event) => setPrivateKey(event.target.value)} className="mt-2 border-white/10 bg-[#101112]" /></label><div className="rounded border border-emerald-400/20 bg-emerald-400/5 p-3 text-xs text-zinc-300"><CheckCircle2 className="mr-2 inline h-4 w-4 text-emerald-300" />Cria a credencial API Key, servidor HTTP, recurso Telemetry, operação POST e schema de energia.</div>{error && <p className="text-sm text-rose-300">{error}</p>}</div><DialogFooter><Button variant="ghost" onClick={() => onOpenChange(false)}>Cancelar</Button><Button disabled={saving || !canCreate} onClick={() => void create()} className="bg-violet-600 hover:bg-violet-500"><Sparkles className="mr-2 h-4 w-4" />{saving ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Criar servidor'}</Button></DialogFooter></DialogContent></Dialog>;
}
