import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Code, Package, Play, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import { useScriptsStore } from '@/stores/scripts-store';

function ScriptCard({
  id,
  title,
  subtitle,
  info,
  badges,
  onPlay,
  onDelete,
}: {
  id?: string;
  title: string;
  subtitle: string;
  info: string;
  badges: string[];
  onPlay: () => void;
  onDelete?: () => void;
}) {
  const navigate = useNavigate();
  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      onClick={() => id && navigate(`/script-hub/${id}`)}
      className="bg-white/[0.03] hover:bg-white/[0.06] border border-white/[0.08] hover:border-white/[0.15] rounded-xl p-5 transition-all duration-300 flex flex-col cursor-pointer group"
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500/20 to-blue-500/20 border border-white/[0.08] flex items-center justify-center group-hover:scale-105 transition-transform">
            <Code size={18} className="text-violet-400" />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white group-hover:text-violet-400 transition-colors">{title}</h3>
            <p className="text-[11px] text-zinc-500">{subtitle}</p>
          </div>
        </div>
      </div>
      <p className="text-xs text-zinc-500 mb-4 line-clamp-2 flex-1">{info}</p>
      {badges.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-4">
          {badges.map((tag) => (
            <span key={tag} className="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-medium bg-white/[0.06] text-zinc-400 border border-white/[0.06]">
              {tag}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
        <Button
          size="sm"
          variant="outline"
          onClick={(event) => {
            event.stopPropagation();
            onPlay();
          }}
          className="gap-1.5 text-xs h-8 border-violet-500/20 text-violet-400 hover:bg-violet-500/10"
        >
          <Play size={12} /> Playground
        </Button>
        {onDelete && (
          <Button
            size="sm"
            variant="ghost"
            onClick={(event) => {
              event.stopPropagation();
              onDelete();
            }}
            className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs"
          >
            Excluir
          </Button>
        )}
      </div>
    </motion.div>
  );
}

export default function ScriptHubPage() {
  const { scripts, samples, loading, fetchAll, fetchSamples, deleteScript } = useScriptsStore();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [activeTab, setActiveTab] = useState<'my-scripts' | 'samples'>('my-scripts');

  useEffect(() => {
    fetchAll();
    fetchSamples();
  }, [fetchAll, fetchSamples]);

  const filteredScripts = scripts.filter((script) => (
    !search ||
    script.name.toLowerCase().includes(search.toLowerCase()) ||
    script.tags?.some((tag: string) => tag.toLowerCase().includes(search.toLowerCase()))
  ));
  const filteredSamples = samples.filter((sample) => !search || sample.toLowerCase().includes(search.toLowerCase()));

  return (
    <main className="grow px-8 py-6 w-full max-w-[1200px] mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500/20 to-blue-500/20 border border-white/[0.08] flex items-center justify-center">
              <Package size={20} className="text-violet-400" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-white tracking-tight">Script Hub</h1>
              <p className="text-sm text-zinc-500">Desenvolva, teste e instale seus scripts.</p>
            </div>
          </div>
        </div>
        <Button onClick={() => navigate('/script-hub/new')} className="bg-violet-600 hover:bg-violet-700 text-white gap-2">
          Novo Script
        </Button>
      </div>

      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" size={14} />
          <input
            type="text"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Buscar scripts..."
            className="w-full pl-9 pr-3 py-2.5 bg-white/[0.04] border border-white/[0.1] rounded-lg text-sm text-white placeholder:text-zinc-600 focus:outline-none focus:border-white/[0.2] transition-colors"
          />
        </div>
      </div>

      <div className="mb-6 inline-flex h-9 items-center justify-center p-1 text-zinc-400">
        <button
          type="button"
          onClick={() => setActiveTab('my-scripts')}
          className={`px-3 py-1 text-sm transition-all ${activeTab === 'my-scripts' ? 'text-white' : 'hover:text-white'}`}
        >
          Instalados ({scripts.length})
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('samples')}
          className={`px-3 py-1 text-sm transition-all ${activeTab === 'samples' ? 'text-white' : 'hover:text-white'}`}
        >
          Exemplos Nativos ({samples.length})
        </button>
      </div>

      {activeTab === 'my-scripts' && (
        loading && scripts.length === 0 ? (
          <div className="flex items-center justify-center py-20">
            <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        ) : filteredScripts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center bg-white/[0.02] border border-white/[0.08] rounded-xl border-dashed">
            <h3 className="text-sm font-medium text-white mb-1">Nenhum script instalado</h3>
            <p className="text-xs text-zinc-500 max-w-sm mb-4">Crie um script pelo editor ou importe um pacote ZIP.</p>
            <Button variant="outline" onClick={() => navigate('/script-hub/new')} className="border-white/[0.1] bg-white/[0.05] hover:bg-white/[0.1] text-white">Criar Script</Button>
          </div>
        ) : (
          <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <AnimatePresence mode="popLayout">
              {filteredScripts.map((script) => (
                <ScriptCard
                  key={script.id}
                  id={script.id}
                  title={script.name}
                  subtitle={`by ${script.author} v${script.version}`}
                  info={script.description || 'Sem descrição'}
                  badges={script.tags || []}
                  onPlay={() => navigate(`/script-hub/${script.id}/playground`)}
                  onDelete={() => deleteScript(script.id)}
                />
              ))}
            </AnimatePresence>
          </motion.div>
        )
      )}

      {activeTab === 'samples' && (
        loading && samples.length === 0 ? (
          <div className="flex items-center justify-center py-20">
            <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        ) : filteredSamples.length === 0 ? (
          <div className="flex w-full h-32 items-center justify-center text-sm text-zinc-500">Nenhum exemplo.</div>
        ) : (
          <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            <AnimatePresence mode="popLayout">
              {filteredSamples.map((sample) => (
                <ScriptCard
                  key={sample}
                  title={sample}
                  subtitle="Biblioteca Nativa"
                  info="Este é um script de exemplo empacotado nativamente."
                  badges={['python']}
                  onPlay={() => navigate(`/script-hub/playground/sample/${encodeURIComponent(sample)}`)}
                />
              ))}
            </AnimatePresence>
          </motion.div>
        )
      )}
    </main>
  );
}
