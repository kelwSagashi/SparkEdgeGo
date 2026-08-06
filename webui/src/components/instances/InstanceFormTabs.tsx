/**
 * Instance Form Component
 * Simplified form for creating instances
 */

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

interface InstanceFormProps {
  projectId: string;
  instanceId?: string;
  onSave: (data: any) => Promise<void>;
  onCancel: () => void;
}

export function InstanceFormTabs(props: InstanceFormProps) {
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    project_id: props.projectId,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    try {
      setLoading(true);
      setError(null);
      await props.onSave(formData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save instance");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Create Instance</CardTitle>
        <CardDescription>Configure a new automation instance</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500 rounded text-red-400 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="text-sm font-medium">Name</label>
          <Input
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="Instance name"
            className="mt-1"
          />
        </div>

        <div>
          <label className="text-sm font-medium">Description</label>
          <Textarea
            value={formData.description}
            onChange={(e) =>
              setFormData({ ...formData, description: e.target.value })
            }
            placeholder="Instance description"
            className="mt-1"
          />
        </div>

        <div className="flex gap-2 pt-4">
          <Button onClick={props.onCancel} variant="outline" disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={loading || !formData.name}>
            {loading ? "Saving..." : "Save"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export default InstanceFormTabs;

