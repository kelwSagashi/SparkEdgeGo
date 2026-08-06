/* eslint-disable @typescript-eslint/no-unused-vars */
import React, { useState } from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput, InputGroupText } from '../ui/input-group';
import { Eye, EyeClosed } from 'lucide-react';
import type { FullServerValues } from '../add-server/schemas';



interface BearerTokenAuthFormProps {
    form: UseFormReturn<any>
}

export const BearerTokenAuthForm: React.FC<BearerTokenAuthFormProps> = ({ form }) => {
    const [tokenInputType, setTokenInputType] = useState<React.HTMLInputTypeAttribute>("password");
    return (
        <div className="space-y-4">
            <Controller
                control={form.control}
                name='authorization.authorization.auth_token'
                render={({ field, fieldState }) => (
                    <div className='flex flex-row gap-4'>
                        <div className='gap-4 items-center flex flex-row w-full'>
                            <Label htmlFor="auth_token" className='text-primary'>Token</Label>
                            <InputGroup>
                                <InputGroupAddon>
                                    <InputGroupText>Bearer</InputGroupText>
                                </InputGroupAddon>
                                <InputGroupInput
                                    type={tokenInputType}
                                    placeholder="Token"
                                    {...field}
                                    value={field.value || ''}
                                    className={cn(fieldState.error && "border-destructive", "text-primary")}
                                />
                                <InputGroupAddon align="inline-end">
                                    <InputGroupButton
                                        variant="ghost"
                                        aria-label="Info"
                                        size="icon-xs"
                                        onClick={() => setTokenInputType(val => {
                                            if (val === "password") return "text";
                                            return "password";
                                        })}
                                    >
                                        {tokenInputType === "password" ? <EyeClosed /> : <Eye />}
                                    </InputGroupButton>
                                </InputGroupAddon>
                            </InputGroup>
                        </div>
                        {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                    </div>
                )} />
        </div>
    );
};
