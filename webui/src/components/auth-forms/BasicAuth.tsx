// frontend/src/components/auth-forms/BasicAuthForm.tsx
import React from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import type { FullServerValues } from '../add-server/schemas';

interface BasicAuthFormProps {
    form: UseFormReturn<any>
}

export const BasicAuthForm: React.FC<BasicAuthFormProps> = ({ form }) => {
    return (
        <div className="space-y-4">
            <Controller
                control={form.control}
                name='authorization.authorization.auth_username'
                render={({ field, fieldState }) => (

                    <div className='space-y-2'>
                        <Label htmlFor="auth_username" className='text-primary'>Usuário</Label>
                        <Input
                            placeholder="Nome de usuário"
                            {...field}
                            value={field.value || ''}
                            className={cn(fieldState.error && "border-destructive", "text-primary rounded")}
                        />
                        {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                    </div>
                )} />
            <Controller
                control={form.control}
                name='authorization.authorization.auth_password'
                render={({ field, fieldState }) => (
                    <div className='space-y-2'>
                        <Label htmlFor="auth_password" className='text-primary'>Senha</Label>
                        <Input
                            type="password"
                            placeholder="Sua senha"
                            {...field}
                            value={field.value || ''}
                            className={cn(fieldState.error && "border-destructive", "text-primary rounded")}
                        />
                        {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                    </div>
                )} />
        </div>
    );
};
