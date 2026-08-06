import { Separator } from '../ui/separator';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { useFieldArray, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Trash } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ItemTypes } from '@/lib/constants';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip';
import { deviceFormSchema, type DeviceConfigFormProps, type DeviceConnectionFormValues, type DeviceFormValues } from './types';
import DroppableInput from './droppable-input';
import React from 'react';


export default function DeviceConfigForm({
    onSubmit,
    defaultValues
}: DeviceConfigFormProps & { defaultValues?: Partial<DeviceFormValues> }) {
    const form = useForm<DeviceFormValues>({
        resolver: zodResolver(deviceFormSchema),
        defaultValues: {
            id: defaultValues?.id || "",
            name: defaultValues?.name || "",
            brand: defaultValues?.brand || "",
            serial_number: defaultValues?.serial_number || undefined,
            description: defaultValues?.description || "",
            ip_address: defaultValues?.ip_address || "",
            location: defaultValues?.location || "",
            connection: defaultValues?.connection || undefined,
            others: defaultValues?.others || []
        }
    });

    // Reset form when defaultValues change
    React.useEffect(() => {
        if (defaultValues) {
            form.reset({
                ...defaultValues,
                id: defaultValues.id || "",
                name: defaultValues.name || "",
                brand: defaultValues.brand || "",
                others: defaultValues.others || []
            });
        }
    }, [defaultValues, form.reset]);

    const { fields, append, remove } = useFieldArray({
        control: form.control,
        name: "others"
    });

    // Função para lidar com a soltura (drop)
    const handleDrop = (value: string, fieldName: keyof DeviceFormValues | `others.${number}.value`) => {
        form.setValue(fieldName, value, { shouldValidate: true });
    };

    return (
        <form id="device_form" onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 py-2">
            {/* Campos Padrão */}
            <div>
                <label className="text-secondary text-sm" htmlFor="device-id">ID</label>
                <DroppableInput
                    id="device-id"
                    placeholder="ID único do dispositivo"
                    {...form.register("id")}
                    className={form.formState.errors.id ? "border-destructive" : ""}
                    accept={ItemTypes.OUTPUT_VALUE}
                    onDrop={(item) => handleDrop(item.value, 'id')}
                />
                {form.formState.errors.id && <p className="text-destructive text-sm mt-1">{form.formState.errors.id.message}</p>}
            </div>
            <div>
                <label className="text-secondary text-sm" htmlFor="device-name">Nome</label>
                <DroppableInput
                    id="device-name"
                    placeholder="Nome do dispositivo (ex: Inversor 1)"
                    {...form.register("name")}
                    className={form.formState.errors.name ? "border-destructive" : ""}
                    accept={ItemTypes.OUTPUT_VALUE}
                    onDrop={(item) => handleDrop(item.value, 'name')}
                />
                {form.formState.errors.name && <p className="text-destructive text-sm mt-1">{form.formState.errors.name.message}</p>}
            </div>
            <div>
                <label className="text-secondary text-sm" htmlFor="device-brand">Marca</label>
                <DroppableInput
                    id="device-brand"
                    placeholder="Marca do dispositivo (ex: Intelbras)"
                    {...form.register("brand")}
                    className={form.formState.errors.brand ? "border-destructive" : ""}
                    accept={ItemTypes.OUTPUT_VALUE}
                    onDrop={(item) => handleDrop(item.value, 'brand')}
                />
                {form.formState.errors.brand && <p className="text-destructive text-sm mt-1">{form.formState.errors.brand.message}</p>}
            </div>
            <div>
                <label className="text-secondary text-sm" htmlFor="device-description">Descrição</label>
                <DroppableInput
                    id="device-description"
                    placeholder="Descrição do dispositivo"
                    {...form.register("description")}
                    className={form.formState.errors.description ? "border-destructive" : ""}
                    accept={ItemTypes.OUTPUT_VALUE}
                    onDrop={(item) => handleDrop(item.value, 'description')}
                />
                {form.formState.errors.description && <p className="text-destructive text-sm mt-1">{form.formState.errors.description.message}</p>}
            </div>
            <div>
                <label className="text-secondary text-sm" htmlFor="device-ip_address">Endereço IP</label>
                <DroppableInput
                    id="device-ip_address"
                    placeholder="Endereço IP do dispositivo"
                    {...form.register("ip_address")}
                    className={form.formState.errors.ip_address ? "border-destructive" : ""}
                    accept={ItemTypes.OUTPUT_VALUE}
                    onDrop={(item) => handleDrop(item.value, 'ip_address')}
                />
                {form.formState.errors.ip_address && <p className="text-destructive text-sm mt-1">{form.formState.errors.ip_address.message}</p>}
            </div>
            <div>
                <label className="text-secondary text-sm" htmlFor="device-location">Locação</label>
                <DroppableInput
                    id="device-location"
                    placeholder="Onde está o dispositivo?"
                    {...form.register("location")}
                    className={form.formState.errors.location ? "border-destructive" : ""}
                    accept={ItemTypes.OUTPUT_VALUE}
                    onDrop={(item) => handleDrop(item.value, 'location')}
                />
                {form.formState.errors.location && <p className="text-destructive text-sm mt-1">{form.formState.errors.location.message}</p>}
            </div>
            <div>
                <label className="text-secondary text-sm" htmlFor={`connection`}>Tipo</label>
                <Select onValueChange={(val: DeviceConnectionFormValues) => form.setValue(`connection`, val)} defaultValue={form.watch(`connection`)}>
                    <SelectTrigger className="w-30 text-primary">
                        <SelectValue placeholder="Selecionar tipo" />
                    </SelectTrigger>
                    <SelectContent className='text-primary'>
                        {Object.values(deviceFormSchema.shape.connection.enum).map((val, _) => (
                            <SelectItem value={val}>{val}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                {form.formState.errors.connection && <p className="text-destructive text-sm mt-1">{form.formState.errors.connection?.message}</p>}
            </div>

            <Separator className="my-6" />

            {/* Campos Adicionais "Others" */}
            <h4 className="text-foreground text-lg font-medium">Outros Campos (Personalizados)</h4>
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
                                    <Tooltip open={(!!form.formState.errors.others?.[index]?.key)}>
                                        <TooltipTrigger asChild>
                                            <Input
                                                id={`others-${index}-key`}
                                                placeholder="Nome do campo (ex: serial_number)"
                                                {...form.register(`others.${index}.key`)}
                                                className={cn(
                                                    form.formState.errors.others?.[index]?.key ? "border-destructive" : "",
                                                    "text-primary border-none focus-visible:ring-0 px-0"
                                                )}
                                            />
                                        </TooltipTrigger>
                                        <TooltipContent>
                                            {form.formState.errors.others?.[index]?.key && <p className="text-destructive text-sm mt-1">{form.formState.errors.others[index]?.key?.message}</p>}
                                        </TooltipContent>
                                    </Tooltip>
                                </div>
                                <DroppableInput
                                    id={`others-${index}-value`}
                                    placeholder="Valor do campo"
                                    {...form.register(`others.${index}.value`)}
                                    className={cn(
                                        form.formState.errors.others?.[index]?.value ? "border-destructive" : "",
                                        ""
                                    )}
                                    accept={ItemTypes.OUTPUT_VALUE}
                                    onDrop={(item) => handleDrop(item.value, `others.${index}.value`)}
                                />
                                {form.formState.errors.others?.[index]?.value && <p className="text-destructive text-sm mt-1">{form.formState.errors.others[index]?.value?.message}</p>}
                            </div>
                            <div>
                                <div className='h-9 py-1'>
                                    <label className="text-secondary text-sm" htmlFor={`others-${index}-type`}>Tipo</label>
                                </div>
                                <Select onValueChange={(val: "text" | "number") => form.setValue(`others.${index}.type`, val)} defaultValue={form.watch(`others.${index}.type`)}>
                                    <SelectTrigger className="w-30 text-primary">
                                        <SelectValue placeholder="Selecionar tipo" />
                                    </SelectTrigger>
                                    <SelectContent className='text-primary'>
                                        <SelectItem value="text">Texto</SelectItem>
                                        <SelectItem value="number">Numérico</SelectItem>
                                    </SelectContent>
                                </Select>
                                {form.formState.errors.others?.[index]?.type && <p className="text-destructive text-sm mt-1">{form.formState.errors.others[index]?.type?.message}</p>}
                            </div>
                        </div>
                    </div>
                </div>
            ))}
            <Button type="button" variant="outline" onClick={() => append({ key: "", value: "", type: "text" })}>
                Adicionar Campo Personalizado
            </Button>
        </form>
    );
}
