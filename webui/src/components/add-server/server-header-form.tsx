import { useFieldArray, type UseFormReturn } from "react-hook-form";
import type { FullServerValues } from "./schemas";
import { Button } from "../ui/button";
import { Trash } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "../ui/select";
import { Input } from "../ui/input";
import { cn } from "@/lib/utils";

export 
function ServerHeaderForm({ form }: {form: UseFormReturn<FullServerValues> }) {
  const {
    append,
    remove,
    fields
  } = useFieldArray({
    control: form.control,
    name: 'server.headers',
  });

  return (
    <div className="gap-6 pt-4">
      <div className="col-span-3 space-y-4">
        <div className="space-y-4">
          <h3 className="text-secondary text-lg font-medium">Configurar Cabeçalhos</h3>
            {fields.map((field, index) => (
              <div key={field.id} className="flex flex-col gap-2 p-3 border border-border rounded-md relative">
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 text-secondary hover:text-destructive"
                    onClick={() => remove(index)}
                >
                    <Trash className="h-4 w-4" />
                </Button>
                <div>
                    <div className='flex flex-row gap-2 w-full justify-between'>
                        <div className='w-full'>
                            <div className='h-9'>
                                <Tooltip open={(!!form.formState.errors.server?.headers?.[index]?.key)}>
                                    <TooltipTrigger asChild>
                                        <Input
                                            id={`server-header-${index}-key`}
                                            placeholder="Nome do campo"
                                            {...form.register(`server.headers.${index}.key`)}
                                            className={cn(
                                                form.formState.errors.server?.headers?.[index]?.key ? "border-destructive" : "",
                                                "text-primary border-none focus-visible:ring-0 px-0"
                                            )}
                                        />
                                    </TooltipTrigger>
                                    <TooltipContent>
                                        {form.formState.errors.server?.headers?.[index]?.key && <p className="text-destructive text-sm mt-1">{form.formState.errors.server?.headers[index]?.key?.message}</p>}
                                    </TooltipContent>
                                </Tooltip>
                            </div>
                            <Input
                                id={`server-header-${index}-value`}
                                placeholder="Valor do campo"
                                {...form.register(`server.headers.${index}.value`)}
                                className={cn(
                                    form.formState.errors.server?.headers?.[index]?.value ? "border-destructive" : "",
                                    "text-primary"
                                )}
                            />
                            {form.formState.errors.server?.headers?.[index]?.value && <p className="text-destructive text-sm mt-1">{form.formState.errors.server?.headers[index]?.value?.message}</p>}
                        </div>
                        <div>
                            <div className='h-9 py-1'>
                                <label className="text-secondary text-sm" htmlFor={`server-header-${index}-type`}>Tipo</label>
                            </div>
                            <Select onValueChange={(val: "text" | "number") => form.setValue(`server.headers.${index}.type`, val)} defaultValue={form.watch(`server.headers.${index}.type`)}>
                                <SelectTrigger className="w-30 text-primary">
                                    <SelectValue placeholder="Selecionar tipo" />
                                </SelectTrigger>
                                <SelectContent className='text-primary'>
                                    <SelectItem value="text">Texto</SelectItem>
                                    <SelectItem value="number">Numérico</SelectItem>
                                </SelectContent>
                            </Select>
                            {form.formState.errors.server?.headers?.[index]?.type && <p className="text-destructive text-sm mt-1">{form.formState.errors.server?.headers?.[index]?.type?.message}</p>}
                        </div>
                    </div>
                </div>
              </div>
            ))}
            <Button type="button" variant="outline" onClick={() => append({ key: "", value: "", type: "text" })}>
              Adicionar
            </Button>
        </div>
      </div>
    </div>
  );
}
