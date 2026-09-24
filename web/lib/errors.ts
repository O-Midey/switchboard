export const ERROR_CODES = {
  NETWORK_ERROR: "NETWORK_ERROR",
  INTERNAL_ERROR: "INTERNAL_ERROR",
} as const;

export type ErrorCode = (typeof ERROR_CODES)[keyof typeof ERROR_CODES] | string;

export class AppError extends Error {
  readonly isOperational: boolean;
  constructor(readonly code: ErrorCode, message: string, readonly statusCode: number, options?: { cause?: unknown; isOperational?: boolean }) {
    super(message, { cause: options?.cause });
    this.name = "AppError";
    this.isOperational = options?.isOperational ?? statusCode < 500;
  }
}

