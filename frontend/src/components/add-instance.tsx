import {useState, useMemo, useEffect} from "react"
import { useForm, useFieldArray, Controller } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
    DialogClose,
    DialogDescription,
    DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { X, Plus } from "lucide-react"
import { Slider } from "./ui/slider"
import Search from "@/components/search.tsx";

// =================== SCHEMA ===================
const formSchema = z.object({
    name: z.string().min(1, "Nome é obrigatório"),
    script: z.string().min(1, "Selecione um script"),
    interval: z.number().min(1, "Intervalo deve ser positivo"),
    unit: z.enum(["seconds", "minutes", "hours"]),
    usina: z.string().min(1, "Selecione uma usina"),
    device: z.string().min(1, "Selecione um dispositivo"),
    deleteAfterSend: z.boolean(),
    storageLimitEnabled: z.boolean(),
    storageLimit: z.number().min(10).max(10240),
    keyValueMode: z.boolean(),
    onError: z.enum(["restart", "stop", "try-again"]),
    args: z.array(
        z.union([
            z.object({ key: z.string(), value: z.string() }),
            z.object({ value: z.string() }),
        ])
    ),
})

// =================== MOCK DATA ===================
const usinas = [
    { id: "u1", name: "Usina A", devices: [{ id: "d1", name: "Dispositivo 1" }, { id: "d2", name: "Dispositivo 2" }] },
    { id: "u2", name: "Usina B", devices: [{ id: "d3", name: "Dispositivo 3" }] },
]

const scripts = [
    { id: "s1", name: "script/intelbras", author: "valber", description: "Captura e lê dados da Intelbras" },
    { id: "s2", name: "script/goodwe", author: "kelwSagashi", description: "Captura e lê dados do inversor Goodwe via TCP/IP" },
]

// =================== COMPONENT ===================
type Props = {
    isOpen: boolean;
    onOpenChange: (isOpen: boolean) => void;
}

