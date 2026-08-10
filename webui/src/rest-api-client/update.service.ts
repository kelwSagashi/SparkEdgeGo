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
  integrity_ready: boolean;
  expected_sha256?: string;
  compatible_asset?: {
    name: string;
    download_url: string;
    size: number;
  } | null;
}

export interface UpdateDownloadResult {
  version: string;
  target: string;
  asset_name: string;
  release_url: string;
  downloaded_path: string;
  size: number;
  sha256: string;
  checksum_verified: boolean;
}

export interface UpdateApplyResult {
  version: string;
  target: string;
  downloaded_path: string;
  staging_path: string;
  backup_path: string;
  script_path?: string;
  applied: boolean;
  prepared_only: boolean;
  restart_required: boolean;
  message: string;
  applied_files: string[];
  preserved_files: string[];
  next_steps: string[];
}

export interface UpdateState {
  last_downloaded_package?: string;
  last_prepared_version?: string;
  last_prepared_target?: string;
  last_apply_result?: UpdateApplyResult | null;
  last_download_result?: UpdateDownloadResult | null;
  updated_at?: string;
}

export const updateService = {
  check: () =>
    axios_api_instance.get<ReturningQueries<UpdateCheckResult>>(`/update/check`),
  status: () =>
    axios_api_instance.get<ReturningQueries<UpdateState>>(`/update/status`),
  download: () =>
    axios_api_instance.post<ReturningQueries<UpdateDownloadResult>>(`/update/download`),
  apply: (downloadedPath: string) =>
    axios_api_instance.post<ReturningQueries<UpdateApplyResult>>(`/update/apply`, {
      downloaded_path: downloadedPath,
    }),
};
