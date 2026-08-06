import { useFormContext, Controller } from "react-hook-form";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Card } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
// instance-form-styles removed — use design tokens directly
import type { InstanceFormValues } from "./instance-form.schemas";

export function InstanceTriggerForm() {
  const {
    control,
    register,
    watch,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();

  const triggerType = watch("triggerType");

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-6">
        {/* Tipo de Trigger */}
        <div className="space-y-2">
          <Label htmlFor="triggerType" className="text-primary font-medium">Tipo de Trigger *</Label>
          <Controller
            name="triggerType"
            control={control}
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id="triggerType" className="text-primary">
                  <SelectValue placeholder="Selecione o tipo de trigger" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup className="text-primary">
                    <SelectItem value="interval">Intervalo Agendado</SelectItem>
                    <SelectItem value="webhook">Webhook (Remoto)</SelectItem>
                    <SelectItem value="interval_and_webhook">
                      Ambos (Intervalo + Webhook)
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            )}
          />
        </div>

        {/* Descrição do tipo selecionado */}
        <Card className="p-4 bg-accent border-border">
          <p className="text-sm text-secondary">
            {triggerType === "interval" &&
              "A instância será executada automaticamente em intervalos regulares."}
            {triggerType === "webhook" &&
              "A instância será executada quando um webhook remoto for enviado para a URL gerada."}
            {triggerType === "interval_and_webhook" &&
              "A instância será executada tanto em intervalos regulares quanto quando um webhook remoto for enviado."}
          </p>
        </Card>

        {/* Configuração de Intervalo */}
        {(triggerType === "interval" ||
          triggerType === "interval_and_webhook") && (
          <div className="space-y-2">
            <Label htmlFor="interval_seconds" className="text-primary font-medium">
              Intervalo de Execução (segundos) *
            </Label>
            <Input
              id="interval_seconds"
              type="number"
              min="10"
              placeholder="Ex: 300 (5 minutos)"
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
            <p className="text-sm text-secondary">
              Mínimo de 60 segundos (1 minuto). Recomendado: 300 segundos (5
              minutos).
            </p>
          </div>
        )}

        {/* Configuração de Webhook */}
        {(triggerType === "webhook" ||
          triggerType === "interval_and_webhook") && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="webhook_path" className="text-primary font-medium">Caminho do Webhook *</Label>
              <Input
                id="webhook_path"
                placeholder="Ex: /webhook/energy-monitor"
                {...register("triggerConfig.webhook_path")}
                className="text-primary"
              />
              {errors.triggerConfig?.webhook_path && (
                <p className="text-sm text-destructive">
                  {errors.triggerConfig.webhook_path.message}
                </p>
              )}
              <p className="text-sm text-secondary">
                O caminho será exposto em http://seu-servidor{"{"}webhook_path
                {"}"}.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="webhook_secret" className="text-primary font-medium">
                Secret do Webhook (Opcional)
              </Label>
              <Input
                id="webhook_secret"
                type="password"
                placeholder="Opcional: segredo para o webhook"
                {...register("triggerConfig.webhook_secret")}
                className="text-primary"
              />
              <p className="text-sm text-secondary">
                Use um secret para validar requisições provenientes da sua
                fonte.
              </p>
            </div>

            <Card className="p-4 bg-muted/40 border-border">
              <p className="text-sm font-medium text-foreground mb-2">
                Informações do Webhook
              </p>
              <div className="space-y-1 text-sm text-secondary">
                <p>
                  URL:{" "}
                  <code className="text-foreground bg-muted/50 px-1 rounded">
                    http://seu-servidor/webhook
                    {watch("triggerConfig.webhook_path")}
                  </code>
                </p>
                <p>
                  Método: <code className="text-foreground">POST</code>
                </p>
                <p>
                  Content-Type:{" "}
                  <code className="text-foreground">application/json</code>
                </p>
              </div>
            </Card>
          </div>
        )}

        {/* Configurações Adicionais */}
        <div className="space-y-4 pt-4 border-t border-border">
          <Label className="text-base text-primary font-semibold">Configurações Adicionais</Label>
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
              <Label
                htmlFor="save_execution_on_server"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 text-primary"
              >
                Salvar Execução no Servidor
              </Label>
              <p className="text-xs text-secondary">
                Se ativado, cada execução será registrada no histórico do
                servidor para auditoria e logs.
              </p>
            </div>
          </div>
        </div>

        {/* Informações gerais */}
        <Card className="p-4 bg-muted border-muted-foreground/20 text-secondary">
          <p className="text-sm text-secondary">
            💡 <strong>Dica:</strong> Configure as credenciais e permissões de
            rede adequadamente para que a execução remota funcione corretamente
            em sua infraestrutura.
          </p>
        </Card>
      </div>
    </ScrollArea>
  );
}

