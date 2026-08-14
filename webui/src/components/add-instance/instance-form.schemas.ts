import { z } from "zod";

/**
 * Schema para validação completa do formulário de instância
 */

export const InstanceBasicSchema = z.object({
  name: z
    .string()
    .min(1, "Nome é obrigatório")
    .min(3, "Nome deve ter pelo menos 3 caracteres"),
  description: z.string().optional(),
  project_id: z.string().min(1, "Projeto é obrigatório"),
  device_id: z.string().optional().nullable(),
  tags: z.array(z.string()),
  includeDeviceData: z.boolean(),
});

export const InstanceScriptSchema = z.object({
  script_id: z.string().min(1, "O script é obrigatório"),
  scriptParameters: z.array(
    z.object({
      key: z.string(),
      value: z.any(),
      sourceType: z.enum(["device_data", "manual", "device_field"]),
      deviceFieldKey: z.string().optional(),
    }),
  ).optional(),
  scriptInputs: z.record(z.any(), z.any()).optional(),
});

export const TriggerConfigSchema = z.object({
  interval_seconds: z.number().optional().nullable(),
  webhook_path: z.string().optional().nullable(),
  webhook_secret: z.string().optional().nullable(),
  event_name: z.string().optional().nullable(),
  mqtt_topic: z.string().optional().nullable(),
  state_field: z.string().optional().nullable(),
  state_equals: z.string().optional().nullable(),
  save_execution_on_server: z.boolean(),
});

export const OrchestrationConfigSchema = z.object({
  workflow_enabled: z.boolean().optional(),
  allow_partial_success: z.boolean().optional(),
  debounce_seconds: z.number().optional().nullable(),
});

export const InstanceTriggerSchema = z.object({
  triggerType: z.enum([
    "interval",
    "webhook",
    "interval_and_webhook",
    "event",
    "mqtt",
    "state_change",
  ]),
  triggerConfig: TriggerConfigSchema,
  dependsOn: z.array(z.string()).optional(),
  executionMode: z.enum(["sequential", "parallel"]),
  orchestrationConfig: OrchestrationConfigSchema,
}).superRefine((value, ctx) => {
  const intervalTypes = new Set(["interval", "interval_and_webhook"]);
  const webhookTypes = new Set(["webhook", "interval_and_webhook"]);
  const dependsOn = Array.from(
    new Set((value.dependsOn || []).map((item) => item.trim()).filter(Boolean)),
  );

  if (intervalTypes.has(value.triggerType)) {
    const seconds = value.triggerConfig.interval_seconds;
    if (seconds == null || Number.isNaN(seconds)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["triggerConfig", "interval_seconds"],
        message: "Informe o intervalo em segundos.",
      });
    } else if (seconds < 10) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["triggerConfig", "interval_seconds"],
        message: "Use pelo menos 10 segundos para evitar execucoes excessivas.",
      });
    }
  }

  if (webhookTypes.has(value.triggerType)) {
    const path = value.triggerConfig.webhook_path?.trim();
    if (!path) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["triggerConfig", "webhook_path"],
        message: "Informe o caminho do webhook.",
      });
    } else if (!path.startsWith("/")) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["triggerConfig", "webhook_path"],
        message: "O caminho do webhook deve comecar com '/'.",
      });
    }
  }

  if (value.triggerType === "event" && !value.triggerConfig.event_name?.trim()) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["triggerConfig", "event_name"],
      message: "Informe o nome do evento interno.",
    });
  }

  if (value.triggerType === "mqtt" && !value.triggerConfig.mqtt_topic?.trim()) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["triggerConfig", "mqtt_topic"],
      message: "Informe o topico MQTT.",
    });
  }

  if (value.triggerType === "state_change") {
    if (!value.triggerConfig.state_field?.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["triggerConfig", "state_field"],
        message: "Informe o campo monitorado.",
      });
    }
    if (!value.triggerConfig.state_equals?.trim()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["triggerConfig", "state_equals"],
        message: "Informe o valor esperado para o disparo.",
      });
    }
  }

  if ((value.orchestrationConfig.debounce_seconds ?? 0) < 0) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["orchestrationConfig", "debounce_seconds"],
      message: "O debounce nao pode ser negativo.",
    });
  }

  if (dependsOn.length !== (value.dependsOn || []).length) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["dependsOn"],
      message: "Revise as dependencias: remova itens vazios ou duplicados.",
    });
  }
});

