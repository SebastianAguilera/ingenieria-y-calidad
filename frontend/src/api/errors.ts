export type ApiErrorInit = {
  status: number
  code: string
  message: string
  campo?: string
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly campo?: string

  constructor({ status, code, message, campo }: ApiErrorInit) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.campo = campo
  }
}

export function esApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}
