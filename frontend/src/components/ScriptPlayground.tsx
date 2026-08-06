import { cn } from "@/lib/utils";
import { JsonViewMain } from "./json-view/json-view";
import { Button } from "./ui/button";
import { Checkbox } from "./ui/checkbox";
import { Input } from "./ui/input";
import { Label } from "./ui/label";

export default function ScriptPlayground({
    handleRun,
    setInputs,
    inputs,
    output,
    schema,
    loading,
    handleSaveStdoutSchema,
    handleSaveStderrSchema
}: {
    loading: boolean;
    schema: any;
    inputs: Record<string, any>;
    output: any;
    handleRun: () => void;
    handleSaveStdoutSchema?: (data: any) => void;
    handleSaveStderrSchema?: (data: any) => void;
    setInputs: React.Dispatch<React.SetStateAction<Record<string, any>>>
}) {
    
    return (
        <div className="flex flex-1 overflow-hidden">
            <div className="w-1/2 p-4 border-r border-white/[0.08] flex flex-col gap-4 overflow-y-auto">
                <div className="flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-white">Inputs</h3>
                </div>
                {loading && !schema ? (
                        <div className="flex justify-center p-4"><div className="w-4 h-4 border-2 animate-spin rounded-full border-t-transparent"></div></div>
                ) : (
                    <div className="space-y-4">
                        {schema?.inputs?.length > 0 ? schema.inputs.map((inp: any) => (
                            <div key={inp.name} className="flex flex-col gap-1.5">
                                <Label className="text-zinc-400 text-xs">{inp.name} {inp.required && <span className="text-red-400">*</span>}</Label>
                                {inp.type === 'boolean' ? (
                                    <Checkbox 
                                        checked={inputs[inp.name] || false}
                                        onCheckedChange={v => setInputs(prev => ({...prev, [inp.name]: !!v}))}
                                    />
                                ) : (
                                    <Input
                                        type={inp.type === 'number' ? 'number' : 'text'}
                                        placeholder={inp.type}
                                        value={inputs[inp.name] || ''}
                                        onChange={e => setInputs(prev => ({...prev, [inp.name]: e.target.value}))}
                                        className="bg-white/[0.04] border-white/[0.1] text-white text-sm h-8"
                                    />
                                )}
                            </div>
                        )) : <p className="text-xs text-zinc-500">Sem inputs definidos</p>}
                    </div>
                )}
                <Button 
                    onClick={handleRun} 
                    disabled={loading || !schema}
                    className="mt-4 bg-violet-600 hover:bg-violet-700 text-white w-full"
                >
                    {loading ? 'Executando...' : 'Executar'}
                </Button>
            </div>
            <div className="w-1/2 p-4 bg-[#0c0c0e] overflow-y-auto flex flex-col">
                <div className="flex items-center justify-between mb-4">
                    <h3 className="text-sm font-semibold text-white">Saída</h3>
                    <div className="flex gap-2">
                        {output?.stdout && handleSaveStdoutSchema && (
                            <Button 
                                variant="outline" 
                                size="sm" 
                                className="h-7 text-[10px] border-violet-500/30 text-violet-400 hover:bg-violet-500/10"
                                onClick={() => handleSaveStdoutSchema(output.stdout)}
                            >
                                Salvar Stdout
                            </Button>
                        )}
                        {output?.stderr && handleSaveStderrSchema && (
                            <Button 
                                variant="outline" 
                                size="sm" 
                                className="h-7 text-[10px] border-red-500/30 text-red-400 hover:bg-red-500/10"
                                onClick={() => handleSaveStderrSchema(output.stderr)}
                            >
                                Salvar Stderr
                            </Button>
                        )}
                    </div>
                </div>
                {output ? (
                    <JsonViewMain
                        pProps={{
                        className: cn(
                            "border-none break-all w-full text-sm py-1 px-2 rounded bg-accent"
                        )
                        }}
                        draggableValue={true}
                        rootPath="$.script"
                        data={output}
                    />
                ) : (
                    <p className="text-xs text-zinc-600 italic">Execute o script para ver os resultados aqui...</p>
                )}
            </div>
        </div>
    )
}
