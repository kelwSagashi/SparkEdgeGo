import { Controller, useFormContext } from "react-hook-form";
import { AlertCircle, Database } from "lucide-react";

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

import type { InstanceFormValues } from "./instance-form.schemas";

export function InstanceFallbackForm() {
  const {
    control,
    register,
    watch,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();

  const fallbackEnabled = watch("fallbackConfig.enabled");
  const fallbackStrategy = watch("fallbackConfig.strategy");
  const errorAction = watch("errorConfig.action");

  return (
    <ScrollArea className="h-full">
      <div className="pr-4 space-y-6">
        <div className="space-y-4">
          <h3 className="font-medium text-primary flex items-center gap-2">
            <Database size={18} /> Configuracao de Fallback
          </h3>
          <p className="text-sm text-secondary">
            Define o comportamento quando o envio de dados para o servidor falhar.
          </p>

          <div className="flex items-center space-x-2">
            <Controller
              name="fallbackConfig.enabled"
              control={control}
              render={({ field }) => (
                <Checkbox
                  id="fallbackEnabled"
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              )}
            />
            <Label
              htmlFor="fallbackEnabled"
              className="font-medium text-primary cursor-pointer"
            >
              Habilitar armazenamento local de fallback
            </Label>
          </div>

          {fallbackEnabled && (
            <>
              <div className="space-y-2">
                <Label htmlFor="fallbackStrategy" className="text-primary font-medium">
                  Estrategia de Fallback
                </Label>
                <Controller
                  name="fallbackConfig.strategy"
                  control={control}
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id="fallbackStrategy" className="text-primary">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup className="text-primary">
                          <SelectItem value="background_job">
                            Armazenamento Local (Melhor Esforco)
                          </SelectItem>
                          <SelectItem value="active_queue">
                            Fila Prioritaria (Garantia de Entrega)
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  )}
                />
              </div>

              <Card className="p-3 bg-muted border-muted-foreground/20 text-secondary">
                <p className="text-xs text-secondary">
                  {fallbackStrategy === "background_job" ? (
                    <>
                      <strong>Melhor Esforco:</strong> Os dados sao salvos localmente e o
                      sistema tentara reenviar periodicamente em segundo plano.
                    </>
                  ) : (
                    <>
                      <strong>Garantia de Entrega:</strong> Os dados sao colocados em uma
                      fila persistente e ordenada para reduzir perda em falhas.
                    </>
                  )}
                </p>
              </Card>

              <div className="space-y-2">
                <Label htmlFor="fallbackRetryInterval" className="text-primary font-medium">
                  Intervalo de Retry (segundos)
                </Label>
                <Input
                  id="fallbackRetryInterval"
                  type="number"
                  min="60"
                  step="60"
                  placeholder="300"
                  {...register("fallbackConfig.retry_interval_seconds", {
                    valueAsNumber: true,
                  })}
                  className="text-primary"
                />
                <p className="text-xs text-secondary">
                  Tempo de espera entre tentativas de reenvio. Minimo: 60 segundos.
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="fallbackMaxRetries" className="text-primary font-medium">
                  Maximo de Tentativas (opcional)
                </Label>
                <Input
                  id="fallbackMaxRetries"
                  type="number"
                  min="1"
                  placeholder="Deixe vazio para ilimitado"
                  {...register("fallbackConfig.max_retries", {
                    setValueAs: (value) =>
                      value === "" || Number.isNaN(value) ? null : parseInt(value, 10),
                  })}
                  className="text-primary"
                />
                <p className="text-xs text-secondary">
                  Numero maximo de vezes para tentar reenviar.
                </p>
              </div>
            </>
          )}
        </div>

        <div className="border-t pt-6 space-y-4">
          <h3 className="font-bold text-xs uppercase text-primary tracking-widest bg-primary/10 w-fit px-2 py-0.5 rounded flex items-center gap-2">
            <AlertCircle size={14} /> Tratamento de Erros
          </h3>
          <p className="text-sm text-secondary">
            Define o comportamento quando ocorrem erros durante a execucao do script.
          </p>

          <div className="space-y-2">
            <Label htmlFor="errorAction" className="text-primary font-medium">
              Acao em Caso de Erro
            </Label>
            <Controller
              name="errorConfig.action"
              control={control}
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id="errorAction" className="text-primary">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="log_only">Apenas Registrar (Log)</SelectItem>
                      <SelectItem value="retry">Tentar Novamente</SelectItem>
                      <SelectItem value="notify_webhook">Notificar via Webhook</SelectItem>
                      <SelectItem value="stop">Parar Execucao</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              )}
            />
          </div>

          <Card className="p-3 bg-muted border-muted-foreground/20 text-secondary">
            <p className="text-xs text-secondary">
              {errorAction === "log_only" && (
                <>
                  <strong>Log Only:</strong> Registra o erro e finaliza a execucao com
                  falha.
                </>
              )}
              {errorAction === "retry" && (
                <>
                  <strong>Retry:</strong> Tenta executar o script novamente de forma
                  automatica.
                </>
              )}
              {errorAction === "notify_webhook" && (
                <>
                  <strong>Notify Webhook:</strong> Envia notificacao para um webhook
                  externo.
                </>
              )}
              {errorAction === "stop" && (
                <>
                  <strong>Stop:</strong> Para a execucao e marca como erro.
                </>
              )}
            </p>
          </Card>

          {errorAction === "notify_webhook" && (
            <div className="space-y-2">
              <Label htmlFor="notifyUrl" className="text-primary font-medium">
                URL do Webhook de Notificacao
              </Label>
              <Input
                id="notifyUrl"
                type="url"
                placeholder="https://seu-servidor.com/errors"
                {...register("errorConfig.notify_url")}
                className="text-primary"
              />
              {errors.errorConfig?.notify_url && (
                <p className="text-sm text-destructive">
                  {errors.errorConfig.notify_url.message}
                </p>
              )}
            </div>
          )}

          {errorAction === "retry" && (
            <>
              <div className="space-y-2">
                <Label htmlFor="errorMaxRetries" className="text-primary font-medium">
                  Maximo de Retries em Erro
                </Label>
                <Input
                  id="errorMaxRetries"
                  type="number"
                  min="0"
                  placeholder="2"
                  {...register("errorConfig.max_retries", {
                    setValueAs: (value) =>
                      value === "" || Number.isNaN(value) ? null : parseInt(value, 10),
                  })}
                  className="text-primary"
                />
                <p className="text-xs text-secondary">
                  Quantidade de novas tentativas apos a primeira falha. Exemplo: 2
                  significa ate 3 execucoes no total.
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="errorRetryInterval" className="text-primary font-medium">
                  Intervalo entre Tentativas (segundos)
                </Label>
                <Input
                  id="errorRetryInterval"
                  type="number"
                  min="1"
                  placeholder="5"
                  {...register("errorConfig.retry_interval_seconds", {
                    setValueAs: (value) =>
                      value === "" || Number.isNaN(value) ? null : parseInt(value, 10),
                  })}
                  className="text-primary"
                />
                <p className="text-xs text-secondary">
                  Tempo de espera antes de tentar executar o script novamente.
                </p>
              </div>
            </>
          )}
        </div>

        <div className="border-t pt-6 space-y-4">
          <h3 className="font-medium text-primary font-bold">Status da Instancia</h3>

          <div className="flex items-center space-x-2">
            <Controller
              name="active"
              control={control}
              render={({ field }) => (
                <Checkbox
                  id="instanceActive"
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              )}
            />
            <Label
              htmlFor="instanceActive"
              className="font-medium text-primary cursor-pointer"
            >
              Instancia ativa (habilitada para execucao)
            </Label>
          </div>

          <Card className="p-3 bg-muted/40 border-border">
            <p className="text-xs text-secondary">
              <strong>Dica:</strong> Desabilite a instancia temporariamente se precisar
              fazer manutencao sem deleta-la.
            </p>
          </Card>
        </div>

        <Card className="p-4 bg-muted/40 border-border space-y-2">
          <p className="font-bold text-xs uppercase text-primary tracking-widest">
            Resumo da Configuracao
          </p>
          <div className="text-xs text-primary space-y-1">
            <div>
              <span className="text-secondary">Fallback:</span>{" "}
              <span className="text-primary font-medium">
                {fallbackEnabled ? `${fallbackStrategy}` : "Desabilitado"}
              </span>
            </div>
            <div>
              <span className="text-secondary">Tratamento de Erros:</span>{" "}
              <span className="text-primary font-medium">{errorAction}</span>
            </div>
            <div>
              <span className="text-secondary">Status:</span>{" "}
              <span
                className={`font-semibold ${watch("active") ? "text-green-400" : "text-yellow-400"}`}
              >
                {watch("active") ? "Ativo" : "Inativo"}
              </span>
            </div>
          </div>
        </Card>
      </div>
    </ScrollArea>
  );
}
