import { useEffect, useState } from 'react';
import { Plus, Smartphone, Wifi, Trash2, Pencil, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { api } from '@/server/server.service';
import { motion, AnimatePresence } from 'framer-motion';
import { useNavigate } from 'react-router-dom';

interface DeviceItem {
  id: string;
  name: string;
  brand: string | null;
  connection_method: string | null;
  ip_address: string | null;
  created_at: string;
}

export default function DevicesPage() {
  const navigate = useNavigate();
  const [devices, setDevices] = useState<DeviceItem[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchDevices = async () => {
    setLoading(true);
    try {
      const res = await api.listAllDevices();
      setDevices(res.data?.data ?? []);
    } finally { setLoading(false); }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Tem certeza que deseja apagar o dispositivo "${name}"?`)) return;
    try {
      const res = await api.deleteDevice(id);
      if (res.data?.data) {
        setDevices(prev => prev.filter(d => d.id !== id));
      } else {
        alert(res.data?.error || 'Falha ao apagar dispositivo');
      }
    } catch (err) {
      console.error('Failed to delete device', err);
      alert('Erro ao apagar dispositivo');
    }
  };

  useEffect(() => { fetchDevices(); }, []);

  return (
    <main className="grow px-8 py-6 w-full max-w-[1400px] mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-semibold text-white tracking-tight">Dispositivos</h1>
          <p className="text-sm text-zinc-500 mt-1">Gerencie os dispositivos monitorados pelo sistema.</p>
        </div>
        <Button onClick={() => navigate("/devices/new")} className="bg-violet-600 hover:bg-violet-700 text-white gap-2">
          Novo Dispositivo
        </Button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-white/20 border-t-white rounded-full animate-spin" />
        </div>
      ) : devices.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="w-16 h-16 rounded-2xl bg-white/[0.04] border border-white/[0.08] flex items-center justify-center mb-4">
            <Smartphone size={24} className="text-zinc-600" />
          </div>
          <h3 className="text-sm font-medium text-white mb-1">Nenhum dispositivo cadastrado</h3>
          <p className="text-xs text-zinc-500 mb-4">Adicione um dispositivo para monitorar.</p>
        </div>
      ) : (
        <motion.div layout className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          <AnimatePresence mode="popLayout">
            {devices.map(device => (
              <motion.div
                key={device.id}
                layout
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -12 }}
                className="group bg-foreground hover:bg-muted-foreground border border-border rounded-xl p-5 transition-all duration-300"
              >
                <div className="flex items-center gap-3 mb-3">
                  <div className="p-2 rounded-lg bg-pink-500/10 text-pink-400">
                    <Smartphone size={16} />
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-white">{device.name}</h3>
                    {device.brand && <span className="text-[10px] uppercase tracking-wider text-zinc-500">{device.brand}</span>}
                  </div>
                </div>
                <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
                  <Button 
                    size="sm" variant="ghost"
                    className="h-8 text-xs text-secondary hover:text-primary"
                    onClick={() => navigate(`/devices/${device.id}`)}
                  >
                    Editar
                  </Button>
                  <Button 
                    size="sm" variant="ghost"
                    className="text-red-400 hover:text-red-300 hover:bg-red-400/10 h-8 text-xs"
                    onClick={() => handleDelete(device.id, device.name)}
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

