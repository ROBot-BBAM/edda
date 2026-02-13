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

export interface Finding {
  id: string;
  host_id?: string;
  host_display?: string;
  port_id?: string;
  port_display?: string;
  url_id?: string;
  url_display?: string;
  title: string;
  severity: string;
  description?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface CreateFindingParams {
  host_id?: string;
  port_id?: string;
  url_id?: string;
  title: string;
  severity?: string;
  description?: string;
  status?: string;
}

export interface FindingsSummary {
  by_severity: Record<string, number>;
  open_count: number;
}

export async function getFindingsSummary() {
  const { data } = await api.get<FindingsSummary>('/findings/summary');
  return data;
}

export async function listFindings(params?: {
  host_id?: string;
  port_id?: string | string[];
  url_id?: string;
  severity?: string;
  status?: string;
}) {
  const search = new URLSearchParams();
  if (params?.host_id) search.set('host_id', params.host_id);
  if (params?.url_id) search.set('url_id', params.url_id);
  if (params?.severity) search.set('severity', params.severity);
  if (params?.status) search.set('status', params.status);
  if (params?.port_id) {
    const ids = Array.isArray(params.port_id) ? params.port_id : [params.port_id];
    ids.forEach((id) => search.append('port_id', id));
  }
  const q = search.toString();
  const url = q ? `/findings?${q}` : '/findings';
  const { data } = await api.get<Finding[]>(url);
  return data;
}

export async function createFinding(body: CreateFindingParams) {
  const { data } = await api.post<Finding>('/findings', body);
  return data;
}

export async function getFinding(id: string) {
  const { data } = await api.get<Finding>(`/findings/${id}`);
  return data;
}

export async function updateFinding(id: string, body: Partial<CreateFindingParams>) {
  const { data } = await api.patch<Finding>(`/findings/${id}`, body);
  return data;
}

export async function deleteFinding(id: string) {
  await api.delete(`/findings/${id}`);
}

// --- Search ---
export interface SearchResult {
  hosts: { id: string; label: string }[];
  urls: { id: string; label: string }[];
  findings: { id: string; title: string; severity: string }[];
}

export async function search(query: string, limit?: number) {
  const params = new URLSearchParams({ q: query });
  if (limit != null) params.set('limit', String(limit));
  const { data } = await api.get<SearchResult>(`/search?${params.toString()}`);
  return data;
}

// --- Global narrative export (for AI / attack narrative) ---
export interface NarrativeExport {
  exported_at: string;
  summary: { hosts: number; ports: number; urls: number; findings: number; notes: number };
  hosts: Array<{ id: string; ip_address: string; hostname?: string; os?: string; reviewed: boolean; created_at: string }>;
  ports: Array<{ id: string; host_id: string; host_ip: string; port: number; protocol: string; state?: string; service_name?: string; service_product?: string; service_version?: string; reviewed: boolean; created_at: string }>;
  urls: Array<{ id: string; host_id?: string; host_display: string; url: string; path: string; method: string; status_code?: number; content_length?: number; words?: number; lines?: number; reviewed: boolean; created_at: string }>;
  findings: Array<{ id: string; title: string; severity: string; status: string; description: string; host_display: string; port_display: string; url_display: string; created_at: string; updated_at: string }>;
  notes: Array<{ id: string; content: string; created_at: string; target_type: string; target_id: string; target_label: string }>;
}

/** Download full engagement export. format: 'json' | 'csv' (csv returns a zip of CSVs). */
export async function downloadNarrativeExport(format: 'json' | 'csv'): Promise<void> {
  const response = await api.get(`/export/narrative?format=${format}`, { responseType: 'blob' });
  const blob = response.data as Blob;
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  const date = new Date().toISOString().slice(0, 10);
  link.download = format === 'csv' ? `edda-export-${date}.zip` : `edda-narrative-export-${date}.json`;
  link.click();
  URL.revokeObjectURL(url);
}

// --- Notes ---
export interface Note {
  id: string;
  content: string;
  created_at: string;
}

export async function createNote(params: { host_id?: string; port_id?: string; url_id?: string; content: string }) {
  const { data } = await api.post<Note>('/notes', params);
  return data;
}

export async function listNotesByHost(hostId: string) {
  const { data } = await api.get<Note[]>(`/hosts/${hostId}/notes`);
  return data;
}

export async function listNotesByPort(portId: string) {
  const { data } = await api.get<Note[]>(`/ports/${portId}/notes`);
  return data;
}

export async function listNotesByURL(urlId: string) {
  const { data } = await api.get<Note[]>(`/urls/${urlId}/notes`);
  return data;
}
