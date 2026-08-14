import {
  AlertCircle,
  ArrowDown,
  ArrowRight,
  CheckCircle2,
  Clock3,
  GitBranch,
  RadioTower,
  Workflow,
  Zap,
} from "lucide-react";
import { Controller, useFormContext } from "react-hook-form";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
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
import { cn } from "@/lib/utils";
import type { InstanceFormValues } from "./instance-form.schemas";

type TriggerFormProps = {
  instances?: Array<{ id: string; name?: string | null }>;
};

const triggerCopy = {
  interval: {
    label: "Intervalo agendado",
    description: "Executa automaticamente em intervalos regulares e previsiveis.",
    accent: "bg-sky-500/10 text-sky-200 border-sky-500/30",
    icon: Clock3,
  },
  webhook: {
    label: "Webhook",
    description: "Executa quando um sistema externo chamar o endpoint configurado.",
    accent: "bg-violet-500/10 text-violet-200 border-violet-500/30",
    icon: Zap,
  },
  interval_and_webhook: {
    label: "Intervalo + webhook",
    description: "Combina agenda recorrente com acionamento remoto sob demanda.",
    accent: "bg-fuchsia-500/10 text-fuchsia-200 border-fuchsia-500/30",
    icon: Workflow,
  },
  event: {
    label: "Evento interno",
    description: "Dispara quando outro fluxo interno publicar o evento configurado.",
    accent: "bg-emerald-500/10 text-emerald-200 border-emerald-500/30",
    icon: GitBranch,
  },
  mqtt: {
    label: "Mensagem MQTT",
    description: "Escuta um topico MQTT e executa quando novas mensagens chegarem.",
    accent: "bg-amber-500/10 text-amber-200 border-amber-500/30",
    icon: RadioTower,
  },
  state_change: {
    label: "Mudanca de estado",
    description: "Executa quando um campo monitorado atingir o valor esperado.",
    accent: "bg-rose-500/10 text-rose-200 border-rose-500/30",
    icon: CheckCircle2,
  },
} as const;

const orchestrationHints = {
  sequential:
    "Bom para pipelines com dependencia forte, ordem deterministica ou compartilhamento de contexto.",
  parallel:
    "Bom para throughput maior quando as etapas podem rodar sem depender uma da outra.",
} as const;