export default function InstanceForm({ isOpen, onOpenChange }: Props) {
    const [scriptDialogOpen, setScriptDialogOpen] = useState(false)
    const [searchScript, setSearchScript] = useState("")
    const [selectedScript, setSelectedScript] = useState<string | null>(null)

    const {
        control,
        register,
        handleSubmit,
        setValue,
        watch,
        reset,
        formState: { errors },
    } = useForm({
        resolver: zodResolver(formSchema),
        defaultValues: {
            name: "",
            script: "",
            interval: 10,
            unit: "seconds",
            usina: "",
            device: "",
            deleteAfterSend: false,
            storageLimitEnabled: false,
            storageLimit: 10,
            keyValueMode: false,
            args: [],
        },
    })

    const { fields, append, remove } = useFieldArray({ control, name: "args" })
    const keyValueMode = watch("keyValueMode")
    const selectedUsina = watch("usina")
    watch("device");
// auto-seleção entre usina e dispositivo
    const filteredDevices = useMemo(() => {
        const usina = usinas.find((u) => u.id === selectedUsina)
        return usina ? usina.devices : []
    }, [selectedUsina])

    function handleDeviceSelect(id: string){
        setValue("device", id)
        const usina = usinas.find((u) => u.devices.some((d) => d.id === id))
        if (usina) setValue("usina", usina.id);
    }

    function handleScriptSelect(id: string) {
        setValue("script", id)
        setSelectedScript(id)
        setScriptDialogOpen(false)
    }

    function onSubmit(values: z.infer<typeof formSchema>) {
        console.log("Nova instância:", values)
        onOpenChange(false)
    }

    useEffect(() => {
        if (!isOpen) {
            const timer = setTimeout(() => {
                reset();
                setSelectedScript(null);
            }, 300);
            return () => clearTimeout(timer);
        }
    }, [isOpen, reset]);


    const filteredScripts = scripts.filter((s) =>
        s.name.toLowerCase().includes(searchScript.toLowerCase())
    )

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[425px] h-[90%]">
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 h-full flex flex-col">
                    <DialogHeader>
                        <DialogTitle className="text-primary">Nova Instância</DialogTitle>
                        <DialogDescription>
                            Crie uma nova instância de script aqui.
                            Pressione o botão Adicionar quando concluir.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 flex-grow overflow-hidden">
                        <ScrollArea className="h-full pr-6">
                            <div className="space-y-4">
                                <Input className={"rounded border-border"} placeholder="ID gerado automaticamente" disabled />
                                <div>
                                    <Input className={"rounded text-input placeholder:text-input border-border"} placeholder="Nome" {...register("name")} />
                                    {errors.name && <p className="text-red-500 text-xs mt-1">{errors.name.message}</p>}
                                </div>

                                {/* SCRIPT SELECTOR */}
                                <div>
                                    <Dialog open={scriptDialogOpen} onOpenChange={setScriptDialogOpen}>
                                        <DialogTrigger asChild>
                                            <Button variant="outline" className="rounded w-full justify-between text-input">
                                                {selectedScript
                                                    ? scripts.find((s) => s.id === selectedScript)?.name
                                                    : "Selecione o Script"}
                                            </Button>
                                        </DialogTrigger>
                                        <DialogContent className="max-w-2xl">
                                            <DialogHeader>
                                                <DialogTitle className={"text-primary"}>Selecionar Script</DialogTitle>
                                            </DialogHeader>
                                            <div className="flex items-center gap-2">
                                                <Search place_holder={"Buscar script"} value={searchScript} onChange={(e) => setSearchScript(e.target.value)}/>
                                            </div>
                                            <ScrollArea className="max-h-[400px] mt-4">
                                                <div className="grid grid-cols-1 gap-3">
                                                    {filteredScripts.map((s) => (
                                                        <Card
                                                            key={s.id}
                                                            onClick={() => handleScriptSelect(s.id)}
                                                            className="cursor-pointer hover:bg-muted transition"
                                                        >
                                                            <CardHeader>
                                                                <CardTitle className="text-blue-500 text-sm">{s.name}</CardTitle>
                                                            </CardHeader>
                                                            <CardContent>
                                                                <p className="text-sm text-gray-400">{s.description}</p>
                                                                <p className="text-xs text-gray-500 mt-1">Autor: {s.author}</p>
                                                            </CardContent>
                                                        </Card>
                                                    ))}
                                                </div>
                                            </ScrollArea>
                                        </DialogContent>
                                    </Dialog>
                                    {errors.script && <p className="text-red-500 text-xs mt-1">{errors.script.message}</p>}
                                </div>

                                <div>
                                    <div className="flex gap-2">
                                        <Input
                                            className={"rounded text-input placeholder:text-input border-border"}
                                            type="number"
                                            placeholder="Intervalo de execução"
                                            {...register("interval", { valueAsNumber: true })}
                                        />
                                        <Controller control={control} name="unit" render={({field}) => (
                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                <SelectTrigger className={"rounded border-border w-[180px]"}>
                                                    <SelectValue className={"placeholder:text-input text-input"} placeholder="Unidade" />
                                                </SelectTrigger>
                                                <SelectContent className={"bg-foreground text-input"}>
                                                    <SelectItem className={"text-input"} value="seconds">Segundos</SelectItem>
                                                    <SelectItem value="minutes">Minutos</SelectItem>
                                                    <SelectItem value="hours">Horas</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        )}/>
                                    </div>
                                    {errors.interval && <p className="text-red-500 text-xs mt-1">{errors.interval.message}</p>}
                                </div>

                                {/* USINA E DISPOSITIVO */}
                                <div>
                                    <Controller control={control} name="usina" render={({field}) => (
                                    <Select onValueChange={field.onChange} value={field.value}>
                                        <SelectTrigger className={"w-full"}>
                                            <SelectValue placeholder="Selecione a Usina" />
                                        </SelectTrigger>
                                        <SelectContent className={"bg-foreground text-input"}>
                                            {usinas.map((u) => (
                                                <SelectItem key={u.id} value={u.id}>
                                                    {u.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                        )}/>
                                    {errors.usina && <p className="text-red-500 text-xs mt-1">{errors.usina.message}</p>}
                                </div>

                                <div>
                                    <Controller control={control} name="device" render={({field}) => (
                                        <Select onValueChange={handleDeviceSelect} value={field.value}>
                                            <SelectTrigger className={"w-full"}>
                                                <SelectValue placeholder="Selecione o Dispositivo" />
                                            </SelectTrigger>
                                            <SelectContent className={"bg-foreground text-input"}>
                                                {filteredDevices.map((d) => (
                                                    <SelectItem key={d.id} value={d.id}>
                                                        {d.name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    )}/>
                                    {errors.device && <p className="text-red-500 text-xs mt-1">{errors.device.message}</p>}
                                </div>
                                
                                <div>
                                    <Controller control={control} name="onError" render={({ field }) => (
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <SelectTrigger className={"w-full"}>
                                                <SelectValue placeholder="Em caso de erro" />
                                            </SelectTrigger>
                                            <SelectContent className={"bg-foreground text-input"}>
                                                <SelectItem value="restart">
                                                    Reiniciar instância
                                                </SelectItem>
                                                <SelectItem value="stop">
                                                    Parar instância
                                                </SelectItem>
                                                <SelectItem value="try-again">
                                                    Tentar Novamente
                                                </SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )}
                                    />
                                    {errors.onError && <p className="text-red-500 text-xs mt-1">{errors.onError.message}</p>}
                                </div>

                                {/* SWITCHES */}
                                <div className="flex items-center gap-2 pt-2">
                                    <Controller
                                        control={control}
                                        name="deleteAfterSend"
                                        render={({ field }) => <Switch checked={field.value} onCheckedChange={field.onChange} />}
                                    />
                                    <label>Deletar arquivos após envio</label>
                                </div>

                                <div className="flex items-center gap-2">
                                    <Controller
                                        control={control}
                                        name="storageLimitEnabled"
                                        render={({ field }) => <Switch checked={field.value} onCheckedChange={field.onChange} />}
                                    />
                                    <label>Adicionar limite de espaço ocupado</label>
                                </div>

                                {watch("storageLimitEnabled") && (
                                    <div>
                                        <div className="flex justify-between text-sm">
                                            <span>{watch("storageLimit")}MB</span>
                                            <span>10GB</span>
                                        </div>
                                        <Controller control={control} name="storageLimit" render={({ field }) => (
                                            <Slider min={10} max={10000} step={10} value={[field.value]} onValueChange={v => field.onChange(v[0])} />)} />
                                    </div>
                                )}

                                {/* ARGUMENTOS */}
                                <div className="space-y-2 pt-2">
                                    <div className="flex items-center gap-2">
                                        <Controller
                                            control={control}
                                            name="keyValueMode"
                                            render={({ field }) => <Switch checked={field.value} onCheckedChange={field.onChange} />}
                                        />
                                        <label>Argumento chave/valor</label>
                                    </div>

                                    {fields.map((item, index) => (
                                        <div key={item.id} className="flex gap-2 items-center">
                                            {keyValueMode && <Input placeholder="Chave" {...register(`args.${index}.key`)} />}
                                            <Input placeholder="Valor" {...register(`args.${index}.value`)} />
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                onClick={() => remove(index)}
                                                type="button"
                                            >
                                                <X />
                                            </Button>
                                        </div>
                                    ))}

                                    <Button
                                        variant="outline"
                                        type="button"
                                        onClick={() =>
                                            append(keyValueMode ? { key: "", value: "" } : { value: "" })
                                        }
                                        className="flex items-center gap-2"
                                    >
                                        <Plus size={16} /> Adicionar argumento
                                    </Button>
                                </div>
                            </div>
                        </ScrollArea>
                    </div>

                    <DialogFooter className="pt-4">
                        <DialogClose asChild>
                            <Button variant="outline" type="button">
                                Cancelar
                            </Button>
                        </DialogClose>
                        <Button type="submit">Adicionar</Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    )
}

