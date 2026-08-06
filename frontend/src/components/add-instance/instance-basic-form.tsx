import { useFormContext, Controller } from "react-hook-form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import TagInput from "@/components/tag-input";
// instance-form-styles removed — use design tokens directly
import type {
  DeviceReturningValues,
  ProjectReturningValues,
} from "@/types/db";
import type { InstanceFormValues } from "./instance-form.schemas";
import { useAuthStore } from "@/stores/auth-store";
import { useShallow } from "zustand/react/shallow";

interface InstanceBasicFormProps {
  projects: ProjectReturningValues[];
  devices: DeviceReturningValues[];
}

export function InstanceBasicForm({
  projects,
  devices,
}: InstanceBasicFormProps) {
  const {
    control,
    register,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();
  const [project] = useAuthStore(useShallow((s) => [s.project]));

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-6">
        {/* Nome */}
        <div className="space-y-2">
          <Label htmlFor="name" className="text-primary font-medium">Nome da Instância *</Label>
          <Input
            id="name"
            placeholder="Ex: Monitoramento Solar"
            {...register("name")}
            className={`text-primary placeholder:text-secondary ${errors.name ? "border-destructive" : ""}`}
          />
          {errors.name && (
            <p className="text-sm text-destructive">{errors.name.message}</p>
          )}
        </div>

        {/* Descrição */}
        <div className="space-y-2">
          <Label htmlFor="description" className="text-primary font-medium">Descrição</Label>
          <Textarea
            id="description"
            placeholder="Descreva o propósito desta instância..."
            rows={4}
            {...register("description")}
            className="text-primary placeholder:text-secondary"
          />
        </div>

        {/* Projeto */}
        <div className="space-y-2">
          <Label htmlFor="project_id" className="text-primary font-medium">Projeto *</Label>
          <Controller
            name="project_id"
            control={control}
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id="project_id" className="text-primary placeholder:text-secondary">
                  <SelectValue placeholder="Selecione um projeto" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup className="text-primary">
                    <SelectLabel className="text-secondary">Projetos</SelectLabel>
                    {projects.map((project) => (
                      <SelectItem key={project.id} value={project.id}>
                        {project.name}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            )}
          />
          {errors.project_id && (
            <p className="text-sm text-destructive">{errors.project_id.message}</p>
          )}
        </div>

        {/* Dispositivo */}
        <div className="space-y-2">
          <Label htmlFor="device_id" className="text-primary font-medium">Dispositivo (Opcional)</Label>
          <Controller
            name="device_id"
            control={control}
            render={({ field }) => (
              <Select
                value={field.value || ""}
                onValueChange={(value) => field.onChange(value || null)}
              >
                <SelectTrigger id="device_id" className="text-primary">
                  <SelectValue placeholder="Nenhum dispositivo selecionado" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel className="text-secondary">Dispositivos Disponíveis</SelectLabel>
                    {devices.map((device) => (
                      <SelectItem className="text-primary" key={device.id} value={device.id}>
                        {device.name}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            )}
          />
        </div>

        {/* Incluir dados do dispositivo */}
        <div className="space-y-2">
          <div className="flex items-center space-x-2">
            <Controller
              name="includeDeviceData"
              control={control}
              render={({ field }) => (
                <Checkbox
                  id="includeDeviceData"
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              )}
            />
            <Label
              htmlFor="includeDeviceData"
              className="font-medium text-primary cursor-pointer"
            >
              Incluir dados do dispositivo no envio
            </Label>
          </div>
          <p className="text-sm text-secondary">
            Quando habilitado, os dados do dispositivo serão incluídos junto com
            os dados do script.
          </p>
        </div>

        {/* Tags */}
        <div className="space-y-2">
          <Label className="text-primary font-medium">Tags</Label>
          <Controller
            name="tags"
            control={control}
            render={({ field }) => (
              <TagInput
                value={field.value as any}
                onChange={field.onChange}
                projectId={project?.id}
              />
            )}
          />
        </div>
      </div>
    </ScrollArea>
  );
}

