import { ApiError } from './errors'

const API_BASE = '/api'

const CODIGO_ERROR_RED = 'NETWORK_ERROR'
const CODIGO_RESPUESTA_INVALIDA = 'RESPUESTA_INVALIDA'

export type QueryValue = string | number | boolean | null | undefined

export type MetodoHTTP = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export type RequestOptions = {
  method?: MetodoHTTP
  body?: unknown
  query?: Record<string, QueryValue>
  headers?: Record<string, string>
  signal?: AbortSignal
}

type OpcionesSinMetodoNiBody = Omit<RequestOptions, 'method' | 'body'>

type CuerpoErrorBackend = {
  error?: unknown
  mensaje?: unknown
  campo?: unknown
}

type ErrorBackendParseado = {
  code?: string
  message?: string
  campo?: string
}

function construirURL(path: string, query?: Record<string, QueryValue>): string {
  const url = `${API_BASE}${path}`
  if (!query) {
    return url
  }

  const params = new URLSearchParams()
  for (const [clave, valor] of Object.entries(query)) {
    if (valor === undefined || valor === null) {
      continue
    }
    params.append(clave, String(valor))
  }

  const queryString = params.toString()
  return queryString ? `${url}?${queryString}` : url
}

function esObjetoJSON(valor: unknown): valor is CuerpoErrorBackend {
  return typeof valor === 'object' && valor !== null
}

function parsearErrorBackend(cuerpo: unknown): ErrorBackendParseado {
  if (!esObjetoJSON(cuerpo)) {
    return {}
  }

  const code =
    typeof cuerpo.error === 'string' && cuerpo.error.trim() !== '' ? cuerpo.error : undefined
  const message = typeof cuerpo.mensaje === 'string' ? cuerpo.mensaje : undefined
  const campo =
    typeof cuerpo.campo === 'string' && cuerpo.campo.trim() !== '' ? cuerpo.campo : undefined

  return { code, message, campo }
}

async function leerCuerpo(response: Response): Promise<string> {
  try {
    return await response.text()
  } catch {
    return ''
  }
}

function construirApiErrorDeRespuesta(response: Response, texto: string): ApiError {
  let cuerpo: unknown
  if (texto.trim() !== '') {
    try {
      cuerpo = JSON.parse(texto)
    } catch {
      cuerpo = undefined
    }
  }

  const parseado = parsearErrorBackend(cuerpo)

  return new ApiError({
    status: response.status,
    code: parseado.code ?? `HTTP_${response.status}`,
    message:
      parseado.message ??
      (response.statusText !== '' ? response.statusText : `Error HTTP ${response.status}`),
    campo: parseado.campo,
  })
}

function esErrorDeAborto(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError'
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, query, headers, signal } = options
  const tieneBody = body !== undefined

  let response: Response
  try {
    response = await fetch(construirURL(path, query), {
      method,
      headers: {
        Accept: 'application/json',
        ...(tieneBody ? { 'Content-Type': 'application/json' } : {}),
        ...headers,
      },
      body: tieneBody ? JSON.stringify(body) : undefined,
      signal,
    })
  } catch (error) {
    if (esErrorDeAborto(error)) {
      throw error
    }
    throw new ApiError({
      status: 0,
      code: CODIGO_ERROR_RED,
      message: error instanceof Error ? error.message : 'No se pudo conectar con el servidor',
    })
  }

  if (!response.ok) {
    const texto = await leerCuerpo(response)
    throw construirApiErrorDeRespuesta(response, texto)
  }

  if (response.status === 204 || response.status === 205) {
    return undefined as T
  }

  const texto = await leerCuerpo(response)
  if (texto.trim() === '') {
    return undefined as T
  }

  try {
    return JSON.parse(texto) as T
  } catch {
    throw new ApiError({
      status: response.status,
      code: CODIGO_RESPUESTA_INVALIDA,
      message: 'La respuesta del servidor no es un JSON válido',
    })
  }
}

export const api = {
  get<T>(path: string, options?: OpcionesSinMetodoNiBody): Promise<T> {
    return request<T>(path, { ...options, method: 'GET' })
  },

  post<T>(path: string, body?: unknown, options?: OpcionesSinMetodoNiBody): Promise<T> {
    return request<T>(path, { ...options, method: 'POST', body })
  },

  put<T>(path: string, body?: unknown, options?: OpcionesSinMetodoNiBody): Promise<T> {
    return request<T>(path, { ...options, method: 'PUT', body })
  },

  patch<T>(path: string, body?: unknown, options?: OpcionesSinMetodoNiBody): Promise<T> {
    return request<T>(path, { ...options, method: 'PATCH', body })
  },

  delete<T>(path: string, options?: OpcionesSinMetodoNiBody): Promise<T> {
    return request<T>(path, { ...options, method: 'DELETE' })
  },
}
