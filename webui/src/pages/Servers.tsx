import { useEffect, useState } from 'react';
import { Plus, Server, Globe, ChevronDown, ClipboardPaste, Cloud } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { useServerssStore } from '@/stores/servers-store';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { SparkCloudServerDialog } from '@/components/servers/spark-cloud-server-dialog';
export default function ServersPage() {
  const navigate = useNavigate();
  const { servers, loading, fetchAll, deleteServer } = useServerssStore();
  const [sparkDialogOpen, setSparkDialogOpen] = useState(false);
  const [importMode, setImportMode] = useState(false);

  useEffect(() => { fetchAll(); }, []);

  return (
    <main className="grow px-8 py-6 w-full max-w-[1400px] mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Servidores</h1>
          <p className="text-sm text-zinc-500 mt-1">Gerencie os servidores de destino dos dados.</p>
        </div>
        <DropdownMenu><DropdownMenuTrigger asChild><Button className="bg-violet-600 hover:bg-violet-700 text-white gap-2">Novo Servidor <ChevronDown size={14} /></Button></DropdownMenuTrigger><DropdownMenuContent align="end" className="border-white/10 bg-[#1b1b1f] text-white"><DropdownMenuItem onSelect={() => navigate('/servers/new')} className="gap-2 focus:bg-white/10 focus:text-white"><Plus size={15} />Adicionar servidor</DropdownMenuItem><DropdownMenuItem onSelect={() => { setImportMode(false); setSparkDialogOpen(true); }} className="gap-2 focus:bg-white/10 focus:text-white"><Cloud size={15} />Adicionar Spark</DropdownMenuItem><DropdownMenuItem onSelect={() => { setImportMode(true); setSparkDialogOpen(true); }} className="gap-2 focus:bg-white/10 focus:text-white"><ClipboardPaste size={15} />Importar curl</DropdownMenuItem></DropdownMenuContent></DropdownMenu>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
      ) : servers.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-16 h-16 rounded-2xl bg-white/[0.04] border border-white/[0.08] flex items-center justify-center mb-4">
            <Server size={24} className="text-zinc-600" />
          </div>
          <h3 className="text-sm font-medium text-white mb-1">Nenhum servidor cadastrado</h3>
          <p className="text-xs text-zinc-500 mb-4">Adicione um servidor para enviar dados das instâncias.</p>
          <Button onClick={() => navigate('/servers/new')} variant="outline" size="sm" className="gap-2 border-white/10 text-white hover:bg-white/5">
            <Plus size={14} /> Adicionar
          </Button>
        </div>
      ) : (
        <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          <AnimatePresence mode="popLayout">
            {servers.map(server => (
              <motion.div
                key={server.id}
                layout
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -12 }}
                className="group bg-foreground hover:bg-muted-foreground border border-border rounded-xl p-5 transition-all duration-300"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-orange-500/10 text-orange-400">
                      <Server size={16} />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold text-white">{server.name}</h3>
                      <span className="text-[10px] uppercase tracking-wider text-zinc-500">{server.type}</span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-[11px] text-zinc-500 mt-2">
                  <Globe size={11} />
                  <span className="truncate">{server.type}</span>
                </div>

                <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
                  <Button size="sm" variant="ghost" onClick={() => navigate(`/servers/${server.id}`)} className="h-8 text-xs text-secondary hover:text-primary">Editar</Button>
                  <Button size="sm" variant="ghost" onClick={() => deleteServer(server.id)} className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs">Excluir</Button>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      )}
      <SparkCloudServerDialog open={sparkDialogOpen} onOpenChange={setSparkDialogOpen} onCreated={fetchAll} initialMode={importMode ? 'curl' : 'spark'} />
    </main>
  );
}

