import DeviceForm from "@/components/device-dialog/device-form";
import { ArrowLeft, Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "@/server/server.service";

export default function DeviceEditPage() {
    const { id } = useParams();
    const navigate = useNavigate();
    const [saving, setSaving] = useState(false);
    const [loading, setLoading] = useState(true);
    const [device, setDevice] = useState<any>(null);

    useEffect(() => {
        const fetchDevice = async () => {
            if (!id) return;
            try {
                const res = await api.getDeviceById(id);
                if (res.data?.data) {
                    setDevice(res.data.data);
                } else {
                    alert('Dispositivo não encontrado');
                    navigate('/devices');
                }
            } catch (err) {
                console.error('Failed to fetch device', err);
                alert('Erro ao carregar dispositivo');
                navigate('/devices');
            } finally {
                setLoading(false);
            }
        };
        fetchDevice();
    }, [id, navigate]);

    return (
        <main className="grow px-8 py-6 w-full mx-auto pb-24 h-full overflow-hidden">
            {/* Header */}
            <div className="flex items-center gap-3 mb-8">
                <button onClick={() => navigate('/devices')} className="p-2 rounded-lg hover:bg-white/[0.05] text-zinc-400 hover:text-white transition-colors">
                    <ArrowLeft size={18} />
                </button>
                <div>
                    <h1 className="text-xl font-semibold text-white tracking-tight">Editar Dispositivo</h1>
                    <p className="text-sm text-zinc-500">Atualize as informações do seu dispositivo.</p>
                </div>
            </div>

            {loading ? (
                <div className="flex-1 flex items-center justify-center py-20">
                    <Loader2 className="w-8 h-8 text-white/20 animate-spin" />
                </div>
            ) : (
                <DeviceForm setSaving={setSaving} initialData={device} />
            )}
        </main>
    );
}

