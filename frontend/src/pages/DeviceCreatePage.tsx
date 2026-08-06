import DeviceForm from "@/components/device-dialog/device-form";
import { ArrowLeft } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function DeviceCreationPage() {
    const navigate = useNavigate();
    const [saving, setSaving] = useState(false);

    return (
        <main className="grow px-8 py-6 w-full mx-auto pb-24 h-full overflow-hidden">
              {/* Header */}
              <div className="flex items-center gap-3 mb-8">
                <button onClick={() => navigate('/devices')} className="p-2 rounded-lg hover:bg-white/[0.05] text-zinc-400 hover:text-white transition-colors">
                  <ArrowLeft size={18} />
                </button>
                <div>
                  <h1 className="text-xl font-semibold text-white tracking-tight">Novo Dispositivo</h1>
                  <p className="text-sm text-zinc-500">Adicione um novo dispositivo para monitoramento.</p>
                </div>
            </div>
            <DeviceForm setSaving={setSaving} />
        </main>
    );
}
