import { Server } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ServersEmptyStateProps {
  onAddServer: () => void;
}

export function ServersEmptyState({ onAddServer }: ServersEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-4">
      <div className="rounded-full bg-primary-foreground/20 p-6 mb-6">
        <Server className="h-12 w-12 text-primary" />
      </div>
      <h3 className="text-xl font-semibold text-primary mb-2">
        No servers found
      </h3>
      <p className="text-secondary text-center mb-6 max-w-md">
        Get started by connecting your first server. You can integrate REST APIs, databases, and other services.
      </p>
      <Button onClick={onAddServer} size="lg">
        Connect your first server
      </Button>
    </div>
  );
}

