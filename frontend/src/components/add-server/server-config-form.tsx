import type { UseFormReturn } from "react-hook-form";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import type { ServerBuilderValues } from "./schemas";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { CredentialSelecion } from './credential-selection';
import type { AdapterMetadata } from "@/server/server.service";

interface ServerConfigFormProps {
  form: UseFormReturn<ServerBuilderValues>;
  metadata: AdapterMetadata[];
}

export function ServerConfigForm({ form, metadata }: ServerConfigFormProps) {
  const { register, watch, setValue, formState: { errors } } = form;
  const selectedType = watch("server.type");
  const validMetadata = metadata.filter(
    (item): item is AdapterMetadata => !!item && typeof item.id === "string",
  );

  return (
    <div className="gap-6 pt-4">
      <div className="col-span-3 space-y-6">
        <div className="space-y-4">
          <h3 className="text-secondary text-lg font-medium">Configurações Gerais</h3>
          
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2 col-span-2">
              <Label className='text-primary'>Nome de Exibição do Servidor</Label>
              <Input className='text-primary rounded h-11' {...register('server.name')} placeholder="Ex: Banco de Produção" />
              {errors.server?.name && <p className="text-sm text-red-500 mt-1">{errors.server?.name.message}</p>}
            </div>

            <div className="space-y-2">
              <Label className='text-primary'>Tipo de atenticação</Label>
              <Select
                value={selectedType || "none"}
                onValueChange={(val) => {
                  setValue("server.type", (val === "none" ? "" : val));
                  setValue("server.driver_key", (val === "none" ? "" : val));
                  const metadataSelected = validMetadata.find((m) => m.id === val);
                  if (metadataSelected) {
                    setValue("server.server_type_id", metadataSelected.server_type_id ?? "");
                  }
                }}
              >
                <SelectTrigger className="text-primary border-white/[0.1] bg-white/[0.04] h-11">
                  <SelectValue placeholder="Selecione o tipo..." />
                </SelectTrigger>
                <SelectContent className='text-primary'>
                  <SelectGroup>
                    <SelectItem value="none">Nenhum</SelectItem>
                    {validMetadata.map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
              {errors.server?.type && <p className="text-sm text-red-500 mt-1">{errors.server?.type.message}</p>}
            </div>

            <div className="pt-2 col-span-2">
               <CredentialSelecion form={form} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

