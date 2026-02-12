import axios from "axios";

// Access environment variable (injected by react-scripts at build time)
const getApiUrl = (): string => {
  return process.env.REACT_APP_API_URL || '/api';
};

const api = axios.create({
  baseURL: getApiUrl(),
});

// Add request interceptor to set Content-Type only for non-FormData requests
api.interceptors.request.use((config: any) => {
  // Don't set Content-Type for FormData - let axios set it automatically with boundary
  if (!(config.data instanceof FormData)) {
    config.headers['Content-Type'] = 'application/json';
  }
  return config;
});

export default api;

export async function setHostReviewed(id: string, reviewed: boolean) {
  const { data } = await api.patch(`/hosts/${id}/reviewed`, { reviewed });
  return data;
}

export async function setPortReviewed(id: string, reviewed: boolean) {
  const { data } = await api.patch(`/ports/${id}/reviewed`, { reviewed });
  return data;
}

export async function setURLReviewed(id: string, reviewed: boolean) {
  const { data } = await api.patch(`/urls/${id}/reviewed`, { reviewed });
  return data;
}
