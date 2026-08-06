import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { 
  Package, User, Tag, Clock, ChevronLeft, Play, Edit, 
  Download, Globe, FileText, Code2, AlertTriangle, CheckCircle2 
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import ReactMarkdown from 'react-markdown';
import { scriptsApi } from '@/rest-api-client/scripts.service';
import type { DownloadedScriptReturningValues } from "@/types/db";

export default function ScriptDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [script, setScript] = useState<DownloadedScriptReturningValues | null>(null);
  const [readme, setReadme] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchScriptData = async () => {
      if (!id) return;
      setLoading(true);
      try {
        const scriptRes: any = await scriptsApi.getById(id);
        if (scriptRes.data) {
          setScript(scriptRes.data);
          
          // Try to fetch README.md
          try {
            const readmeRes: any = await scriptsApi.getFileContent(id, 'README.md');
            if (readmeRes.data) {
              setReadme(readmeRes.data);
            }
          } catch (e) {
            console.log('README.md not found for this script');
          }
        } else {
          setError('Script não encontrado');
        }
      } catch (err: any) {
        setError(err.message || 'Erro ao carregar script');
      } finally {
        setLoading(false);
      }
    };

    fetchScriptData();
  }, [id]);

  if (loading) {
    return (
      <div className="p-8 space-y-6 max-w-6xl mx-auto">
        <Skeleton className="h-10 w-48 mb-6" />
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          <div className="lg:col-span-3 space-y-4">
            <Skeleton className="h-12 w-3/4" />
            <Skeleton className="h-[400px] w-full" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-64 w-full" />
          </div>
        </div>
      </div>
    );
  }

  if (error || !script) {
    return (
      <div className="p-20 text-center space-y-4">
        <AlertTriangle className="w-12 h-12 text-amber-500 mx-auto" />
        <h2 className="text-xl font-semibold text-white">{error || 'Algo deu errado'}</h2>
        <Button onClick={() => navigate('/script-hub')}>Voltar ao Hub</Button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-400 pb-20">
      {/* Top Header */}
      <div className="border-b border-white/[0.06] bg-white/[0.02]">
        <div className="max-w-6xl mx-auto px-6 py-6">
          <Link 
            to="/script-hub" 
            className="inline-flex items-center gap-2 text-xs text-zinc-500 hover:text-white transition-colors mb-6 group"
          >
            <ChevronLeft size={14} className="group-hover:-translate-x-0.5 transition-transform" />
            Voltar para Script Hub
          </Link>

          <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
            <div className="flex items-start gap-5">
              <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-violet-600 to-blue-600 p-[1px]">
                 <div className="w-full h-full rounded-2xl bg-[#09090b] flex items-center justify-center">
                    <Package size={32} className="text-white" />
                 </div>
              </div>
              <div className="space-y-1">
                <h1 className="text-3xl font-bold text-white tracking-tight">{script.name}</h1>
                <div className="flex items-center gap-4 text-sm">
                  <span className="flex items-center gap-1.5">
                    <User size={14} /> {script.author}
                  </span>
                  <span className="text-zinc-600">•</span>
                  <span className="flex items-center gap-1.5">
                    <Clock size={14} /> v{script.version}
                  </span>
                  {script.source === 'local' && (
                    <Badge variant="outline" className="bg-green-500/10 text-green-400 border-green-500/20 gap-1 capitalize">
                      <CheckCircle2 size={10} /> {script.source}
                    </Badge>
                  )}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <Button 
                variant="outline" 
                asChild
                className="border-white/[0.1] bg-white/[0.02] hover:bg-white/[0.06] text-white"
              >
                <Link to={`/script-hub/${script.id}/edit`}>
                   <Edit className="w-4 h-4 mr-2" /> Editar
                </Link>
              </Button>
              <Button className="bg-violet-600 hover:bg-violet-700 text-white shadow-lg shadow-violet-600/20">
                <Play className="w-4 h-4 mr-2" /> Playground
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-6xl mx-auto px-6 pt-10">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-10">
          {/* Main Content: README */}
          <div className="lg:col-span-3 space-y-8">
            <div className="bg-white/[0.02] border border-white/[0.06] rounded-2xl p-8 min-h-[500px]">
              {readme ? (
                 <article className="prose prose-invert prose-zinc max-w-none 
                    prose-h1:text-white prose-h2:text-white prose-h3:text-zinc-100
                    prose-p:text-zinc-400 prose-strong:text-zinc-200
                    prose-code:text-violet-300 prose-pre:bg-black/50 prose-pre:border prose-pre:border-white/10">
                    <ReactMarkdown>{readme}</ReactMarkdown>
                 </article>
              ) : (
                <div className="flex flex-col items-center justify-center py-20 text-center space-y-4">
                  <FileText className="w-12 h-12 text-zinc-700" />
                  <div>
                    <h3 className="text-lg font-medium text-zinc-300">Sem documentação</h3>
                    <p className="text-sm text-zinc-500">Este script não possui um arquivo README.md.</p>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Sidebar: Metadata & Info */}
          <div className="space-y-6">
            <Card className="bg-white/[0.02] border border-white/[0.06] overflow-hidden">
               <CardContent className="p-6 space-y-6">
                  <div>
                    <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-4">Informações</h4>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between text-sm">
                        <span className="flex items-center gap-2"><Globe size={14} /> Source</span>
                        <span className="text-white capitalize">{script.source}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="flex items-center gap-2"><Code2 size={14} /> Entrypoint</span>
                        <span className="text-white font-mono text-[10px]">{script.main_file}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="flex items-center gap-2"><Tag size={14} /> Tags</span>
                        <div className="flex flex-wrap justify-end gap-1">
                          {script.tags?.map(tag => (
                            <Badge key={tag} variant="secondary" className="text-[10px] py-0 px-1.5 bg-white/[0.05] text-zinc-400 border-none">
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>

                  <Separator className="bg-white/[0.06]" />

                  <div>
                    <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">Descrição curta</h4>
                    <p className="text-sm text-zinc-400 leading-relaxed">
                      {script.description || 'Nenhuma descrição fornecida.'}
                    </p>
                  </div>

                  <div className="pt-4">
                    <Button variant="ghost" className="w-full text-zinc-500 hover:text-white hover:bg-white/[0.05] text-xs">
                       <Download className="w-3 h-3 mr-2" /> Download Bundle
                    </Button>
                  </div>
               </CardContent>
            </Card>

            <Card className="bg-gradient-to-br from-violet-600/10 to-blue-600/10 border border-violet-500/20">
               <CardContent className="p-6">
                  <h4 className="text-sm font-semibold text-white mb-2">Instalação via CLI</h4>
                  <div className="relative group">
                    <pre className="bg-black/40 p-3 rounded-lg text-[10px] font-mono text-zinc-300 overflow-x-auto border border-white/5">
                      spark-edge install {script.id}
                    </pre>
                  </div>
               </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

