import axios from 'axios';
import type { UserReturningValues } from "@/types/db";

const baseURL = '/api';
const axiosInstance = axios.create({ baseURL, withCredentials: true });

type APIResponse<T> = { data: T; error: string | null };
export type LoginRequest = { email: string; password: string };
export type RegisterRequest = { email: string; password: string; first_name?: string; last_name?: string };

export class AuthAPI {
  async login(payload: LoginRequest) {
    return axiosInstance.post<APIResponse<UserReturningValues>>('/auth/login', payload);
  }

  async register(payload: RegisterRequest) {
    return axiosInstance.post<APIResponse<UserReturningValues>>('/auth/register', payload);
  }

  async me() {
    return axiosInstance.get<APIResponse<UserReturningValues | null>>('/auth/me');
  }

  async logout() {
    return axiosInstance.post('/auth/logout');
  }

  async generateNewApiKey(userId: string) {
    return axiosInstance.post<APIResponse<{
      userId: string;
      apiKey: string;
    }>>(`/auth/generate-new-api-key/${userId}`);
  }
}

export const authApi = new AuthAPI();

