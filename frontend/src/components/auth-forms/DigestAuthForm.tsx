// frontend/src/components/auth-forms/DigestAuthForm.tsx
import React from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import type { FullServerValues } from '../add-server/schemas';


interface DigestAuthFormProps {
    form: UseFormReturn<any>
}

export const DigestAuthForm: React.FC<DigestAuthFormProps> = ({ form }) => {

    return (
        <div className="space-y-4">
            <Controller
                control={form.control}
                name='authorization.authorization.auth_username'
                render={({ field, fieldState }) => (
                    <div className='space-y-2'>
                        <Label htmlFor="digest_auth_username" className="text-primary">Usuário</Label>
                        <Input
                            id="digest_auth_username"
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
                        <Label htmlFor="digest_auth_password" className="text-primary">Senha</Label>
                        <Input
                            id="digest_auth_password"
                            type="password"
                            placeholder="Sua senha"
                            {...field}
                            value={field.value || ''}
                            className={cn(fieldState.error && "border-destructive", "text-primary rounded")}
                        />
                        {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                    </div>
                )} />
            <Controller
                control={form.control}
                name='authorization.authorization.auth_realm'
                render={({ field }) => (
                    <div className='space-y-2'>
                        <Label htmlFor="digest_auth_realm" className="text-primary">Realm (Opcional)</Label>
                        <Input
                            id="digest_auth_realm"
                            placeholder="Realm de autenticação"
                            className={cn("text-primary rounded")}
                            {...field}
                            value={field.value || ''}
                        />
                    </div>
                )} />

        </div>
    );
};
