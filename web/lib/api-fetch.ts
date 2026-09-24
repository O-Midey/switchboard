import { AppError, ERROR_CODES } from "./errors";

type WireError = { code: string; message: string; statusCode: number };

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${process.env.SWITCHBOARD_ADMIN_URL ?? "http://localhost:9090"}${path}`, { ...init, cache: "no-store" });
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === "AbortError") throw cause;
    throw new AppError(ERROR_CODES.NETWORK_ERROR, "Switchboard is not reachable.", 0, { cause, isOperational: true });
  }
  if (response.ok) return (await response.json()) as T;
  try {
    const error = (await response.json()) as WireError;
    throw new AppError(error.code, error.message, error.statusCode);
  } catch (cause) {
    if (cause instanceof AppError) throw cause;
    throw new AppError(ERROR_CODES.INTERNAL_ERROR, "Switchboard returned an unreadable response.", 500, { cause });
  }
}