export const RetryPolicySchema = z.object({
  maxRetries: z.number().optional().nullable(),
  retryInterval: z.number().optional().nullable(),
  timeoutSeconds: z.number().optional().nullable(),
  continueOnError: z.boolean().optional(),
  isolationMode: z.enum(["isolate", "continue"]).optional(),
  circuitBreakerThreshold: z.number().optional().nullable(),
  circuitBreakerCooldownSeconds: z.number().optional().nullable(),
});

export const DataMappingSchema = z.object({
  instanceDestinationId: z.string(),
  mapping: z.record(z.string(), z.string()).optional(),
  payloadTemplate: z.any().optional(),
  customFields: z
    .array(
      z.object({
        key: z.string(),
        value: z.string(),
      }),
    )
    .optional(),
  transformScript: z.string().optional().nullable(),
  availableFields: z
    .array(
      z.object({
        label: z.string(),
        key: z.string(),
        type: z.enum(["string", "number", "boolean", "json"]),
        source: z.enum(["script_output", "device_data", "custom"]),
        sourceKey: z.string().optional(),
      }),
    )
    .optional(),
  requiredFields: z.array(z.string()).optional(),
});

export const InstanceDestinationSchema = z.object({
  resourceOperationId: z.string().min(1, "Operação é obrigatória"),
  serverId: z.string().optional().nullable(),
  enabled: z.boolean(),
  priority: z.number(),
  retryPolicy: RetryPolicySchema.optional(),
  dataMapping: DataMappingSchema,
});

export const InstanceDestinationsSchema = z.object({
  destinations: z
    .array(InstanceDestinationSchema)
    .min(1, "Pelo menos um destino é obrigatório"),
});

export const FallbackConfigSchema = z.object({
  enabled: z.boolean(),
  strategy: z.enum(["background_job", "active_queue"]),
  retry_interval_seconds: z.number(),
  max_retries: z.number().optional().nullable(),
});

export const ErrorConfigSchema = z.object({
  action: z.enum(["log_only", "retry", "notify_webhook", "stop"]),
  notify_url: z.string().url().optional().nullable(),
  max_retries: z.number().optional().nullable(),
});

export const InstanceFallbackSchema = z.object({
  fallbackConfig: FallbackConfigSchema,
  errorConfig: ErrorConfigSchema,
  active: z.boolean(),
});

/**
 * Schema completo da instância
 */
export const InstanceFormSchema = z.object({
  ...InstanceBasicSchema.shape,
  ...InstanceScriptSchema.shape,
  ...InstanceTriggerSchema.shape,
  ...InstanceDestinationsSchema.shape,
  ...InstanceFallbackSchema.shape,
});

export type InstanceFormValues = z.infer<typeof InstanceFormSchema>;
export type InstanceBasicValues = z.infer<typeof InstanceBasicSchema>;
export type InstanceScriptValues = z.infer<typeof InstanceScriptSchema>;
export type InstanceTriggerValues = z.infer<typeof InstanceTriggerSchema>;
export type InstanceDestinationsValues = z.infer<
  typeof InstanceDestinationsSchema
>;
export type InstanceFallbackValues = z.infer<typeof InstanceFallbackSchema>;
export type InstanceDestinationValue = z.infer<
  typeof InstanceDestinationSchema
>;
export type DataMappingValue = z.infer<typeof DataMappingSchema>;
export type ErrorConfigValue = z.infer<typeof ErrorConfigSchema>;
export type FallbackConfigValue = z.infer<typeof FallbackConfigSchema>;

