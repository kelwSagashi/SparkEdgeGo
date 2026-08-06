/**
 * Instance Fallback Tab - Placeholder
 * Configure fallback behavior for failed sends
 */

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface InstanceFallbackTabProps {
  data: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceFallbackTab({
  data,
  onChange,
  errors,
}: InstanceFallbackTabProps) {
  return (
    <Card className="bg-zinc-900/50 border-zinc-800">
      <CardHeader>
        <CardTitle className="text-lg">Fallback</CardTitle>
        <CardDescription>
          Configure o que fazer quando o envio para o servidor falhar
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

