/**
 * Instance Basic Info Tab
 * Name, description, and tags configuration
 */

import { useCallback, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";

interface InstanceBasicInfoTabProps {
  projectId: string;
  data: any;
  onChange: (data: any) => void;
  errors: Record<string, string>;
}

export default function InstanceBasicInfoTab({
  projectId,
  data,
  onChange,
  errors,
}: InstanceBasicInfoTabProps) {
  const [tagInput, setTagInput] = useState("");

  const handleNameChange = useCallback(
    (value: string) => {
      onChange({
        ...data,
        name: value,
      });
    },
    [data, onChange],
  );

  const handleDescriptionChange = useCallback(
    (value: string) => {
      onChange({
        ...data,
        description: value,
      });
    },
    [data, onChange],
  );

  const handleAddTag = useCallback(
    (tag: string) => {
      const trimmed = tag.trim().toLowerCase();
      if (trimmed && !data.tags.includes(trimmed)) {
        onChange({
          ...data,
          tags: [...data.tags, trimmed],
        });
        setTagInput("");
      }
    },
    [data, onChange],
  );

  const handleRemoveTag = useCallback(
    (tag: string) => {
      onChange({
        ...data,
        tags: data.tags.filter((t: string) => t !== tag),
      });
    },
    [data, onChange],
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        handleAddTag(tagInput);
      }
    },
    [tagInput, handleAddTag],
  );

  return (
    <div className="space-y-6">
      <Card className="bg-zinc-900/50 border-zinc-800">
        <CardHeader>
          <CardTitle className="text-lg">Informações Básicas</CardTitle>
          <CardDescription>
            Configure o nome, descrição e categorias da instância
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Name */}
          <div className="space-y-2">
            <Label htmlFor="instance-name" className="text-sm font-medium">
              Nome <span className="text-red-500">*</span>
            </Label>
            <Input
              id="instance-name"
              placeholder="Ex: Leitura de Temperatura"
              value={data.name}
              onChange={(e) => handleNameChange(e.target.value)}
              className={errors.name ? "border-red-500" : ""}
              maxLength={100}
            />
            {errors.name && (
              <p className="text-xs text-red-400">{errors.name}</p>
            )}
            <p className="text-xs text-zinc-400">
              {data.name.length} / 100 caracteres
            </p>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label
              htmlFor="instance-description"
              className="text-sm font-medium"
            >
              Descrição
            </Label>
            <Textarea
              id="instance-description"
              placeholder="Descreva o propósito desta instância..."
              value={data.description || ""}
              onChange={(e) => handleDescriptionChange(e.target.value)}
              maxLength={500}
              rows={4}
              className="resize-none"
            />
            <p className="text-xs text-zinc-400">
              {(data.description || "").length} / 500 caracteres
            </p>
          </div>

          {/* Tags */}
          <div className="space-y-3">
            <Label htmlFor="instance-tags" className="text-sm font-medium">
              Categorias / Tags
            </Label>

            {data.tags && data.tags.length > 0 && (
              <div className="flex flex-wrap gap-2 p-3 bg-zinc-800/50 rounded-lg">
                {data.tags.map((tag: string) => (
                  <Badge
                    key={tag}
                    variant="secondary"
                    className="gap-1.5 pl-2 pr-1 py-1"
                  >
                    {tag}
                    <button
                      onClick={() => handleRemoveTag(tag)}
                      className="hover:text-red-400 transition-colors"
                      aria-label={`Remover tag ${tag}`}
                    >
                      <X size={14} />
                    </button>
                  </Badge>
                ))}
              </div>
            )}

            <div className="flex gap-2">
              <Input
                id="instance-tags"
                placeholder="Digite uma tag e pressione Enter..."
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleKeyDown}
                maxLength={30}
              />
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => handleAddTag(tagInput)}
                disabled={!tagInput.trim()}
              >
                Adicionar
              </Button>
            </div>

            <p className="text-xs text-zinc-400">
              Use tags para organizar e filtrar suas instâncias
            </p>
          </div>

          {/* Info Box */}
          <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 space-y-2">
            <h4 className="text-sm font-medium text-blue-300">💡 Dica</h4>
            <p className="text-sm text-blue-200">
              Informações básicas bem descritas ajudam na organização e
              compreensão do que cada instância faz. Você pode atualizar estes
              dados a qualquer momento.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

