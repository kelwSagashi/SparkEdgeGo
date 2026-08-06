/**
 * Instance Destinations Tab - Placeholder
 * Configure where to send data
 */

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface InstanceDestinationsTabProps {
  projectId: string;
  data: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceDestinationsTab({
  projectId,
  data,
  onChange,
  errors,
}: InstanceDestinationsTabProps) {
  return (
    <Card className="bg-zinc-900/50 border-zinc-800">
      <CardHeader>
        <CardTitle className="text-lg">Destinos</CardTitle>
        <CardDescription>
          Configure onde os dados coletados serão enviados
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

