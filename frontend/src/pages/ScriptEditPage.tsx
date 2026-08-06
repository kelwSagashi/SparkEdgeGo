import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { 
  Save, ChevronLeft, Package, User, Tag, FileText, Bookmark, Info, Loader2
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { scriptsApi } from '@/rest-api-client/scripts.service';
import type { DownloadedScriptReturningValues } from "@/types/db";

export default function ScriptEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [script, setScript] = useState<DownloadedScriptReturningValues | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form states
  const [formData, setFormData] = useState({
    name: '',
    version: '',
    author: '',
    description: '',
    tags: [] as string[]
  });

  useEffect(() => {
    const fetchScript = async () => {
      if (!id) return;
      try {
        const res: any = await scriptsApi.getById(id);
        if (res.data) {
          setScript(res.data);
          setFormData({
            name: res.data.name || '',
            version: res.data.version || '',
            author: res.data.author || '',
            description: res.data.description || '',
            tags: res.data.tags || []
          });
        }
      } catch (err: any) {
        setError(err.message || 'Erro ao carregar script');
      } finally {
        setLoading(false);
      }
    };
    fetchScript();
  }, [id]);

  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await scriptsApi.update(id, formData);
      navigate(`/script-hub/${id}`);
    } catch (err: any) {
      setError(err.message || 'Erro ao salvar alterações');
    } finally {
      setSaving(false);
    }
  };

  const handleTagsChange = (val: string) => {
    setFormData(prev => ({
      ...prev,
      tags: val.split(',').map(t => t.trim()).filter(Boolean)
    }));
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-violet-500" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-400 pb-20">
      <div className="max-w-4xl mx-auto px-6 py-10">
        <Link 
          to={script ? `/script-hub/${script.id}` : "/script-hub"} 
          className="inline-flex items-center gap-2 text-xs text-zinc-500 hover:text-white transition-colors mb-8 group"
        >
          <ChevronLeft size={14} className="group-hover:-translate-x-0.5 transition-transform" />
          Voltar {script ? `para ${script.name}` : ''}
        </Link>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="md:col-span-2 space-y-6">
            <h1 className="text-3xl font-bold text-white tracking-tight">Editar Script</h1>
            
            <Card className="bg-white/[0.02] border-white/[0.06]">
              <CardHeader>
                <CardTitle className="text-white text-lg flex items-center gap-2">
                   <Bookmark className="w-4 h-4 text-violet-400" /> Identificação
                </CardTitle>
                <CardDescription className="text-zinc-500">Ajuste o nome e versão do script.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-2">
                  <Label className="text-zinc-400 flex items-center gap-2"><Package size={14} /> Nome do Script</Label>
                  <Input 
                    value={formData.name} 
                    onChange={e => setFormData(p => ({ ...p, name: e.target.value }))}
                    className="bg-white/[0.04] border-white/[0.1] text-white focus:border-violet-500/50" 
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label className="text-zinc-400 flex items-center gap-2"><User size={14} /> Autor</Label>
                    <Input 
                       value={formData.author} 
                       onChange={e => setFormData(p => ({ ...p, author: e.target.value }))}
                       className="bg-white/[0.04] border-white/[0.1] text-white" 
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label className="text-zinc-400 flex items-center gap-2">Versão</Label>
                    <Input 
                       value={formData.version} 
                       onChange={e => setFormData(p => ({ ...p, version: e.target.value }))}
                       className="bg-white/[0.04] border-white/[0.1] text-white" 
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="bg-white/[0.02] border-white/[0.06]">
              <CardHeader>
                <CardTitle className="text-white text-lg flex items-center gap-2">
                   <FileText className="w-4 h-4 text-violet-400" /> Descrição e Tags
                </CardTitle>
                <CardDescription className="text-zinc-500">Metadados para facilitar a busca e organização.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-2">
                  <Label className="text-zinc-400">Descrição Curta</Label>
                  <Textarea 
                    value={formData.description} 
                    onChange={e => setFormData(p => ({ ...p, description: e.target.value }))}
                    className="bg-white/[0.04] border-white/[0.1] text-white min-h-[100px]" 
                  />
                </div>
                <div className="grid gap-2">
                  <Label className="text-zinc-400 flex items-center gap-2"><Tag size={14} /> Tags (separadas por vírgula)</Label>
                  <Input 
                    value={formData.tags.join(', ')} 
                    onChange={e => handleTagsChange(e.target.value)}
                    placeholder="tag1, tag2, research..."
                    className="bg-white/[0.04] border-white/[0.1] text-white" 
                  />
                </div>
              </CardContent>
            </Card>

            <div className="flex justify-end gap-3">
               <Button variant="outline" asChild className="border-white/10 text-zinc-400">
                  <Link to={`/script-hub/${id}`}>Cancelar</Link>
               </Button>
               <Button 
                onClick={handleSave} 
                disabled={saving}
                className="bg-violet-600 hover:bg-violet-700 text-white min-w-[120px]"
               >
                 {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <><Save className="w-4 h-4 mr-2" /> Salvar Alterações</>}
               </Button>
            </div>
            {error && <p className="text-red-400 text-sm mt-2">{error}</p>}
          </div>

          <div className="space-y-6 text-zinc-500">
            <div className="bg-white/[0.02] rounded-xl p-6 border border-white/[0.06] space-y-4">
               <div className="flex items-center gap-2 text-white font-medium">
                  <Info size={16} className="text-violet-400" />
                  Dica
               </div>
               <p className="text-xs leading-relaxed">
                 O arquivo README.md do script é usado automaticamente para a documentação na página de visualização. 
                 Para alterar o conteúdo detalhado, edite o arquivo no bundle do script.
               </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

