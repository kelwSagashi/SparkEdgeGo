import { Separator } from '../ui/separator';
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '../ui/select';
import { Button } from '../ui/button';
import React, { useCallback, useEffect, useState, useMemo } from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { JsonViewMain } from '@/components/json-view/json-view';
import { api } from '@/server/server.service';
import type { 
    DeviceUpsertValues, 
    ServerReturningValues, 
} from "@/types/db";
import { type DeviceFormValues } from './types';
import DeviceConfigForm from './device-config-form';
import { Play, Loader2, Info } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { useNavigate } from 'react-router-dom';


export default function DeviceForm({ 
    setSaving,
    initialData
}: { 
    setSaving: React.Dispatch<React.SetStateAction<boolean>>;
    initialData?: any;
    }) {
    const navigate = useNavigate();
    const isEdit = !!initialData?.id;
    // Helper State
    const [servers, setServers] = useState<ServerReturningValues[]>([]);
    const [selectedServerId, setSelectedServerId] = useState<string>("");
    const [resources, setResources] = useState<any[]>([]); 
    const [selectedResourceId, setSelectedResourceId] = useState<string>("");
    const [selectedOperationId, setSelectedOperationId] = useState<string>("");
    const [executionInput, setExecutionInput] = useState<any>({});
    const [executionResult, setExecutionResult] = useState<any>(null);
    const [executing, setExecuting] = useState(false);
    const [loadingResources, setLoadingResources] = useState(false);

    // Initial Fetch
    useEffect(() => {
        const fetchServers = async () => {
            try {
                const res = await api.listAllServers();
                if (res.data?.data) {
                    setServers(res.data.data);
                }
            } catch (err) {
                console.error("Failed to fetch servers", err);
            }
        };
        fetchServers();
    }, []);

    // Handle Server Change
    useEffect(() => {
        if (!selectedServerId) {
            setResources([]);
            setSelectedResourceId("");
            return;
        }

        const fetchResources = async () => {
            setLoadingResources(true);
            try {
                // api.listResources uses the server ID to get its resources and operations
                const res = await api.listResources(selectedServerId);
                if (res.data?.data) {
                    setResources(res.data.data);
                } else {
                    setResources([]);
                }
            } catch (err) {
                console.error("Failed to fetch resources", err);
                setResources([]);
            } finally {
                setLoadingResources(false);
            }
        };
        fetchResources();
    }, [selectedServerId]);

    // Derived Selection
    const selectedResource = useMemo(() => 
        resources.find(r => r.resource.id === selectedResourceId),
    [resources, selectedResourceId]);

    const selectedOperation = useMemo(() => 
        selectedResource?.operations.find((o: any) => o.id === selectedOperationId),
    [selectedResource, selectedOperationId]);

    // When operation changes, set execution input from its schema if available
    useEffect(() => {
        if (selectedOperation?.input_schema) {
            setExecutionInput(selectedOperation.input_schema);
        } else {
            setExecutionInput({});
        }
        setExecutionResult(null);
    }, [selectedOperation]);

    const updateNestedValue = useCallback((obj: any, path: string, key: string, value: any) => {
        const current = JSON.parse(JSON.stringify(obj || {}));
        if (!path) {
            current[key] = value;
            return current;
        }

        const parts = path.split(".");
        let ref = current;
        for (const part of parts) {
            if (!ref[part] || typeof ref[part] !== "object") {
                ref[part] = {};
            }
            ref = ref[part];
        }
        ref[key] = value;
        return current;
    }, []);

    const handleExecute = useCallback(async () => {
        if (!selectedOperationId) return;
        setExecuting(true);
        try {
            const res = await api.executeOperation(selectedOperationId, executionInput);
            console.log(res);
            if (res.data?.success) {
                setExecutionResult(res.data.data);
            } else {
                alert(res.data?.error || "Falha na execução");
            }
        } catch (err: any) {
            alert("Erro: " + (err.message || String(err)));
        } finally {
            setExecuting(false);
        }
    }, [selectedOperationId, executionInput]);

    const handleSaveDevice = useCallback(async (deviceData: DeviceFormValues) => {
        setSaving(true);
        try {
            const payload: DeviceUpsertValues = {
                id: initialData?.id,
                device_id: deviceData.id,
                name: deviceData.name,
                brand: deviceData.brand,
                serial_number: deviceData.serial_number,
                connection_method: deviceData.connection as any,
                others: deviceData.others,
                description: deviceData.description,
                ip_address: deviceData.ip_address,
                location: deviceData.location,
                resource_operation_id: selectedOperationId || deviceData.resource_operation_id
            };

            const res = isEdit 
                ? await api.updateDevice(initialData.id, payload)
                : await api.createDevice(payload);
                
            const data = res.data?.data
            if (data) {
                navigate('/devices');
                alert(isEdit ? 'Dispositivo atualizado com sucesso!' : 'Dispositivo criado com sucesso!');
            } else {
                alert(res.data.error ?? 'Falha ao salvar dispositivo');
            }
        } catch (err: any) {
            alert('Erro: ' + (err?.message ?? String(err)));
        } finally {
            setSaving(false);
        }
    }, [selectedOperationId, setSaving, initialData, isEdit]);

    return (
        <div className="w-full flex h-[80vh] overflow-hidden">
            {/* Main Form Section */}
            <div className="flex-1 flex flex-col border-r border-border pr-6 mr-6 h-full overflow-hidden">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-xl font-semibold">Configuração do Dispositivo</h2>
                </div>
                <ScrollArea className='h-full overflow-hidden'>
                    <div className="flex-1 pr-4">
                        <DeviceConfigForm 
                            onSubmit={handleSaveDevice} 
                            defaultValues={initialData ? {
                                id: initialData.device_id,
                                name: initialData.name,
                                brand: initialData.brand,
                                description: initialData.description || "",
                                ip_address: initialData.ip_address || "",
                                location: initialData.location || "",
                                connection: initialData.connection_method,
                                others: initialData.others || []
                            } : undefined} 
                        />
                    </div>
                </ScrollArea>
                <div className="pt-4 border-t border-border mt-auto flex justify-end gap-2">
                    <Button variant="outline" onClick={() => navigate('/devices')}>Cancelar</Button>
                    <Button type="submit" form="device_form">{isEdit ? 'Atualizar Dispositivo' : 'Criar Dispositivo'}</Button>
                </div>
            </div>

            {/* Execution Helper Section */}
            <div className="w-[450px] flex flex-col h-full overflow-hidden">
                <div className="flex items-center gap-2 mb-4">
                    <div className="p-2 bg-primary/10 rounded-full">
                        <Play className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                        <h2 className="text-lg font-semibold leading-none">Auxiliar de Dados</h2>
                        <p className="text-xs text-secondary mt-1">Busque dados de um servidor para preencher o formulário</p>
                    </div>
                </div>

                <div className="space-y-4 mb-4">
                    <div>
                        <label className="text-xs text-secondary mb-1 block">Servidor/Serviço</label>
                        <Select value={selectedServerId} onValueChange={setSelectedServerId}>
                            <SelectTrigger>
                                <SelectValue placeholder="Selecione um servidor" />
                            </SelectTrigger>
                            <SelectContent>
                                {servers.map(s => (
                                    <SelectItem className='text-primary' key={s.id} value={s.id!}>
                                        <div className='text-primary'>{s.name}</div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    {selectedServerId && (
                        <div className="grid grid-cols-2 gap-2">
                            <div>
                                <label className="text-xs text-secondary mb-1 block">Recurso</label>
                                <Select value={selectedResourceId} onValueChange={setSelectedResourceId} disabled={loadingResources}>
                                    <SelectTrigger>
                                        <SelectValue placeholder={loadingResources ? "Carregando..." : "Selecione"} />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {resources.map(r => (
                                            <SelectItem key={r.resource.id} value={r.resource.id!}>
                                                <div className='text-primary'>{r.resource.name}</div>
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <div>
                                <label className="text-xs text-secondary mb-1 block">Operação</label>
                                <Select value={selectedOperationId} onValueChange={setSelectedOperationId} disabled={!selectedResourceId}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Selecione" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {selectedResource?.operations.map((o: any) => (
                                            <SelectItem key={o.id} value={o.id!}>
                                                <div className='text-primary'>{o.name || o.type}</div>
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>
                    )}
                </div>

                <div className="flex-1 flex flex-col min-h-0 bg-accent/30 rounded-lg p-4 border border-border">
                    {!selectedOperationId ? (
                        <div className="flex-1 flex flex-col items-center justify-center text-center p-6 empty-state">
                            <Info className="w-10 h-10 text-secondary/30 mb-2" />
                            <p className="text-sm text-secondary">Selecione uma operação para começar a buscar dados</p>
                        </div>
                    ) : (
                        <div className="flex flex-col min-h-0">
                            <div>
                                <div className="flex items-center justify-between mb-2">
                                    <h3 className="text-xs font-bold uppercase tracking-wider text-secondary">Execução</h3>
                                    <Button 
                                        size="sm" 
                                        onClick={handleExecute} 
                                        disabled={executing}
                                        className="h-7 px-3 text-xs"
                                    >
                                        {executing ? <Loader2 className="w-3 h-3 animate-spin mr-1" /> : <Play className="w-3 h-3 mr-1" />}
                                        Executar
                                    </Button>
                                </div>
                            </div>

                        <div className="overflow-hidden">
                            <ScrollArea className="h-full mb-6">
                                <div className="space-y-4">
                                    {selectedOperation?.input_schema && (
                                        <div>
                                            <span className="text-[10px] text-secondary font-mono mb-1 block">PARÂMETROS DE ENTRADA</span>
                                            <JsonViewMain 
                                                data={executionInput}
                                                templateContext={executionResult && typeof executionResult === "object" ? executionResult : undefined}
                                                onParamChange={(path, param, val) => {
                                                    setExecutionInput((prev: any) => updateNestedValue(prev, path, param, val));
                                                }}
                                                inputProps={{ className: "h-7 text-xs" }}
                                            />
                                        </div>
                                    )}

                                    {executionResult && (
                                        <div className="">
                                            <span className="text-[10px] text-emerald-500 font-mono mb-1 block uppercase font-bold">Resultado (Arraste os valores)</span>
                                            <JsonViewMain 
                                                data={executionResult}
                                                draggableValue={true}
                                                pProps={{
                                                    className: "text-xs py-1 px-2 rounded hover:bg-white/10 cursor-grab active:cursor-grabbing text-emerald-300 font-mono border border-transparent hover:border-emerald-500/30"
                                                }}
                                            />
                                        </div>
                                    )}
                                </div>
                            </ScrollArea>
                        </div>
                        </div>
                    )}
                </div>

                {/* <Alert className="mt-4 bg-primary/5 border-primary/20">
                    <Info className="h-4 w-4" />
                    <AlertTitle className="text-xs">Dica</AlertTitle>
                    <AlertDescription className="text-[10px]">
                        Você pode arrastar os campos do resultado diretamente para os inputs do formulário à esquerda.
                    </AlertDescription>
                </Alert> */}
            </div>
        </div>
    );
}
