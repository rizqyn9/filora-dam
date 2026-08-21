import { useAuthStore } from "@/features/auth/auth-store";

const BASE = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

interface ApiError {
  code: string;
  message: string;
}

export class ApiRequestError extends Error {
  constructor(
    public status: number,
    public error: ApiError,
  ) {
    super(error.message);
    this.name = "ApiRequestError";
  }
}

/** Thin fetch wrapper that handles the standard Filora API response envelope. */
export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const token = useAuthStore.getState().token;
  // Public auth routes live at /auth/*, protected routes at /api/v1/*
  const url = path.startsWith("/auth") ? path : `${BASE}${path}`;

  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });

  const body = await res.json();

  if (!res.ok || !body.success) {
    throw new ApiRequestError(res.status, body.error);
  }

  return body.data as T;
}
