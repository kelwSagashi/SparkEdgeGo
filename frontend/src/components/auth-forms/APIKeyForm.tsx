/* eslint-disable @typescript-eslint/no-unused-vars */
// frontend/src/components/auth-forms/ApiKeyAuthForm.tsx
import React, { useState } from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label'; // Assumindo que Label vem do shadcn/ui/label ou similar
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { cn } from '@/lib/utils'; // Para classes de utilidade
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput, InputGroupText } from '../ui/input-group';
import { Eye, EyeClosed } from 'lucide-react';
import type { FullServerValues } from '../add-server/schemas';

interface ApiKeyAuthFormProps {
    form: UseFormReturn<any>
}

export const ApiKeyAuthForm: React.FC<ApiKeyAuthFormProps> = ({ form }) => {
    const [apiKeyInputType, setApiKeyInputType] = useState<React.HTMLInputTypeAttribute>("password");

    return (
        <div className="w-full space-y-4">
            <Controller
                control={form.control}
                name="authorization.authorization.auth_token"
                render={({ field, fieldState }) => (
                    <div className='space-y-2'>
                        <Label htmlFor="auth_token" className='text-primary'>API Key</Label>
                        <InputGroup>
                            <InputGroupAddon>
                                <InputGroupText>API Key</InputGroupText>
                            </InputGroupAddon>
                            <InputGroupInput
                                type={apiKeyInputType}
                                placeholder="Sua API Key"
                                {...field}
                                value={field.value || ''}
                                className={cn(fieldState.error && "border-destructive", "w-full text-primary")}
                            />

                            <InputGroupAddon align="inline-end">
                                <InputGroupButton
                                    variant="ghost"
                                    aria-label="Info"
                                    size="icon-xs"
                                    onClick={() => setApiKeyInputType(val => {
                                        if (val === "password") return "text";
                                        return "password";
                                    })}
                                >
                                    {apiKeyInputType === "password" ? <EyeClosed /> : <Eye />}
                                </InputGroupButton>
                            </InputGroupAddon>
                        </InputGroup>
                        {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                    </div>
                )} />

            <Controller
                control={form.control}
                name='authorization.authorization.auth_api_key_location'
                render={({ field, fieldState }) => (

                    <div className='space-y-2'>
                        <Label htmlFor="auth_api_key_location" className='text-primary'>Adicionar ao</Label>
                        <Select
                            onValueChange={field.onChange}
                            value={field.value}
                        >
                            <SelectTrigger className={cn(fieldState.error && "border-destructive", "w-full text-primary")}>
                                <SelectValue placeholder="Onde enviar a API Key?" />
                            </SelectTrigger>
                            <SelectContent className='text-primary w-full'>
                                <SelectGroup>
                                    <SelectItem value="header">Cabeçalho HTTP</SelectItem>
                                    <SelectItem value="query">Parâmetro de Query (URL)</SelectItem>
                                </SelectGroup>
                            </SelectContent>
                        </Select>
                        {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                    </div>
                )} />

            <Controller
                control={form.control}
                name='authorization.authorization.auth_header_name'
                render={({ field, fieldState }) => (
                    <>
                        {form.getValues('authorization.authorization.auth_api_key_location') === "header" && (
                            <div className='space-y-2'>
                                <Label htmlFor="auth_header_name" className='text-primary'>Nome do Cabeçalho</Label>
                                <Input
                                    placeholder="Ex: X-API-Key, Authorization"
                                    {...field}
                                    value={field.value || ''}
                                    className={cn(fieldState.error && "border-destructive", "w-full text-primary")}
                                />
                                {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                            </div>
                        )}
                    </>
                )
                } />

            <Controller
                control={form.control}
                name='authorization.authorization.auth_query_param_name'
                render={({ field, fieldState }) => (
                    <>
                        {form.getValues('authorization.authorization.auth_api_key_location') === "query" && (
                            <div className='space-y-2'>
                                <Label htmlFor="auth_query_param_name" className='text-primary'>Nome do Parâmetro de Query</Label>
                                <Input
                                    placeholder="Ex: api_key, key"
                                    {...field}
                                    value={field.value || ''}
                                    className={cn(fieldState.error && "border-destructive", "w-full text-primary")}
                                />
                                {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                            </div>
                        )}

                    </>
                )} />
        </div>
    );
};
