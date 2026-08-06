import axios from "axios";

export const axios_api_instance = axios.create({
  baseURL: "/api",
  withCredentials: true,
});