export function InstanceTriggerForm({ instances = [] }: TriggerFormProps) {
  const {
    control,
    register,
    watch,
    setValue,
    formState: { errors },
  } = useFormContext<InstanceFormValues>();

  const instanceName = watch("name");
  const triggerType = watch("triggerType");
  const executionMode = watch("executionMode");
  const workflowEnabled = watch("orchestrationConfig.workflow_enabled");
  const allowPartialSuccess = watch("orchestrationConfig.allow_partial_success");
  const debounceSeconds = watch("orchestrationConfig.debounce_seconds");
  const triggerConfig = watch("triggerConfig");
  const dependsOn = (watch("dependsOn") || [])
    .map((item) => item.trim())
    .filter(Boolean);

  const selectedDependencies = dependsOn
    .map((id) => instances.find((instance) => instance.id === id) || { id, name: null })
    .filter(
      (item, index, list) => list.findIndex((candidate) => candidate.id === item.id) === index,
    );
  const unresolvedDependencies = selectedDependencies.filter(
    (instance) => !instances.some((item) => item.id === instance.id),
  );
  const currentTrigger = triggerCopy[triggerType];
  const TriggerIcon = currentTrigger.icon;
  const currentInstanceLabel = instanceName?.trim() || "Instancia atual";

  const dependencySummary =
    selectedDependencies.length === 0
      ? "Nenhuma dependencia selecionada. Esta instancia pode iniciar de forma independente."
      : selectedDependencies.length === 1
        ? "1 dependencia selecionada. A execucao tende a se comportar como uma sequencia simples."
        : `${selectedDependencies.length} dependencias selecionadas. A execucao vira um pequeno fluxo com convergencia.`;

  const triggerSummary =
    triggerType === "interval" || triggerType === "interval_and_webhook"
      ? `Intervalo: ${triggerConfig?.interval_seconds ?? "-"}s`
      : triggerType === "webhook" || triggerType === "interval_and_webhook"
        ? `Endpoint: ${triggerConfig?.webhook_path || "/webhook/..."}`
        : triggerType === "event"
          ? `Evento: ${triggerConfig?.event_name || "nao definido"}`
          : triggerType === "mqtt"
            ? `Topico: ${triggerConfig?.mqtt_topic || "nao definido"}`
            : `Campo: ${triggerConfig?.state_field || "nao definido"}`;

  const toggleDependency = (instanceId: string) => {
    if (dependsOn.includes(instanceId)) {
      setValue(
        "dependsOn",
        dependsOn.filter((item) => item !== instanceId),
        { shouldDirty: true, shouldValidate: true },
      );
      return;
    }

    setValue("dependsOn", [...dependsOn, instanceId], {
      shouldDirty: true,
      shouldValidate: true,
    });
  };

  return (
    <ScrollArea className="h-full">
      <div className="space-y-6 pr-4">
        <div className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
          <Card className="border-border bg-accent/40 p-5">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge className={cn("border", currentTrigger.accent)}>
                    <TriggerIcon className="mr-1 h-3.5 w-3.5" />
                    {currentTrigger.label}
                  </Badge>
                  <Badge variant="outline" className="border-border text-secondary">
                    {executionMode === "parallel" ? "Paralelo" : "Sequencial"}
                  </Badge>
                  {workflowEnabled ? (
                    <Badge
                      variant="outline"
                      className="border-emerald-500/40 text-emerald-300"
                    >
                      Workflow visual ativo
                    </Badge>
                  ) : null}
                </div>
                <div>
                  <h3 className="text-base font-semibold text-primary">
                    Como esta instancia entra no fluxo
                  </h3>
                  <p className="mt-1 text-sm text-secondary">{currentTrigger.description}</p>
                </div>
                <div className="grid gap-2 sm:grid-cols-2">
                  <div className="rounded-xl border border-border bg-background/50 p-3">
                    <p className="text-xs uppercase tracking-[0.24em] text-secondary/80">
                      Dependencias
                    </p>
                    <p className="mt-1 text-sm text-primary">{dependencySummary}</p>
                  </div>
                  <div className="rounded-xl border border-border bg-background/50 p-3">
                    <p className="text-xs uppercase tracking-[0.24em] text-secondary/80">
                      Politica de execucao
                    </p>
                    <p className="mt-1 text-sm text-primary">
                      {orchestrationHints[executionMode]}
                    </p>
                  </div>
                </div>
              </div>

              <div className="min-w-[260px] flex-1 rounded-2xl border border-border bg-background/65 p-4">
                <div className="mb-3 flex items-center gap-2">
                  <Workflow className="h-4 w-4 text-primary" />
                  <p className="text-sm font-semibold text-primary">Mini workflow</p>
                </div>
                <div className="space-y-3">
                  <div className="rounded-xl border border-border bg-muted/30 p-3">
                    <p className="text-xs uppercase tracking-[0.24em] text-secondary/80">
                      Trigger
                    </p>
                    <p className="mt-1 text-sm font-medium text-primary">
                      {currentTrigger.label}
                    </p>
                    <p className="mt-1 text-xs text-secondary">{triggerSummary}</p>
                  </div>

                  <div className="flex justify-center text-secondary/70">
                    <ArrowDown className="h-4 w-4" />
                  </div>

                  {selectedDependencies.length > 0 ? (
                    <div className="grid gap-2">
                      {selectedDependencies.map((dependency) => {
                        const unresolved = unresolvedDependencies.some(
                          (item) => item.id === dependency.id,
                        );
                        return (
                          <div
                            key={dependency.id}
                            className={cn(
                              "rounded-xl border p-3",
                              unresolved
                                ? "border-amber-500/40 bg-amber-500/10"
                                : "border-border bg-muted/30",
                            )}
                          >
                            <p className="text-sm font-medium text-primary">
                              {dependency.name || dependency.id}
                            </p>
                            <p className="mt-1 text-xs text-secondary">
                              {unresolved
                                ? "Dependencia mantida por ID, mas nao encontrada na lista atual."
                                : `ID: ${dependency.id}`}
                            </p>
                          </div>
                        );
                      })}
                      <div className="flex justify-center text-secondary/70">
                        <ArrowDown className="h-4 w-4" />
                      </div>
                    </div>
                  ) : null}

                  <div className="rounded-xl border border-primary/30 bg-primary/10 p-3">
                    <p className="text-sm font-semibold text-primary">{currentInstanceLabel}</p>
                    <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-secondary">
                      <span className="rounded-full border border-border px-2 py-0.5">
                        {executionMode === "parallel" ? "fan-out paralelo" : "ordem sequencial"}
                      </span>
                      <span className="rounded-full border border-border px-2 py-0.5">
                        debounce {debounceSeconds ?? 0}s
                      </span>
                      <span className="rounded-full border border-border px-2 py-0.5">
                        {allowPartialSuccess === false
                          ? "falha bloqueante"
                          : "sucesso parcial permitido"}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Card>

          <div className="space-y-3">
            {unresolvedDependencies.length > 0 ? (
              <Alert className="border-amber-500/30 bg-amber-500/10 text-amber-100">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Dependencias fora da lista atual</AlertTitle>
                <AlertDescription>
                  {unresolvedDependencies.map((item) => item.name || item.id).join(", ")} ainda
                  estao registradas por ID. Isso e util em migracoes, mas vale revisar se a cadeia
                  de execucao continua correta.
                </AlertDescription>
              </Alert>
            ) : null}

            {executionMode === "parallel" && selectedDependencies.length > 1 ? (
              <Alert className="border-sky-500/30 bg-sky-500/10 text-sky-100">
                <Workflow className="h-4 w-4" />
                <AlertTitle>Fluxo com convergencia</AlertTitle>
                <AlertDescription>
                  Esta instancia depende de varias etapas e depois segue em modo paralelo.
                  Verifique se o script aceita receber resultados sincronizados de mais de uma
                  origem.
                </AlertDescription>
              </Alert>
            ) : null}

            {triggerType === "interval_and_webhook" ? (
              <Alert className="border-fuchsia-500/30 bg-fuchsia-500/10 text-fuchsia-100">
                <Zap className="h-4 w-4" />
                <AlertTitle>Trigger combinado</AlertTitle>
                <AlertDescription>
                  Esta instancia pode executar tanto por agenda quanto por chamada remota. Isso e
                  otimo para reprocessamento sob demanda, mas aumenta a chance de sobreposicao.
                </AlertDescription>
              </Alert>
            ) : null}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="triggerType" className="font-medium text-primary">
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
                      <SelectItem value="interval_and_webhook">Intervalo + Webhook</SelectItem>
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
            <Label htmlFor="executionMode" className="font-medium text-primary">
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

        <Card className="space-y-2 border-border bg-accent p-4">
          <p className="text-sm text-secondary">{currentTrigger.description}</p>
          <p className="text-xs text-secondary/80">
            Modo atual:{" "}
            <strong className="text-primary">
              {executionMode === "parallel" ? "Paralelo" : "Sequencial"}
            </strong>
          </p>
        </Card>

        {(triggerType === "interval" || triggerType === "interval_and_webhook") && (
          <div className="space-y-2">
            <Label htmlFor="interval_seconds" className="font-medium text-primary">
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
            <p className="text-xs text-secondary">
              Para coleta recorrente, prefira janelas que respeitem o tempo medio do script e do
              envio.
            </p>
          </div>
        )}

        {(triggerType === "webhook" || triggerType === "interval_and_webhook") && (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="webhook_path" className="font-medium text-primary">
                Caminho do Webhook
              </Label>
              <Input
                id="webhook_path"
                placeholder="/webhook/minha-instancia"
                {...register("triggerConfig.webhook_path")}
                className="text-primary"
              />
              {errors.triggerConfig?.webhook_path && (
                <p className="text-sm text-destructive">
                  {errors.triggerConfig.webhook_path.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="webhook_secret" className="font-medium text-primary">
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
            <div className="rounded-xl border border-border bg-muted/30 p-3 text-xs text-secondary md:col-span-2">
              Use um caminho estavel e um segredo quando o endpoint puder ser chamado fora da rede
              controlada.
            </div>
          </div>
        )}

        {triggerType === "event" && (
          <div className="space-y-2">
            <Label htmlFor="event_name" className="font-medium text-primary">
              Nome do Evento
            </Label>
            <Input
              id="event_name"
              placeholder="sensor.updated"
              {...register("triggerConfig.event_name")}
              className="text-primary"
            />
            {errors.triggerConfig?.event_name && (
              <p className="text-sm text-destructive">
                {errors.triggerConfig.event_name.message}
              </p>
            )}
            <p className="text-xs text-secondary">
              Use nomes semanticamente claros, como{" "}
              <code className="text-primary">device.synced</code> ou{" "}
              <code className="text-primary">telemetry.enriched</code>.
            </p>
          </div>
        )}

        {triggerType === "mqtt" && (
          <div className="space-y-2">
            <Label htmlFor="mqtt_topic" className="font-medium text-primary">
              Topico MQTT
            </Label>
            <Input
              id="mqtt_topic"
              placeholder="sparkedge/devices/+/events"
              {...register("triggerConfig.mqtt_topic")}
              className="text-primary"
            />
            {errors.triggerConfig?.mqtt_topic && (
              <p className="text-sm text-destructive">
                {errors.triggerConfig.mqtt_topic.message}
              </p>
            )}
            <p className="text-xs text-secondary">
              Aceita curingas do MQTT. Revise o escopo para evitar disparos excessivos.
            </p>
          </div>
        )}

        {triggerType === "state_change" && (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="state_field" className="font-medium text-primary">
                Campo Monitorado
              </Label>
              <Input
                id="state_field"
                placeholder="$.device.status"
                {...register("triggerConfig.state_field")}
                className="text-primary"
              />
              {errors.triggerConfig?.state_field && (
                <p className="text-sm text-destructive">
                  {errors.triggerConfig.state_field.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="state_equals" className="font-medium text-primary">
                Valor Esperado
              </Label>
              <Input
                id="state_equals"
                placeholder="online"
                {...register("triggerConfig.state_equals")}
                className="text-primary"
              />
              {errors.triggerConfig?.state_equals && (
                <p className="text-sm text-destructive">
                  {errors.triggerConfig.state_equals.message}
                </p>
              )}
            </div>
          </div>
        )}

        <div className="space-y-3 border-t border-border pt-2">
          <Label className="text-base font-semibold text-primary">
            Dependencias e Workflow
          </Label>
          <div className="space-y-2">
            <Label htmlFor="dependsOnCsv" className="font-medium text-primary">
              Dependencias entre Instancias
            </Label>
            <Textarea
              id="dependsOnCsv"
              value={dependsOn.join(", ")}
              onChange={(event) => {
                const next = Array.from(
                  new Set(
                    event.target.value
                      .split(",")
                      .map((item) => item.trim())
                      .filter(Boolean),
                  ),
                );
                setValue("dependsOn", next, { shouldDirty: true, shouldValidate: true });
              }}
              placeholder="id-instancia-a, id-instancia-b"
              className="min-h-20 text-primary"
            />
            <p className="text-xs text-secondary">
              Use IDs de instancias para definir precedencia de execucao. As dependencias ficam
              registradas na configuracao da instancia.
            </p>
            {errors.dependsOn && (
              <p className="text-sm text-destructive">{errors.dependsOn.message as string}</p>
            )}
          </div>

          {instances.length > 0 && (
            <Card className="border-border bg-muted/30 p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="text-xs font-semibold text-primary">
                    Instancias disponiveis para dependencia
                  </p>
                  <p className="text-xs text-secondary">
                    Clique para montar a ordem do fluxo sem precisar digitar IDs manualmente.
                  </p>
                </div>
                <Badge variant="outline" className="border-border text-secondary">
                  {selectedDependencies.length} selecionada(s)
                </Badge>
              </div>
              <div className="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
                {instances.map((instance) => {
                  const selected = dependsOn.includes(instance.id);
                  return (
                    <button
                      key={instance.id}
                      type="button"
                      className={cn(
                        "rounded-xl border p-3 text-left transition-colors",
                        selected
                          ? "border-primary/40 bg-primary/10 text-primary"
                          : "border-border text-secondary hover:border-primary/30 hover:text-primary",
                      )}
                      onClick={() => toggleDependency(instance.id)}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <p className="text-sm font-medium">
                            {instance.name || "Instancia sem nome"}
                          </p>
                          <p className="mt-1 break-all text-[11px] opacity-80">{instance.id}</p>
                        </div>
                        {selected ? <CheckCircle2 className="h-4 w-4 shrink-0" /> : null}
                      </div>
                    </button>
                  );
                })}
              </div>
            </Card>
          )}

          {selectedDependencies.length > 0 ? (
            <Card className="border-border bg-background/50 p-4">
              <p className="text-sm font-semibold text-primary">Cadeia atual de dependencia</p>
              <p className="text-xs text-secondary">
                Uma leitura rapida da ordem esperada ate a instancia atual.
              </p>
              <div className="mt-3 flex flex-col gap-2 md:flex-row md:flex-wrap md:items-center">
                {selectedDependencies.map((dependency) => (
                  <div key={dependency.id} className="flex items-center gap-2">
                    <div className="rounded-xl border border-border bg-muted/30 px-3 py-2">
                      <p className="text-sm font-medium text-primary">
                        {dependency.name || dependency.id}
                      </p>
                    </div>
                    <ArrowRight className="hidden h-4 w-4 text-secondary/70 md:block" />
                  </div>
                ))}
                <div className="rounded-xl border border-primary/30 bg-primary/10 px-3 py-2">
                  <p className="text-sm font-semibold text-primary">{currentInstanceLabel}</p>
                </div>
              </div>
            </Card>
          ) : null}

          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
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
              <Label htmlFor="debounce_seconds" className="font-medium text-primary">
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
              {errors.orchestrationConfig?.debounce_seconds && (
                <p className="text-sm text-destructive">
                  {errors.orchestrationConfig.debounce_seconds.message}
                </p>
              )}
            </div>
          </div>

          <div className="grid gap-3 lg:grid-cols-3">
            <div className="rounded-xl border border-border bg-muted/25 p-3">
              <p className="text-xs uppercase tracking-[0.24em] text-secondary/80">
                Workflow visual
              </p>
              <p className="mt-1 text-sm text-primary">
                {workflowEnabled
                  ? "Sinaliza que a instancia faz parte de um fluxo orquestrado."
                  : "Desativado. A instancia continua funcional, mas sem destaque de workflow."}
              </p>
            </div>
            <div className="rounded-xl border border-border bg-muted/25 p-3">
              <p className="text-xs uppercase tracking-[0.24em] text-secondary/80">
                Sucesso parcial
              </p>
              <p className="mt-1 text-sm text-primary">
                {allowPartialSuccess === false
                  ? "Uma falha interrompe a percepcao de sucesso do fluxo."
                  : "Falhas isoladas podem coexistir com progresso parcial."}
              </p>
            </div>
            <div className="rounded-xl border border-border bg-muted/25 p-3">
              <p className="text-xs uppercase tracking-[0.24em] text-secondary/80">
                Debounce
              </p>
              <p className="mt-1 text-sm text-primary">
                {(debounceSeconds ?? 0) > 0
                  ? `O fluxo ignora disparos repetidos por ${debounceSeconds}s.`
                  : "Sem debounce. Cada disparo valido pode gerar uma nova execucao."}
              </p>
            </div>
          </div>
        </div>

        <div className="space-y-4 border-t border-border pt-4">
          <Label className="text-base font-semibold text-primary">
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
