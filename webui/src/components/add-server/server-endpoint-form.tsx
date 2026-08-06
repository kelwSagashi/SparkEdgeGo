import { Controller, useFieldArray, type UseFormReturn } from "react-hook-form";
import { Methods, type FullServerValues } from "./schemas";
import { useCallback, useMemo, useState } from "react";
import axios from "axios";
import { headersArrayToRecord } from "./helpers";
import { Button } from "../ui/button";
import { CodeXml, Play, Trash } from "lucide-react";
import { Label } from "../ui/label";
import { Input } from "../ui/input";
import { cn } from "@/lib/utils";
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from "../ui/input-group";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "../ui/select";
import { JsonViewMain } from "../json-view/json-view";

function EndpointSelect({
  onChange,
  defaultValue,
}: {
  onChange: (value: string) => void;
  defaultValue: string;
}) {
  const [selectedMethod, setSelectedMethod] = useState('GET');
  const methodStyle = useMemo(() => Methods.find((m) => m.value === selectedMethod)?.className, [selectedMethod]);
  return (

    <Select value={selectedMethod} onValueChange={(value) => {
      setSelectedMethod(value);
      onChange(value);
    }}>
      <SelectTrigger className={cn("text-primary border-none", methodStyle)}>
        <SelectValue placeholder="Método" />
      </SelectTrigger>
      <SelectContent className='text-primary'>
        {Methods.map(item => (
          <SelectItem value={item.value} className={cn(item.className, 'font-medium')}>{item.value}</SelectItem>
        ))}
      </SelectContent>
    </Select>

  )
}

function EndpointInsertVariable({
  onClick,
  index,
}: { onClick: (value: string) => void, index: number }) {

  function isSelectionInsidePlaceholder(value: string, selectionStart: number, selectionEnd: number) {
    const lastOpen = value.lastIndexOf("{{", Math.max(0, selectionEnd - 1));
    if (lastOpen === -1) return false;
    const closeAfterOpen = value.indexOf("}}", lastOpen + 2);
    if (closeAfterOpen === -1) return false;
    return closeAfterOpen >= selectionStart;
  }

  const handleInsertVariable = useCallback(() => {
    const input = document.getElementById(`endpoint-input-${index}`) as HTMLInputElement;
    if (!input) return;

    const start = input.selectionStart ?? 0;
    const end = input.selectionEnd ?? 0;
    const value = input.value;

    if (isSelectionInsidePlaceholder(value, start, end)) {
      return;
    }

    const before = value.slice(0, start);
    const after = value.slice(end);
    const newValue = `${before}{{}}${after}`;
    onClick(newValue);

    setTimeout(() => {
      input.focus();
      const newCursorPos = start + 2;
      input.setSelectionRange(newCursorPos, newCursorPos);
    }, 0);
  }, [onClick, index]);

  return (
    <InputGroupAddon align="inline-end">
      <InputGroupButton
        type="button"
        variant="ghost"
        onClick={handleInsertVariable}
        title="Inserir variável"
      >
        <CodeXml />
      </InputGroupButton>
    </InputGroupAddon>
  )
}

