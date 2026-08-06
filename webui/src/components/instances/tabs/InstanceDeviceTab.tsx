/**
 * Instance Device Tab - Placeholder
 * Link optional device to instance
 */

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";

interface InstanceDeviceTabProps {
  projectId: string;
  data: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceDeviceTab({
  projectId,
  data,
  onChange,
  errors,
}: InstanceDeviceTabProps) {
  return (
    <Card className="bg-zinc-900/50 border-zinc-800">
      <CardHeader>
        <CardTitle className="text-lg">Dispositivo</CardTitle>
        <CardDescription>
          Vincule um dispositivo à instância (opcional)
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

