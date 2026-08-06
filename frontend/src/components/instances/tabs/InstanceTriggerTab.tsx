/**
 * Instance Trigger Tab - Placeholder
 * Configure when instance executes
 */

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface InstanceTriggerTabProps {
  data: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceTriggerTab({
  data,
  onChange,
  errors,
}: InstanceTriggerTabProps) {
  return (
    <Card className="bg-zinc-900/50 border-zinc-800">
      <CardHeader>
        <CardTitle className="text-lg">Acionamento</CardTitle>
        <CardDescription>
          Configure como e quando a instância será executada
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

