import React, { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Play, Loader2, AlertCircle, CheckCircle2, Settings2 } from 'lucide-react';
import { scriptsApi } from '@/rest-api-client/scripts.service';
import { JsonViewMain } from './json-view/json-view';
import { AlertDialog, AlertDialogDescription, AlertDialogTitle, AlertDialogContent, AlertDialogHeader, AlertDialogFooter, AlertDialogAction } from './ui/alert-dialog';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Checkbox } from './ui/checkbox';
import { cn } from '@/lib/utils';
import ScriptPlayground from './ScriptPlayground';

interface InstancePlaygroundProps {
  scriptId: string;
  deviceId?: string;
  onOutputReceived: (output: any) => void;
  initialOutput?: any;
}

export function InstancePlayground({ scriptId, deviceId, onOutputReceived, initialOutput }: InstancePlaygroundProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [output, setOutput] = useState<any>(initialOutput || null);
  const [schema, setSchema] = useState<any>(null);
  const [inputs, setInputs] = useState<Record<string, any>>({});

  useEffect(() => {
    if (!scriptId) {
      setSchema(null);
      return;
    }

    const fetchSchema = async () => {
      setLoading(true);
      try {
        const res: any = await scriptsApi.get(scriptId);
        // Assuming schema is in script.schema_config as per ScriptHub logic
        const scriptData = res.data;
        setSchema(scriptData?.schema_config || { inputs: [], outputs: [] });
      } catch (err) {
        console.error("Failed to fetch script schema", err);
      } finally {
        setLoading(false);
      }
    };
    fetchSchema();
  }, [scriptId]);

  const handleRun = async () => {
    if (!scriptId) return;
     setLoading(true);
    try {
      const res: any = await scriptsApi.runPlayground({ script_id: scriptId, inputs });
      console.log(inputs, res)
      setOutput(res.data);
    } catch(err: any) {
      setOutput({ stdout: null, stderr: err.message });
    } finally {
        setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <ScriptPlayground
        handleRun={handleRun}
        setInputs={setInputs}
        inputs={inputs}
        output={output}
        schema={schema}
        loading={loading}
      />

      <AlertDialog open={!!error} onOpenChange={() => setError(null)}>
        <AlertDialogContent className="bg-zinc-950 border-white/[0.1]">
          <AlertDialogHeader>
            <div className="flex items-center gap-3 text-red-400 mb-2">
              <AlertCircle size={24} />
              <AlertDialogTitle>Erro na Execução</AlertDialogTitle>
            </div>
            <AlertDialogDescription className="text-zinc-400 text-sm font-mono p-4 bg-red-400/5 border border-red-400/10 rounded-lg">
              {error}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction className="bg-zinc-800 hover:bg-zinc-700 text-white">Fechar</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

