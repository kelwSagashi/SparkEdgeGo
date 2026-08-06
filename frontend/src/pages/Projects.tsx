import { useEffect, useState } from 'react';
import { useProjectsStore } from '@/stores/projects-store';
import { Button } from '@/components/ui/button';
import { Plus, FolderKanban, Trash2, Edit2, X, Check } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

export default function ProjectsPage() {
  const { projects, loading, fetchAll, createProject, deleteProject } = useProjectsStore();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newKey, setNewKey] = useState('');
  const [newDesc, setNewDesc] = useState('');

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const handleCreate = async () => {
    if (!newName.trim() || !newKey.trim()) return;
    await createProject({ name: newName, key: newKey.toUpperCase(), description: newDesc || undefined });
    setNewName(''); setNewKey(''); setNewDesc('');
    setShowCreate(false);
  };

  const inputClasses = "w-full px-4 py-3 bg-white/[0.04] border border-white/[0.1] rounded-xl text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] transition-all";

  return (
    <main className="grow px-8 py-6 w-full max-w-[1400px] mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Projetos</h1>
          <p className="text-sm text-zinc-500 mt-1">Organize suas instâncias em projetos.</p>
        </div>
        <Button onClick={() => setShowCreate(true)} className="bg-violet-600 hover:bg-violet-700 text-primary gap-2">
          Novo Projeto
        </Button>
      </div>

      {/* Create form */}
      <AnimatePresence>
        {showCreate && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mb-6 overflow-hidden"
          >
            <div className="bg-white/[0.03] border border-white/[0.08] rounded-xl p-5 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Nome *</label>
                  <input type="text" value={newName} onChange={e => setNewName(e.target.value)} placeholder="Meu Projeto" className={inputClasses} />
                </div>
                <div>
                  <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Tag *</label>
                  <input type="text" value={newKey} onChange={e => setNewKey(e.target.value.toUpperCase())} placeholder="PROJ" className={inputClasses} />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">Descrição</label>
                <input type="text" value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="Descrição do projeto..." className={inputClasses} />
              </div>
              <div className="flex gap-2">
                <Button onClick={handleCreate} disabled={!newName.trim() || !newKey.trim()} className="gap-2 bg-emerald-500 hover:bg-emerald-400 text-white font-medium">
                  <Check size={14} /> Criar
                </Button>
                <Button onClick={() => setShowCreate(false)} variant="ghost" className="text-zinc-400 hover:text-white">
                  <X size={14} className="mr-1" /> Cancelar
                </Button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
      ) : projects.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-16 h-16 rounded-2xl bg-white/[0.04] border border-white/[0.08] flex items-center justify-center mb-4">
            <FolderKanban size={24} className="text-zinc-600" />
          </div>
          <h3 className="text-sm font-medium text-white mb-1">Nenhum projeto criado</h3>
          <p className="text-xs text-zinc-500 mb-4">Crie um projeto para organizar suas instâncias.</p>
        </div>
      ) : (
        <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          <AnimatePresence mode="popLayout">
            {projects.map(project => (
              <motion.div
                key={project.id}
                layout
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -12 }}
                className="group bg-white/[0.03] hover:bg-white/[0.06] border border-white/[0.08] hover:border-white/[0.15] rounded-xl p-5 transition-all duration-300"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
                      <FolderKanban size={16} />
                    </div>
                    <div>
                      <h3 className="text-sm font-semibold text-white">{project.name}</h3>
                      <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono bg-white/[0.06] text-zinc-400 mt-1">{project.key}</span>
                    </div>
                  </div>
                </div>
                <p className="text-xs text-zinc-500 mt-3 h-8">{project.description ?? ""}</p>
                <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
                  <Button 
                    size="sm" variant="ghost"
                    className="h-8 text-xs text-secondary hover:text-primary"
                    onClick={() => {}}
                  >
                    Editar
                  </Button>
                  <Button 
                    size="sm" variant="ghost"
                    className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs"
                    onClick={() => deleteProject(project.id)}
                  >
                    Excluir
                  </Button>
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      )}
    </main>
  );
}

