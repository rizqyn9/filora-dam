import { env } from "@/lib/env";

/**
 * Standard API response envelope (see apps/api/API.md).
 */
export type ApiSuccess<T> = { success: true; data: T };
export type ApiFailure = {
  success: false;
  error: { code: string; message: string };
};
export type ApiEnvelope<T> = ApiSuccess<T> | ApiFailure;

/**
 * Thrown when the API returns `{ success: false }` or a non-2xx status.
 */
export class ApiError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Auth-agnostic token provider. Wire this to Clerk (or any auth) at bootstrap:
 *
 *   setAuthTokenProvider(() => clerk.session?.getToken() ?? null)
 *
 * Kept as an injectable so the API client never depends on a specific auth SDK.
 */
type TokenProvider = () => string | null | Promise<string | null>;

let tokenProvider: TokenProvider = () => null;

export function setAuthTokenProvider(provider: TokenProvider) {
  tokenProvider = provider;
}

/**
 * Called when the API responds 401, so the app can clear session / redirect.
 */
let onUnauthorized: (() => void) | null = null;

export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler;
}

export interface RequestOptions extends Omit<RequestInit, "body"> {
  /** JSON body (auto-serialized). Use `formData` for multipart. */
  body?: unknown;
  /** Raw FormData for file uploads (multipart/form-data). */
  formData?: FormData;
  /** Query parameters appended to the URL. */
  params?: Record<string, string | number | boolean | undefined>;
}

function buildUrl(path: string, params?: RequestOptions["params"]): string {
  const base = env.VITE_API_BASE_URL.replace(/\/$/, "");
  const url = `${base}${path.startsWith("/") ? path : `/${path}`}`;
  if (!params) return url;

  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined) search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `${url}?${qs}` : url;
}

/**
 * Core request function. Attaches the bearer token, sends JSON (or multipart),
 * unwraps the response envelope, and throws {@link ApiError} on failure.
 */
export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { body, formData, params, headers, ...rest } = options;

  const token = await tokenProvider();
  const finalHeaders = new Headers(headers);
  if (token) finalHeaders.set("Authorization", `Bearer ${token}`);

  let payload: BodyInit | undefined;
  if (formData) {
    payload = formData; // browser sets multipart boundary
  } else if (body !== undefined) {
    finalHeaders.set("Content-Type", "application/json");
    payload = JSON.stringify(body);
  }

  const response = await fetch(buildUrl(path, params), {
    ...rest,
    headers: finalHeaders,
    body: payload,
  });

  if (response.status === 401) onUnauthorized?.();

  // 204 No Content
  if (response.status === 204) return undefined as T;

  const json = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || !json.success) {
    const error = "error" in json ? json.error : undefined;
    throw new ApiError(
      error?.code ?? "UNKNOWN",
      error?.message ?? response.statusText,
      response.status,
    );
  }

  return json.data;
}

/** Convenience verb helpers. */
export const api = {
  get: <T>(path: string, options?: RequestOptions) =>
    apiRequest<T>(path, { ...options, method: "GET" }),
  post: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    apiRequest<T>(path, { ...options, method: "POST", body }),
  patch: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    apiRequest<T>(path, { ...options, method: "PATCH", body }),
  delete: <T>(path: string, options?: RequestOptions) =>
    apiRequest<T>(path, { ...options, method: "DELETE" }),
  upload: <T>(path: string, formData: FormData, options?: RequestOptions) =>
    apiRequest<T>(path, { ...options, method: "POST", formData }),
};
