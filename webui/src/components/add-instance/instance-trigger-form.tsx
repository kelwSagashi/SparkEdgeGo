import { Controller, useFormContext } from "react-hook-form";
import { Card } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import type { InstanceFormValues } from "./instance-form.schemas";

type TriggerFormProps = {
  instances?: Array<{ id: string; name?: string | null }>;
};

export function InstanceTriggerForm({ instances = [] }: TriggerFormProps) {
  const {
    control,
    register,
    watch,
    setValue,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();

  const triggerType = watch("triggerType");
  const executionMode = watch("executionMode");
  const dependsOn = watch("dependsOn") || [];

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="triggerType" className="text-primary font-medium">
              Tipo de Trigger
            </Label>
            <Controller
              name="triggerType"
              control={control}
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id="triggerType" className="text-primary">
                    <SelectValue placeholder="Selecione o trigger" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup className="text-primary">
                      <SelectItem value="interval">Intervalo Agendado</SelectItem>
                      <SelectItem value="webhook">Webhook</SelectItem>
                      <SelectItem value="interval_and_webhook">
                        Intervalo + Webhook
                      </SelectItem>
                      <SelectItem value="event">Evento Interno</SelectItem>
                      <SelectItem value="mqtt">Mensagem MQTT</SelectItem>
                      <SelectItem value="state_change">Mudanca de Estado</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              )}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="executionMode" className="text-primary font-medium">
              Modo de Execucao
            </Label>
            <Controller
              name="executionMode"
              control={control}
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id="executionMode" className="text-primary">
                    <SelectValue placeholder="Selecione o modo" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup className="text-primary">
                      <SelectItem value="sequential">Sequencial</SelectItem>
                      <SelectItem value="parallel">Paralelo</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              )}
            />
          </div>
        </div>

        <Card className="p-4 bg-accent border-border space-y-2">
          <p className="text-sm text-secondary">
            {triggerType === "interval" &&
              "Executa automaticamente em intervalos regulares."}
            {triggerType === "webhook" &&
              "Executa quando receber um POST no endpoint configurado."}
            {triggerType === "interval_and_webhook" &&
              "Executa por agenda e tambem por chamada remota."}
            {triggerType === "event" &&
              "Executa quando um evento interno com o nome configurado for disparado."}
            {triggerType === "mqtt" &&
              "Executa quando houver mensagem no topico MQTT configurado."}
            {triggerType === "state_change" &&
              "Executa quando um campo monitorado mudar para o valor desejado."}
          </p>
          <p className="text-xs text-secondary/80">
            Modo atual:{" "}
            <strong className="text-primary">
              {executionMode === "parallel" ? "Paralelo" : "Sequencial"}
            </strong>
          </p>
        </Card>

        {(triggerType === "interval" || triggerType === "interval_and_webhook") && (
          <div className="space-y-2">
            <Label htmlFor="interval_seconds" className="text-primary font-medium">
              Intervalo (segundos)
            </Label>
            <Input
              id="interval_seconds"
              type="number"
              min="10"
              {...register("triggerConfig.interval_seconds", {
                valueAsNumber: true,
              })}
              className="text-primary"
            />
            {errors.triggerConfig?.interval_seconds && (
              <p className="text-sm text-destructive">
                {errors.triggerConfig.interval_seconds.message}
              </p>
            )}
          </div>
        )}

        {(triggerType === "webhook" || triggerType === "interval_and_webhook") && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="webhook_path" className="text-primary font-medium">
                Caminho do Webhook
              </Label>
              <Input
                id="webhook_path"
                placeholder="/webhook/minha-instancia"
                {...register("triggerConfig.webhook_path")}
                className="text-primary"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="webhook_secret" className="text-primary font-medium">
                Segredo do Webhook
              </Label>
              <Input
                id="webhook_secret"
                type="password"
                placeholder="Opcional"
                {...register("triggerConfig.webhook_secret")}
                className="text-primary"
              />
            </div>
          </div>
        )}

        {triggerType === "event" && (
          <div className="space-y-2">
            <Label htmlFor="event_name" className="text-primary font-medium">
              Nome do Evento
            </Label>
            <Input
              id="event_name"
              placeholder="sensor.updated"
              {...register("triggerConfig.event_name")}
              className="text-primary"
            />
          </div>
        )}

        {triggerType === "mqtt" && (
          <div className="space-y-2">
            <Label htmlFor="mqtt_topic" className="text-primary font-medium">
              Topico MQTT
            </Label>
            <Input
              id="mqtt_topic"
              placeholder="sparkedge/devices/+/events"
              {...register("triggerConfig.mqtt_topic")}
              className="text-primary"
            />
          </div>
        )}

        {triggerType === "state_change" && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="state_field" className="text-primary font-medium">
                Campo Monitorado
              </Label>
              <Input
                id="state_field"
                placeholder="$.device.status"
                {...register("triggerConfig.state_field")}
                className="text-primary"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="state_equals" className="text-primary font-medium">
                Valor Esperado
              </Label>
              <Input
                id="state_equals"
                placeholder="online"
                {...register("triggerConfig.state_equals")}
                className="text-primary"
              />
            </div>
          </div>
        )}

        <div className="space-y-3 pt-2 border-t border-border">
          <Label className="text-base text-primary font-semibold">
            Dependencias e Workflow
          </Label>
          <div className="space-y-2">
            <Label htmlFor="dependsOnCsv" className="text-primary font-medium">
              Dependencias entre Instancias
            </Label>
            <Textarea
              id="dependsOnCsv"
              value={dependsOn.join(", ")}
              onChange={(event) => {
                const next = event.target.value
                  .split(",")
                  .map((item) => item.trim())
                  .filter(Boolean);
                setValue("dependsOn", next, { shouldDirty: true, shouldValidate: true });
              }}
              placeholder="id-instancia-a, id-instancia-b"
              className="text-primary min-h-20"
            />
            <p className="text-xs text-secondary">
              Use IDs de instancias para definir precedencia de execucao. As
              dependencias ficam registradas na configuracao da instancia.
            </p>
          </div>

          {instances.length > 0 && (
            <Card className="p-3 bg-muted/30 border-border">
              <p className="text-xs font-semibold text-primary mb-2">
                Instancias disponiveis para dependencia
              </p>
              <div className="flex flex-wrap gap-2">
                {instances.slice(0, 12).map((instance) => (
                  <button
                    key={instance.id}
                    type="button"
                    className="px-2 py-1 rounded border border-border text-xs text-secondary hover:text-primary hover:border-primary/30"
                    onClick={() => {
                      if (dependsOn.includes(instance.id)) {
                        setValue(
                          "dependsOn",
                          dependsOn.filter((item) => item !== instance.id),
                          { shouldDirty: true, shouldValidate: true },
                        );
                        return;
                      }
                      setValue("dependsOn", [...dependsOn, instance.id], {
                        shouldDirty: true,
                        shouldValidate: true,
                      });
                    }}
                  >
                    {instance.name || instance.id}
                  </button>
                ))}
              </div>
            </Card>
          )}

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center space-x-2">
              <Controller
                name="orchestrationConfig.workflow_enabled"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="workflow_enabled"
                    checked={!!field.value}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="workflow_enabled" className="text-primary">
                Ativar workflow visual
              </Label>
            </div>

            <div className="flex items-center space-x-2">
              <Controller
                name="orchestrationConfig.allow_partial_success"
                control={control}
                render={({ field }) => (
                  <Checkbox
                    id="allow_partial_success"
                    checked={field.value !== false}
                    onCheckedChange={field.onChange}
                  />
                )}
              />
              <Label htmlFor="allow_partial_success" className="text-primary">
                Permitir sucesso parcial
              </Label>
            </div>

            <div className="space-y-2">
              <Label htmlFor="debounce_seconds" className="text-primary font-medium">
                Debounce (segundos)
              </Label>
              <Input
                id="debounce_seconds"
                type="number"
                min="0"
                {...register("orchestrationConfig.debounce_seconds", {
                  valueAsNumber: true,
                })}
                className="text-primary"
              />
            </div>
          </div>
        </div>

        <div className="space-y-4 pt-4 border-t border-border">
          <Label className="text-base text-primary font-semibold">
            Configuracoes Adicionais
          </Label>
          <div className="flex items-center space-x-2">
            <Controller
              name="triggerConfig.save_execution_on_server"
              control={control}
              render={({ field }) => (
                <Checkbox
                  id="save_execution_on_server"
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              )}
            />
            <div className="grid gap-1.5 leading-none">
              <Label htmlFor="save_execution_on_server" className="text-primary">
                Salvar execucao no historico
              </Label>
              <p className="text-xs text-secondary">
                Mantem auditoria e facilita diagnostico por etapa.
              </p>
            </div>
          </div>
        </div>
      </div>
    </ScrollArea>
  );
}
