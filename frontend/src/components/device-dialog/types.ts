import z from "zod";
const DeviceConnectionMethods = ["none", "serial", "tcp", "udp"];
export const deviceFormSchema = z.object({
    id: z.string().optional(),
    name: z.string().min(1, "Nome é obrigatório"),
    brand: z.string().min(1, "Marca é obrigatória"),
    connection: z.enum(DeviceConnectionMethods),
    serial_number: z.string().optional(),
    description: z.string().optional(),
    location: z.string().optional(),
    ip_address: z.string().optional(),
    others: z.array(z.object({
        key: z.string().min(1, "Chave é obrigatória"),
        value: z.string().min(1, "Valor é obrigatório"),
        type: z.enum(["text", "number"])
    })),
    resource_operation_id: z.string().optional(),
});
export type DeviceFormValues = z.infer<typeof deviceFormSchema>;
export type DeviceConnectionFormValues = keyof typeof deviceFormSchema.shape.connection.enum;

export interface DeviceConfigFormProps {
    onSubmit: (deviceData: DeviceFormValues) => void;
}

export type DroppableInputProps = React.ComponentProps<"input"> & {
    accept: string;
    onDrop?: (item: any) => void;
}

export type ServiceType = 'rest' | 'google_drive' | null;

export type DeviceDialogProps = {
    isOpen: boolean;
    onOpenChange: (isOpen: boolean) => void;
}

export const endpointSchema = z.object({
    endpoints: z.array(z.object({
        id: z.string(),
        uniqueKey: z.string(),
        params: z.record(z.string(), z.string().min(1, { error: "Campo obrigatório!" })).optional()
    }))
});

export type EndpointsFormValues = z.infer<typeof endpointSchema>;

