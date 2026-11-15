import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { login, registrar } from './auth'
import { ApiError } from './errors'
import { CLAVE_SESION } from '../session/sesion'
import type { Usuario } from './types'

type FetchFn = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

const fetchMock = vi.fn<FetchFn>()

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

function respuestaMock({
  status = 200,
  statusText = '',
  body = '',
}: {
  status?: number
  statusText?: string
  body?: string
} = {}): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText,
    text: async () => body,
  } as unknown as Response
}

async function capturarError(promesa: Promise<unknown>): Promise<unknown> {
  try {
    await promesa
  } catch (error) {
    return error
  }
  throw new Error('se esperaba un error y la promesa se resolvio')
}

describe('api/auth', () => {
  beforeEach(() => {
    localStorage.clear()
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('login: envía POST /api/auth/login con email y contraseña', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({ body: JSON.stringify({ token: 'jwt-vigente', usuario: ana }) }),
    )

    const sesion = await login('ana@example.com', 'Secreto123')

    expect(sesion).toEqual({ token: 'jwt-vigente', usuario: ana })
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/auth/login')
    expect(init?.method).toBe('POST')
    expect(init?.body).toBe(JSON.stringify({ email: 'ana@example.com', password: 'Secreto123' }))
    const headers = init?.headers as Record<string, string>
    expect(headers['Content-Type']).toBe('application/json')
  })

  it('login: sin sesión previa no envía Authorization', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({ body: JSON.stringify({ token: 'jwt-vigente', usuario: ana }) }),
    )

    await login('ana@example.com', 'Secreto123')

    const [, init] = fetchMock.mock.calls[0]
    const headers = init?.headers as Record<string, string>
    expect(headers.Authorization).toBeUndefined()
  })

  it('login: con sesión previa el token viejo viaja como Bearer', async () => {
    localStorage.setItem(
      CLAVE_SESION,
      JSON.stringify({ token: 'token-anterior', usuario: ana }),
    )
    fetchMock.mockResolvedValue(
      respuestaMock({ body: JSON.stringify({ token: 'jwt-vigente', usuario: ana }) }),
    )

    await login('ana@example.com', 'Secreto123')

    const [, init] = fetchMock.mock.calls[0]
    const headers = init?.headers as Record<string, string>
    expect(headers.Authorization).toBe('Bearer token-anterior')
  })

  it('login: un 401 CREDENCIALES_INVALIDAS no limpia la sesión guardada', async () => {
    localStorage.setItem(
      CLAVE_SESION,
      JSON.stringify({ token: 'token-anterior', usuario: ana }),
    )
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 401,
        body: JSON.stringify({
          error: 'CREDENCIALES_INVALIDAS',
          mensaje: 'email o contraseña incorrectos',
        }),
      }),
    )

    const error = await capturarError(login('ana@example.com', 'mala'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 401, code: 'CREDENCIALES_INVALIDAS' })
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual({
      token: 'token-anterior',
      usuario: ana,
    })
  })

  it('login: un error de red se propaga como NETWORK_ERROR', async () => {
    fetchMock.mockRejectedValue(new Error('sin conexion'))

    const error = await capturarError(login('ana@example.com', 'Secreto123'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 0, code: 'NETWORK_ERROR' })
  })

  it('registrar: envía POST /api/auth/registro con nombre, email y contraseña', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({ body: JSON.stringify({ token: 'jwt-nuevo', usuario: ana }) }),
    )

    const sesion = await registrar('Ana Pérez', 'ana@example.com', 'Secreto123')

    expect(sesion).toEqual({ token: 'jwt-nuevo', usuario: ana })
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/auth/registro')
    expect(init?.method).toBe('POST')
    expect(init?.body).toBe(
      JSON.stringify({ nombre: 'Ana Pérez', email: 'ana@example.com', password: 'Secreto123' }),
    )
    const headers = init?.headers as Record<string, string>
    expect(headers['Content-Type']).toBe('application/json')
  })

  it('registrar: la ruta es pública y no envía Authorization', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({ body: JSON.stringify({ token: 'jwt-nuevo', usuario: ana }) }),
    )

    await registrar('Ana Pérez', 'ana@example.com', 'Secreto123')

    const [, init] = fetchMock.mock.calls[0]
    const headers = init?.headers as Record<string, string>
    expect(headers.Authorization).toBeUndefined()
  })

  it('registrar: un 422 USUARIO_DUPLICADO se propaga con el campo email', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 422,
        body: JSON.stringify({
          error: 'USUARIO_DUPLICADO',
          mensaje: 'el email ya está registrado',
          campo: 'email',
        }),
      }),
    )

    const error = await capturarError(registrar('Ana Pérez', 'ana@example.com', 'Secreto123'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 422, code: 'USUARIO_DUPLICADO', campo: 'email' })
  })

  it('registrar: un 422 VALIDACION se propaga con el campo resaltado', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 422,
        body: JSON.stringify({
          error: 'VALIDACION',
          mensaje: 'el email del usuario es obligatorio y debe tener un formato válido',
          campo: 'email',
        }),
      }),
    )

    const error = await capturarError(registrar('Ana Pérez', 'no-es-email', 'Secreto123'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 422,
      code: 'VALIDACION',
      campo: 'email',
      message: 'el email del usuario es obligatorio y debe tener un formato válido',
    })
  })

  it('registrar: un error de red se propaga como NETWORK_ERROR', async () => {
    fetchMock.mockRejectedValue(new Error('sin conexion'))

    const error = await capturarError(registrar('Ana Pérez', 'ana@example.com', 'Secreto123'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 0, code: 'NETWORK_ERROR' })
  })
})