export function ServerEndpointForm({ form }: { form: UseFormReturn<FullServerValues> }) {
  const { control, setValue, register, watch } = form;
  const { fields, append, remove } = useFieldArray({ control, name: "endpoints" });
  
  const [testResults, setTestResults] = useState<Record<number, any>>({});
  const [testing, setTesting] = useState<Record<number, boolean>>({});
  
  const addressWatcher = watch('server.address');

  const testEndpoint = async (index: number) => {
      const ep = form.getValues(`endpoints.${index}`);
      if (!ep.path) {
          alert('Preencha a URL base do servidor na aba Geral e o caminho do endpoint antes de testar!');
          return;
      }
      
      setTesting(prev => ({...prev, [index]: true}));
      setTestResults(prev => ({...prev, [index]: undefined}));
      
      try {
          // Substituição simples para sandbox (em vez de pedir preenchimento de params)
          const testPath = ep.path.replace(/\{\{[^}]+\}\}/g, 'test_sandbox_param');
          const fullUrl = `${addressWatcher}${testPath}`;
          
          let response: any;
          try {
              const res = await axios({
                  url: fullUrl,
                  method: ep.method,
                  headers: headersArrayToRecord(ep.headers) as any,
              });
              response = { status: res.status, headers: res.headers, data: res.data };
          } catch(err: any) {
              response = { 
                  error: err.message, 
                  response: err.response ? { status: err.response.status, data: err.response.data } : 'No response (CORS ou erro de rede)'
              };
          }
          setTestResults(prev => ({...prev, [index]: response}));
      } catch (err) {
          console.error(err);
      } finally {
          setTesting(prev => ({...prev, [index]: false}));
      }
  };

  return (
    <div className="gap-6 pt-4">
      <div className="col-span-3 space-y-4">
        <div className="space-y-4">
          <h3 className="text-secondary text-lg font-medium">Configurar Endpoints</h3>
          {fields.map((field, index) => (
            <div key={field.id} className="p-4 border border-white/[0.08] bg-white/[0.02] rounded-xl flex flex-col gap-4 relative">
              <Button type="button" variant="ghost" size="icon" className="absolute top-2 right-2 text-secondary hover:text-destructive" onClick={() => remove(index)}>
                <Trash className="h-4 w-4" />
              </Button>
              <Controller
                control={form.control}
                name={`endpoints.${index}.name`}
                render={({ field, fieldState }) => (
                  <div className='space-y-2 pr-8'>
                    <Label className='text-primary'>Nome</Label>
                    <Input
                      placeholder="Nome do endpoint"
                      {...field}
                      className={cn(fieldState.error && "border-destructive", "text-primary border-white/[0.1] bg-white/[0.04]")}
                    />
                    {fieldState.error && <p className="text-destructive text-sm mt-1">{fieldState.error.message}</p>}
                  </div>
                )} />
              <div>
                <div className="flex items-center gap-2">
                  <InputGroup className="border border-white/[0.1]">
                    <InputGroupAddon className="border-r border-white/[0.1]">
                      <EndpointSelect defaultValue={field.method} onChange={(value) => {
                        register(`endpoints.${index}.method`).onChange({ target: { value } })
                      }} />
                    </InputGroupAddon>
                    <InputGroupInput
                      id={`endpoint-input-${index}`}
                      {...register(`endpoints.${index}.path`)}
                      placeholder="/"
                      className="text-primary grow bg-transparent"
                    />
                    <EndpointInsertVariable index={index} onClick={(value) => {
                      setValue(`endpoints.${index}.path`, value)
                    }} />
                  </InputGroup>
                </div>
                {form.formState.errors.endpoints?.[index]?.path && <p className="text-sm text-red-500 mt-1">{form.formState.errors.endpoints?.[index]?.path?.message}</p>}
              </div>

              {/* Sandbox Test Action */}
              <div className="pt-4 border-t border-white/[0.05] mt-2">
                  <div className="flex items-center justify-between">
                      <Label className="text-emerald-500 flex items-center gap-2">
                          Sandbox Test
                      </Label>
                      <Button 
                          type="button" 
                          size="sm" 
                          variant="secondary" 
                          className="bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20"
                          onClick={() => testEndpoint(index)}
                          disabled={testing[index]}
                      >
                          <Play size={14} className="mr-2" />
                          {testing[index] ? 'Testando...' : 'Testar Endpoint do Front'}
                      </Button>
                  </div>
                  {testResults[index] && (
                      <div className="mt-4 bg-[#000] p-4 rounded-lg overflow-x-auto border border-white/[0.08] text-sm max-h-[300px] overflow-y-auto">
                          <JsonViewMain
                            draggableValue={false}
                            pProps={{
                                className: cn(
                                    "border-none break-all w-full text-sm py-1 px-2 rounded hover:bg-accent"
                                )
                            }}
                            data={testResults[index]}
                          />
                      </div>
                  )}
              </div>
            </div>
          ))}
          <Button className='border-emerald-500 text-emerald-400 font-medium' type="button" variant="outline" onClick={() => append({ path: '', method: "GET", name: '', description: '' })}>
            + Adicionar Endpoint
          </Button>
        </div>
      </div>
    </div>
  );
}

