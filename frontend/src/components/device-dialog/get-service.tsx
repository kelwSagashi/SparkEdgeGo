import { api } from "@/server/server.service";
import type { ServerWithRelations } from "@/interfaces/servers";
import { useEffect, useState } from "react";
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "../ui/select";

export default function GetService({
    setServerSelected
}: {
    setServerSelected: React.Dispatch<React.SetStateAction<ServerWithRelations | undefined>>
}) {
    const [selection, setSelection] = useState<ServerWithRelations>();
    const [servers, setServers] = useState<ServerWithRelations[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        let mounted = true;
        setLoading(true);
        api.listServers()
            .then((res) => {
                if (!mounted) return;
                const data = res.data?.data ?? res.data ?? [];
                setServers(data as ServerWithRelations[]);
            })
            .catch((err) => setError(err?.message ?? String(err)))
            .finally(() => setLoading(false));

        return () => { mounted = false };
    }, []);
    useEffect(() => {
        setServerSelected(selection);
    }, [setServerSelected, selection]);

    return (
        <div className='w-full'>
            <span className="font-medium text-secondary">Escolha o serviço</span>
            <Select onValueChange={(val) => {
                const selectedServer = servers.find(s => s.id === val);
                if (selectedServer) setSelection(selectedServer);
            }}>
                <SelectTrigger className="w-full p-6">
                    <SelectValue placeholder="Selecionar um servidor" />
                </SelectTrigger>
                <SelectContent>
                    <SelectGroup>
                        <SelectLabel>Servidores</SelectLabel>
                        {loading && <div className='p-4 text-secondary'>Loading...</div>}
                        {error && <div className='p-4 text-destructive text-sm'>{error}</div>}
                        {!loading && !error && servers.map(item => (
                            <SelectItem key={item.id} id={item.id} value={item.id} className='w-full justify-start p-6'>
                                <div className="text-left">
                                    <p className="font-semibold text-primary">{item.name}</p>
                                    <p className="text-sm text-secondary">{item.base_url}</p>
                                </div>
                            </SelectItem>
                        ))}
                    </SelectGroup>
                </SelectContent>
            </Select>
        </div>
    )
}
