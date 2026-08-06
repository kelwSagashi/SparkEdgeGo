import ServerStepForm from "@/components/add-server/server-form";
import { ArrowLeft } from "lucide-react";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";

export default function ServerEditPage() {
    const navigate = useNavigate();
    const { id } = useParams();

    if (!id) return (
        <div className="h-full flex items-center justify-center">
            <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
    );

    return (
        <main className="grow px-8 py-6 w-full mx-auto pb-24">
            <div className="flex items-center gap-3 mb-8 shrink-0">
                <button onClick={() => navigate('/servers')} className="p-2 rounded-lg hover:bg-white/[0.05] text-zinc-400 hover:text-white transition-colors">
                    <ArrowLeft size={18} />
                </button>
                <div>
                    <h1 className="text-xl font-semibold text-white tracking-tight">Editar Servidor</h1>
                    <p className="text-sm text-zinc-500">Atualize as configurações do servidor e seus endpoints.</p>
                </div>
            </div>
            <ServerStepForm serverId={id} onClose={() => navigate('/servers')} />
        </main>
    );
}

