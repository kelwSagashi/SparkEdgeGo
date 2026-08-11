export type ReturningQueries<T> = {
  error?: unknown;
  data: T;
};

export type SchemaConfigIO = {
  key?: string;
  name?: string;
  label?: string;
  type: string;
  required?: boolean;
  description?: string;
  nullable?: boolean;
  default?: unknown;
  fields?: SchemaConfigIO[];
  options?: Array<{ label: string; value: unknown }>;
};

export type SchemaConfig = {
  inputs: SchemaConfigIO[];
  outputs: SchemaConfigIO[];
};

export type ServerTypeReturningValues = {
  id: string;
  key: string;
  name: string;
  description?: string | null;
  [key: string]: unknown;
};

export type ServerTypeUpsertValues = Partial<ServerTypeReturningValues>;

export type AuthorizationsTypeReturningValues = {
  id: string;
  key?: string;
  name: string;
  strategy?: string | null;
  fields?: Array<Record<string, unknown>>;
  server_type_id?: string | null;
  [key: string]: unknown;
};

export type AuthorizationsTypeUpsertValues = Partial<AuthorizationsTypeReturningValues>;

export type CredentialReturningValues = {
  id: string;
  name: string;
  auth_type_id: string;
  data?: Record<string, unknown>;
  owner_id?: string | null;
  project_id?: string | null;
  created_at?: string;
  [key: string]: unknown;
};

export type CredentialUpsertValues = Partial<CredentialReturningValues>;

export type ServerReturningValues = {
  id: string;
  name: string;
  type: string;
  address?: string | null;
  base_url?: string | null;
  server_type_id?: string | null;
  driver_key?: string | null;
  credential_id?: string | null;
  headers?: Record<string, unknown>;
  project_id?: string | null;
  created_by?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type ServerUpsertValues = Partial<ServerReturningValues>;

export type ServerResourceReturningValues = {
  id: string;
  server_id: string;
  name: string;
  type?: string | null;
  config?: Record<string, unknown>;
  created_at?: string;
  [key: string]: unknown;
};

export type ServerResourceUpsertValues = Partial<ServerResourceReturningValues>;

export type ResourceOperationReturningValues = {
  id: string;
  resource_id: string;
  name: string;
  type?: string | null;
  config?: Record<string, unknown>;
  input_schema?: Record<string, unknown>;
  output_schema?: Record<string, unknown>;
  created_at?: string;
  [key: string]: unknown;
};

export type ResourceOperationUpsertValues = Partial<ResourceOperationReturningValues>;

export type DeviceReturningValues = {
  id: string;
  name: string;
  brand?: string | null;
  connection?: string | null;
  connection_method?: string | null;
  serial_number?: string | null;
  description?: string | null;
  location?: string | null;
  ip_address?: string | null;
  resource_operation_id?: string | null;
  others?: Array<Record<string, unknown>>;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type DeviceUpsertValues = Partial<DeviceReturningValues>;

export type UserReturningValues = {
  id: string;
  email: string;
  username?: string | null;
  first_name?: string | null;
  last_name?: string | null;
  role?: string | null;
  is_active?: boolean | null;
  api_key?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type UserUpsertValues = Partial<UserReturningValues> & {
  password?: string;
};

export type ProjectReturningValues = {
  id: string;
  key?: string;
  name: string;
  description?: string | null;
  owner_id?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type ProjectUpsertValues = Partial<ProjectReturningValues>;

export type DownloadedScriptReturningValues = {
  id: string;
  name: string;
  description?: string | null;
  author?: string | null;
  version?: string | null;
  source?: string | null;
  github_repo?: string | null;
  github_ref?: string | null;
  local_path?: string | null;
  main_file?: string | null;
  venv_path?: string | null;
  requirements_file?: string | null;
  venv_ready?: boolean | null;
  language?: string | null;
  tags?: string[];
  schema_config?: { inputs?: SchemaConfigIO[]; outputs?: SchemaConfigIO[] } | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type DownloadedScriptHistoryEntry = {
  id: string;
  script_id: string;
  action: "installed" | "bundle_updated" | "metadata_updated" | string;
  name: string;
  description?: string | null;
  author?: string | null;
  version?: string | null;
  main_file?: string | null;
  requirements_file?: string | null;
  tags?: string[];
  schema_config?: { inputs?: SchemaConfigIO[]; outputs?: SchemaConfigIO[] } | null;
  created_at?: string;
};

export type DownloadedScriptUpsertValues = Partial<DownloadedScriptReturningValues>;

export type InstanceReturningValues = {
  id: string;
  name: string;
  description?: string | null;
  tags?: string[];
  status?: "idle" | "running" | "paused" | "error";
  active?: boolean | null;
  project_id?: string | null;
  device_id?: string | null;
  script_id?: string | null;
  include_device_data?: boolean | null;
  script_parameters?: Record<string, unknown> | Array<Record<string, unknown>>;
  trigger_type?: string | null;
  trigger_config?: Record<string, unknown> | null;
  fallback_enabled?: boolean | null;
  fallback_strategy?: string | null;
  fallback_retry_interval_seconds?: number | null;
  on_error_action?: string | null;
  on_error_config?: Record<string, unknown> | null;
  created_by?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type InstanceUpsertValues = Partial<InstanceReturningValues>;

export type InstanceDetailReturningValues = {
  instance: InstanceReturningValues;
  destinations: InstanceDestinationReturningValues[];
};

export type InstanceExecutionReturningValues = {
  id: string;
  instance_id: string;
  status?: string | null;
  trigger_type?: string | null;
  output?: string | null;
  error_message?: string | null;
  destination_sent?: boolean | null;
  fallback_used?: boolean | null;
  duration_ms?: number | null;
  logs?: Array<Record<string, unknown>>;
  started_at?: string | null;
  finished_at?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type InstanceDestinationReturningValues = {
  id: string;
  instance_id?: string | null;
  resource_operation_id?: string | null;
  enabled?: boolean | null;
  priority?: number | null;
  retry_policy?: Record<string, unknown> | null;
  created_at?: string;
  mapping?: DataMappingReturningValues | null;
  data_mapping?: DataMappingReturningValues | null;
  dataMapping?: DataMappingReturningValues | null;
  [key: string]: unknown;
};

export type InstanceDestinationUpsertValues = Partial<InstanceDestinationReturningValues>;

export type DataMappingReturningValues = {
  id: string;
  instance_destination_id?: string | null;
  mapping?: Record<string, unknown>;
  payload_template?: Record<string, unknown>;
  custom_fields?: Array<{ key: string; value: string }>;
  transform_script?: string | null;
  created_at?: string;
  [key: string]: unknown;
};

export type DataMappingUpsertValues = Partial<DataMappingReturningValues>;

export type LocalFallbackItemReturningValues = {
  id: string;
  instance_id?: string | null;
  destination_id?: string | null;
  execution_id?: string | null;
  status?: string | null;
  payload?: string | null;
  filepath?: string | null;
  retry_count?: number | null;
  last_retry_at?: string | null;
  next_retry_at?: string | null;
  last_error?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};

export type LocalFallbackItemUpsertValues = Partial<LocalFallbackItemReturningValues>;

export type TagReturningValues = {
  id: string;
  name: string;
  color?: string | null;
  project_id?: string | null;
  created_at?: string;
  updated_at?: string;
  [key: string]: unknown;
};
