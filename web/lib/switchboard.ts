import { apiFetch } from "./api-fetch";

export type BackendState = {
  id: string;
  url: string;
  healthy: boolean;
  lastCheck: string;
  lastError?: string;
  metrics: { activeRequests: number; requests: number; errors: number; errorRate: number; avgLatencyMs: number };
};

export type SwitchboardState = {
  generatedAt: string;
  uptimeSeconds: number;
  strategy: string;
  policySource: string;
  backends: BackendState[];
};

export function getSwitchboardState(): Promise<SwitchboardState> { return apiFetch("/api/v1/state"); }

