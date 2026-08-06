import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ShieldAlert, Check, X, Loader2 } from 'lucide-react';
import { Textarea } from '@/components/ui/textarea';
import { useCredentialsStore } from '@/stores/credentials-store';
import { api } from '@/server/server.service';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/auth-store';

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  credentialId?: string | null;
};

export function AddCredentialDialog({ open, onOpenChange, credentialId }: Props) {
  const { user } = useAuthStore();
  const { credentials, metadata, authTypes, fetchMetadata, createCredential, updateCredential } = useCredentialsStore();
  
  const [name, setName] = useState('');
  const [authType, setAuthType] = useState<string>('');
  const [data, setData] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState(false);
  
  // Connection Test States
  const [testUrl, setTestUrl] = useState('');
  const [testLoading, setTestLoading] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);

  useEffect(() => {
    if (open) {
      if (!metadata) {
        fetchMetadata();
      }
      setTestResult(null);
      setTestUrl('');
      if (credentialId) {
        const cred = credentials.find(c => c.id === credentialId);
        if (cred) {
          setName(cred.name);
          setAuthType(cred.auth_type_id);
          setData(cred.data || {});
        }
      } else {
        setName('');
        setAuthType(authTypes[0]?.id || 'api_key');
        setData({});
      }
    }
  }, [open, credentialId, credentials, metadata, authTypes, fetchMetadata]);

  const handleSave = async () => {
    setLoading(true);
    try {
      if (credentialId) {
        await updateCredential(credentialId, { name, auth_type_id: authType, data, owner_id: user?.id });
      } else {
        await createCredential({ name, auth_type_id: authType, data, owner_id: user?.id });
      }
      onOpenChange(false);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    setTestLoading(true);
    setTestResult(null);
    try {
      console.log("test url", authType);
      const res = await api.testCredential({
        auth_type_id: authType,
        data: { ...data, testUrl }
      });

      console.log(res)

      if (res.data.success) {
        setTestResult({ success: true, message: 'Conectado com sucesso!' });
      } else {
        setTestResult({ success: false, message: res.data.error || 'Erro desconhecido' });
      }
    } catch (e: any) {
      console.log(e)
      setTestResult({ success: false, message: e.message });
    } finally {
      setTestLoading(false);
    }
  };

  const handleDataChange = (key: string, value: string) => {
    setData((prev) => ({ ...prev, [key]: value }));
  };

  const currentFields = (metadata?.[authType] || []) as any[];
  const isHttpBased = ['api_key', 'bearer_token', 'basic_auth', 'supabase', 'no_auth', 'digest_auth'].includes(authType);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px] bg-[#09090b] border-white/[0.08] text-white">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldAlert size={18} className="text-violet-400" />
            {credentialId ? 'Editar Credencial' : 'Nova Credencial'}
          </DialogTitle>
        </DialogHeader>

        <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto pr-2">
          <div className="flex flex-col gap-2">
            <Label>Nome da Credencial</Label>
            <Input 
              value={name} 
              onChange={e => setName(e.target.value)} 
              placeholder="e.g. My Database"
              className="bg-white/[0.04] border-white/[0.1] h-9"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label>Tipo de Autenticação / Plataforma</Label>
            <Select value={authType} onValueChange={t => { setAuthType(t); setData({}); setTestResult(null); }}>
              <SelectTrigger className="bg-white/[0.04] border-white/[0.1] h-9 text-zinc-300">
                <SelectValue placeholder="Selecione o tipo..." />
              </SelectTrigger>
              <SelectContent className="bg-[#18181b] border-white/[0.1] text-white">
                {authTypes.map(item => (
                  <SelectItem key={item.id} value={item.id}>{item.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Dynamic Fields */}
          <div className="grid grid-cols-2 gap-4">
            {currentFields.map((field) => (
              <div key={field.key} className={cn("flex flex-col gap-2", field.grid || "col-span-2")}>
                <Label>{field.label}</Label>
                {field.type === 'textarea' ? (
                  <Textarea 
                    value={data[field.key] || ''} 
                    onChange={e => handleDataChange(field.key, e.target.value)} 
                    placeholder={field.placeholder}
                    className="bg-white/[0.04] border-white/[0.1] h-32 font-mono text-xs"
                  />
                ) : (
                  <Input 
                    type={field.type}
                    value={data[field.key] || ''} 
                    onChange={e => handleDataChange(field.key, e.target.value)} 
                    placeholder={field.placeholder}
                    className="bg-white/[0.04] border-white/[0.1] h-9"
                  />
                )}
              </div>
            ))}
          </div>

          {/* Test URL Field (Conditional) */}
          {isHttpBased && (
             <div className="mt-4 pt-4 border-t border-white/[0.05]">
                <Label className="text-zinc-500 text-[10px] uppercase tracking-wider mb-2 block">Endpoint para Teste</Label>
                <div className="flex gap-2">
                  <Input 
                    value={testUrl} 
                    onChange={e => setTestUrl(e.target.value)} 
                    placeholder="https://sua-api.com/status" 
                    className="bg-white/[0.02] border-white/[0.08] h-8 text-xs flex-1"
                  />
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={handleTest} 
                    disabled={testLoading}
                    className="border-white/[0.1] bg-white/[0.05] hover:bg-white/[0.1] text-zinc-300 text-[10px] h-8 px-3"
                  >
                    {testLoading ? <Loader2 size={12} className="animate-spin" /> : 'Testar'}
                  </Button>
                </div>
             </div>
          )}

          {(!isHttpBased) && (
            <div className="mt-4 pt-4 border-t border-white/[0.05]">
               <Button 
                  variant="outline" 
                  size="sm"
                  onClick={handleTest} 
                  disabled={testLoading}
                  className="w-full border-white/[0.1] bg-white/[0.05] hover:bg-white/[0.1] text-zinc-300 text-xs h-9 px-4"
                >
                  {testLoading ? <Loader2 size={14} className="animate-spin" /> : 'Testar Conexão'}
                </Button>
            </div>
          )}

          {testResult && (
            <div className={cn(
               "mt-2 p-3 rounded-lg text-xs flex items-center gap-2",
               testResult.success ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" : "bg-red-500/10 text-red-400 border border-red-500/20"
            )}>
               {testResult.success ? <Check size={14} /> : <X size={14} />}
               <span className="flex-1">{testResult.message}</span>
            </div>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="ghost" className="text-zinc-400 hover:text-white h-9" onClick={() => onOpenChange(false)}>Cancelar</Button>
          <Button onClick={handleSave} disabled={loading || !name} className="bg-violet-600 hover:bg-violet-700 text-white h-9">
            {loading ? 'Salvando...' : 'Salvar'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
