import { useEffect, useState } from 'react';
import { ShieldAlert, Plus, Shield, Search, KeyRound } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import { useCredentialsStore } from '@/stores/credentials-store';
import { AddCredentialDialog } from '@/components/credentials/add-credential-dialog';

export default function CredentialsPage() {
  const { credentials, loading, fetchAll, deleteCredential } = useCredentialsStore();
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [activeCredentialId, setActiveCredentialId] = useState<string | null>(null);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const handleEdit = (id: string) => {
    setActiveCredentialId(id);
    setDialogOpen(true);
  };

  const handleNew = () => {
    setActiveCredentialId(null);
    setDialogOpen(true);
  };

  const filtered = credentials.filter(c => !search || c.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <main className="grow px-8 py-6 w-full max-w-[1400px] mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500/20 to-blue-500/20 border border-white/[0.08] flex items-center justify-center">
            <ShieldAlert size={20} className="text-violet-400" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold text-white tracking-tight">Credentials</h1>
            <p className="text-sm text-zinc-500">Gerencie chaves, senhas e conexões seguras estilo N8N.</p>
          </div>
        </div>
        <Button onClick={handleNew} className="bg-violet-600 hover:bg-violet-700 text-white gap-2">
          Nova Credencial
        </Button>
      </div>

      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={14} />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar credencial..."
            className="w-full pl-9 pr-3 py-2.5 bg-white/[0.04] border border-white/[0.1] rounded-lg text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] transition-colors"
          />
        </div>
      </div>

      {loading && credentials.length === 0 ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
      ) : credentials.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center bg-white/[0.02] border border-white/[0.08] rounded-xl border-dashed">
          <h3 className="text-sm font-medium text-white mb-1">Nenhuma credencial cadastrada</h3>
          <p className="text-xs text-zinc-500 max-w-sm mb-4">Adicione uma credencial do Supabase, Postgres, Token etc.</p>
          <Button variant="outline" onClick={handleNew} className="border-white/[0.1] bg-white/[0.05] hover:bg-white/[0.1] text-white">Criar Credencial</Button>
        </div>
      ) : (
        <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          <AnimatePresence mode="popLayout">
            {filtered.map(cred => (
              <motion.div
                key={cred.id}
                layout
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -12 }}
                className="group bg-foreground hover:bg-muted-foreground border border-border rounded-xl p-5 transition-all duration-300"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-violet-500/10 text-violet-400">
                      <KeyRound size={16} />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold text-primary">{cred.name}</h3>
                    </div>
                  </div>
                </div>
                <div className="flex flex-1 items-end justify-between pt-4 mt-2 border-t border-white/[0.06]">
                   <Button size="sm" variant="ghost" onClick={() => handleEdit(cred.id)} className="h-8 text-xs text-zinc-300 hover:text-white">Editar</Button>
                   <Button size="sm" variant="ghost" onClick={() => deleteCredential(cred.id)} className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs">Excluir</Button>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      )}

      <AddCredentialDialog 
        open={dialogOpen}
        onOpenChange={(v) => {
          setDialogOpen(v);
          if (!v) setActiveCredentialId(null);
        }}
        credentialId={activeCredentialId}
      />
    </main>
  );
}

