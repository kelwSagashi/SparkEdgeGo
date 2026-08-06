/**
 * Instance Data Mapping Tab - Placeholder
 * Map script output and device data to destination schema
 */

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface InstanceDataMappingTabProps {
  instanceData: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceDataMappingTab({
  instanceData,
  onChange,
  errors,
}: InstanceDataMappingTabProps) {
  return (
    <Card className="bg-zinc-900/50 border-zinc-800">
      <CardHeader>
        <CardTitle className="text-lg">Mapeamento de Dados</CardTitle>
        <CardDescription>
          Mapeie saída do script e dados do dispositivo para o schema do
          servidor
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-zinc-400">
          Componente em desenvolvimento...
        </p>
      </CardContent>
    </Card>
  );
}

