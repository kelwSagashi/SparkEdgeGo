import React from 'react';
import { useFormContext } from 'react-hook-form';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';

export interface FieldMetadata {
  key: string;
  label: string;
  type: 'text' | 'number' | 'select' | 'password' | 'textarea' | 'boolean' | 'json';
  placeholder?: string;
  description?: string;
  required?: boolean;
  options?: { label: string; value: string }[];
  defaultValue?: any;
}

interface DynamicFieldProps {
  field: FieldMetadata;
  path: string; // e.g. "server.config" or "resources.0.config"
}

export const DynamicField: React.FC<DynamicFieldProps> = ({ field, path }) => {
  const { control } = useFormContext();
  const fieldPath = `${path}.${field.key}`;

  return (
    <FormField
      control={control}
      name={fieldPath}
      render={({ field: formField }) => (
        <FormItem className="flex flex-col space-y-1.5">
          <div className="flex items-center justify-between">
            <FormLabel>
              <div className="text-primary">
                {field.label}
              </div>
            </FormLabel>
            {field.type === 'boolean' && (
              <FormControl>
                <Switch
                  checked={formField.value}
                  onCheckedChange={formField.onChange}
                />
              </FormControl>
            )}
          </div>
          
          {field.type !== 'boolean' && (
            <FormControl>
              {(() => {
                switch (field.type) {
                  case 'select':
                    return (
                      <Select
                        onValueChange={formField.onChange}
                        defaultValue={formField.value}
                      >
                        <SelectTrigger className='text-primary'>
                          <SelectValue placeholder={field.placeholder || "Selecione uma opção"} />
                        </SelectTrigger>
                        <SelectContent>
                          {field.options?.map((opt, idx) => (
                            <SelectItem key={`${fieldPath}-${opt.value}-${idx}`} value={opt.value}>
                              <div className="text-primary">
                                {opt.label}
                              </div>
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    );
                  case 'textarea':
                  case 'json':
                    return (
                      <Textarea
                        {...formField}
                        placeholder={field.placeholder}
                        className="bg-input min-h-[100px] text-secondary"
                      />
                    );
                  case 'password':
                    return (
                      <Input
                        {...formField}
                        type="password"
                        placeholder={field.placeholder}
                        className="bg-input text-secondary"
                      />
                    );
                  case 'number':
                    return (
                      <Input
                        {...formField}
                        type="number"
                        placeholder={field.placeholder}
                        onChange={(e) => formField.onChange(e.target.valueAsNumber)}
                        className="bg-input text-secondary"
                      />
                    );
                  default:
                    return (
                      <Input
                        {...formField}
                        placeholder={field.placeholder}
                        className="bg-input text-secondary"
                      />
                    );
                }
              })()}
            </FormControl>
          )}

          {field.description && <FormDescription><div className="text-secondary"> {field.description}</div></FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
};

