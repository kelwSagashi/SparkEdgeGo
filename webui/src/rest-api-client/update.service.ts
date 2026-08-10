import { axios_api_instance } from "@/lib/api-client";
import type { ReturningQueries } from "@/types/db";

export interface UpdateCheckResult {
  enabled: boolean;
  provider: string;
  repository: string;
  current_version: string;
  current_target: string;
  can_compare: boolean;
  update_available: boolean;
  checked_at: string;
  latest_version?: string;
  release_name?: string;
  release_notes?: string;
  release_url?: string;
  published_at?: string;
  compatibility_message?: string;
  compatible_asset?: {
    name: string;
    download_url: string;
    size: number;
  } | null;
}

export const updateService = {
  check: () =>
    axios_api_instance.get<ReturningQueries<UpdateCheckResult>>(`/update/check`),
};
