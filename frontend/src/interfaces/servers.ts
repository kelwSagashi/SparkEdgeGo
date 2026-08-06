import type { 
  ServerReturningValues, 
  ServerTypeReturningValues,
  ProjectReturningValues,
  CredentialReturningValues 
} from "@/types/db";

/**
 * Extended Server interface with related data for UI display
 */
export interface ServerWithRelations extends ServerReturningValues {
  base_url?: string;
  serverType?: ServerTypeReturningValues;
  project?: ProjectReturningValues;
  credential?: CredentialReturningValues;
}

/**
 * Server form data for creating/editing servers
 */
export interface ServerFormData {
  name: string;
  type: string;
  address: string;
  credential_id?: string;
  headers?: Record<string, string>;
  project_id: string;
}

/**
 * Header key-value pair for dynamic header input
 */
export interface HeaderKeyValue {
  key: string;
  value: string;
}

