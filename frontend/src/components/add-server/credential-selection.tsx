import type { UseFormReturn } from "react-hook-form";
import type { FullServerValues } from "./schemas";
import { useCredentialsStore } from "@/stores/credentials-store";
import { ScrollArea } from "../ui/scroll-area";
import { Label } from "../ui/label";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "../ui/select";

export function CredentialSelecion({ form }: { form: UseFormReturn<any> }) {
    const { credentials } = useCredentialsStore();
    const { setValue, watch } = form;
    const selectedCredId = watch("server.credential_id");
    const serverType = watch("server.type");

    return (
        <ScrollArea className='h-full w-full'>
            <div className="space-y-6 max-w-xl p-4">
                <div>
                    <h3 className="text-secondary text-lg font-medium">Credencial de Autorização</h3>
                    <p className="text-sm text-zinc-400 mt-1">
                    Selecione a credencial que este servidor deve usar para se autenticar. 
                    </p>
                </div>

                <div className="space-y-4">
                    <div className="space-y-3">
                        <Label className='text-primary'>Credencial Utilizada</Label>
                        <Select
                            value={selectedCredId || "none"}
                            onValueChange={(val) => {
                                setValue("server.credential_id", (val === "none" ? undefined : val) as any);
                            }}
                        >
                        <SelectTrigger className="text-primary border-white/[0.1] bg-white/[0.04] h-11">
                            <SelectValue placeholder="Nenhuma (Sem Autorização)" />
                        </SelectTrigger>
                        <SelectContent className='text-primary'>
                            <SelectGroup>
                            <SelectItem value="none">Nenhuma (Sem Autorização)</SelectItem>
                            {credentials.filter(c => c.auth_type_id === serverType).map(c => (
                                <SelectItem key={c.id} value={c.id}>
                                    {c.name}
                                    <span className="opacity-50 text-xs ml-2"></span>
                                </SelectItem>
                            ))}
                            </SelectGroup>
                        </SelectContent>
                        </Select>
                    </div>
                    
                    <div className="pt-2 flex flex-col gap-4">
                        <p className="text-sm text-zinc-500">
                            Não encontrou sua credencial? <a href="/credentials" target="_blank" className="underline hover:text-white transition-colors">Crie uma nova credencial aqui.</a>
                        </p>
                    </div>
                </div>
            </div>
        </ScrollArea>
    )
}

