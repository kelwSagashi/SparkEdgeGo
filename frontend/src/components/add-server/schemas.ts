import { z } from 'zod';

const AuthorizationTypes = ['No Auth', 'API Key', 'Bearer Token', 'Basic Auth', 'Digest Auth'];
const ServerEndpointMethods = ["GET", "POST", "PUT", "DELETE", "PATCH"];
export const Methods = [
  { value: ServerEndpointMethods[0], className: 'text-[#04D397] focus:text-[#04D397]' },
  { value: ServerEndpointMethods[1], className: 'text-[#EBD176] focus:text-[#EBD176]' },
  { value: ServerEndpointMethods[2], className: 'text-[#52A1DF] focus:text-[#52A1DF]' },
  { value: ServerEndpointMethods[3], className: 'text-[#C0A3DB] focus:text-[#C0A3DB]' },
  { value: ServerEndpointMethods[4], className: 'text-[#F79185] focus:text-[#F79185]' },
]
const digestAuthSchema = z.object({
  auth_username: z.string().min(1, "Nome de usuário é obrigatório"),
  auth_password: z.string().min(1, "Senha é obrigatória"),
  auth_realm: z.string().optional(),
});

export const DigestAuthSchema = digestAuthSchema;
export type DigestAuthFormValues = z.infer<typeof digestAuthSchema>;

const basicAuthSchema = z.object({
  auth_username: z.string().min(1, "Nome de usuário é obrigatório"),
  auth_password: z.string().min(1, "Senha é obrigatória"),
});
export const BasicAuthSchema = basicAuthSchema;
export type BasicAuthFormValues = z.infer<typeof basicAuthSchema>;

const bearerTokenSchema = z.object({
  auth_token: z.string().min(1, "Bearer Token é obrigatório"),
});

export const BearerTokenSchema = bearerTokenSchema;
export type BearerTokenFormValues = z.infer<typeof bearerTokenSchema>;

// Schema Zod para validação do formulário
const apiKeySchema = z.object({
  auth_token: z.string().min(1, "API Key é obrigatória"),
  auth_api_key_location: z.enum(["header", "query"]),
  auth_header_name: z.string().min(1, "Nome do cabeçalho é obrigatório").optional(),
  auth_query_param_name: z.string().min(1, "Nome do parâmetro de query é obrigatório").optional(),
}).superRefine((data, ctx) => {
  // Lógica condicional de validação:
  if (data.auth_api_key_location === "header" && !data.auth_header_name) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Nome do cabeçalho é obrigatório quando a API Key é enviada no cabeçalho.",
      path: ['auth_header_name'],
    });
  }
  if (data.auth_api_key_location === "query" && !data.auth_query_param_name) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Nome do parâmetro de query é obrigatório quando a API Key é enviada na URL.",
      path: ['auth_query_param_name'],
    });
  }
});

export const ApiKeySchema = apiKeySchema;
export type ApiKeyFormValues = z.infer<typeof apiKeySchema>;

export const ServerHeaderSchema = z.object({
  key: z.string().min(1, "Nome do header é obrigatório"),
  value: z.string().min(1, "Valor do header é obrigatório"),
  type: z.enum(["text", "number"])
});

export const ServerEndpointSchema = z.object({
  path: z.string().min(1, "O caminho é obrigatório")
    .regex(/^\//, "O caminho deve começar com '/'")
    .refine((val) => {
      // Verifica se as chaves estão balanceadas
      const open = (val.match(/{{/g) || []).length;
      const close = (val.match(/}}/g) || []).length;
      return open === close;
    }, { message: "As variáveis '{{ }}' estão desbalanceadas" })
    .refine((val) => !val.match(/{[^{}]+{/) && !val.match(/}[^{}]+}/), { message: "Formato inválido de variável no caminho" })
    .refine((val) => {
      // Não permite ocorrências de {{}} vazias, e valida conteúdo interno de cada placeholder
      const re = /\{\{([\s\S]*?)\}\}/g;
      let m: RegExpExecArray | null;
      while ((m = re.exec(val)) !== null) {
        const inner = m[1].trim();
        // proíbe vazio
        if (inner.length === 0) return false;
        // valida caracteres permitidos no nome da variável (ajuste o regex se quiser outros caracteres)
        if (!/^[a-zA-Z0-9._-]+$/.test(inner)) return false;
      }
      return true;
    }, { message: "Placeholders devem conter um nome válido (ex: {{id}}) — não pode ser '{{}}'." })
    .regex(/^\/([a-zA-Z0-9_\-\/{}.]*)$/, "O caminho contém caracteres inválidos"),
  method: z.enum(ServerEndpointMethods),
  name: z.string().min(1, "O nome é obrigatória"),
  description: z.string().optional(),
  payload_schema: z.object().optional(),
  response_schema: z.object().optional(),
  headers: z.array(
    ServerHeaderSchema
  ).optional(),
});

export const ServerFormSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, "O nome é obrigatório"),
  address: z.string().optional(),
  credential_id: z.string(),
  headers: z.array(
    ServerHeaderSchema
  ).optional(),
});

export const FullServerSchema = z.object({
  server: ServerFormSchema,
  endpoints: z.array(ServerEndpointSchema),
});

export type FullServerValues = z.infer<typeof FullServerSchema>;

// --- NEW DYNAMIC BUILDER SCHEMAS ---

export const OperationSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, "O nome da operação é obrigatório"),
  type: z.string().min(1, "O tipo da operação é obrigatório"),
  config: z.record(z.string(), z.any()),
  input_schema: z.any().optional(),
  output_schema: z.any().optional(),
});

export const ResourceSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1, "O nome do recurso é obrigatório"),
  type: z.string().min(1, "O tipo do recurso é obrigatório"),
  config: z.record(z.string(), z.any()),
  operations: z.array(OperationSchema),
});

export const ServerBuilderSchema = z.object({
  server: z.object({
    name: z.string().min(1, "O nome do servidor é obrigatório"),
    type: z.string().min(1, "O tipo do servidor é obrigatório"),
    server_type_id: z.string().optional(),
    driver_key: z.string().optional(),
    credential_id: z.string().optional(),
  }),
  resources: z.array(ResourceSchema),
});

export type ServerBuilderValues = z.infer<typeof ServerBuilderSchema>;
export type ResourceValues = z.infer<typeof ResourceSchema>;
export type OperationValues = z.infer<typeof OperationSchema>;
