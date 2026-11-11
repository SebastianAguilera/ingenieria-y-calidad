import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, request } from './client'
import { ApiError } from './errors'
import { CLAVE_SESION, EVENTO_SESION_EXPIRADA } from '../session/sesion'
import type { ListadoProyectos, Usuario } from './types'

type FetchFn = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

const fetchMock = vi.fn<FetchFn>()

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

describe('cliente HTTP', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('GET: dado un 200 con JSON, devuelve el cuerpo tipado', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({ body: JSON.stringify({ total: 2, proyectos: [] }) }),
    )

    const data = await api.get<ListadoProyectos>('/proyectos')

    expect(data.total).toBe(2)
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/proyectos')
    expect(init?.method).toBe('GET')
  })

  it('GET: codifica los query params y omite null y undefined', async () => {
    fetchMock.mockResolvedValue(respuestaMock({ body: '{}' }))

    await request('/historias', {
      query: { estado: 'Backlog', pagina: 2, activo: true, vacio: undefined, nulo: null },
    })

    const [url] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/historias?estado=Backlog&pagina=2&activo=true')
  })

  it('POST: envia Content-Type y serializa el body', async () => {
    fetchMock.mockResolvedValue(respuestaMock({ status: 201, body: '{"id":1}' }))

    await api.post('/proyectos', { nombre: 'Demo', fecha_inicio: '2026-01-01' })

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/proyectos')
    expect(init?.method).toBe('POST')
    expect(init?.body).toBe(JSON.stringify({ nombre: 'Demo', fecha_inicio: '2026-01-01' }))
    const headers = init?.headers as Record<string, string>
    expect(headers.Accept).toBe('application/json')
    expect(headers['Content-Type']).toBe('application/json')
  })

  it('DELETE: dado un 204, devuelve undefined', async () => {
    fetchMock.mockResolvedValue(respuestaMock({ status: 204 }))

    const resultado = await api.delete<undefined>('/proyectos/1')

    expect(resultado).toBeUndefined()
  })

  it('GET: dado un cuerpo vacio, devuelve undefined', async () => {
    fetchMock.mockResolvedValue(respuestaMock({ status: 200, body: '' }))

    const resultado = await api.get<undefined>('/vacio')

    expect(resultado).toBeUndefined()
  })

  it('GET: dado un JSON invalido, lanza ApiError RESPUESTA_INVALIDA', async () => {
    fetchMock.mockResolvedValue(respuestaMock({ status: 200, body: '<html>no json</html>' }))

    const error = await capturarError(api.get('/roto'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      name: 'ApiError',
      status: 200,
      code: 'RESPUESTA_INVALIDA',
    })
  })

  it('GET: dado un error del backend, mapea status, code, message y campo', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 404,
        statusText: 'Not Found',
        body: JSON.stringify({
          error: 'PROYECTO_NO_ENCONTRADO',
          mensaje: 'el proyecto no existe',
          campo: 'id',
        }),
      }),
    )

    const error = await capturarError(api.get('/proyectos/9'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 404,
      code: 'PROYECTO_NO_ENCONTRADO',
      message: 'el proyecto no existe',
      campo: 'id',
    })
  })

  it('GET: dado un error HTTP sin JSON, usa el fallback HTTP_<status>', async () => {
    fetchMock.mockResolvedValue(
      respuestaMock({ status: 500, statusText: 'Internal Server Error', body: 'boom' }),
    )

    const error = await capturarError(api.get('/falla'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 500,
      code: 'HTTP_500',
      message: 'Internal Server Error',
    })
  })

  it('GET: dado un fallo de red, lanza ApiError NETWORK_ERROR', async () => {
    fetchMock.mockRejectedValue(new Error('sin conexion'))

    const error = await capturarError(api.get('/red'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 0,
      code: 'NETWORK_ERROR',
      message: 'sin conexion',
    })
  })
})

describe('cliente HTTP con sesión (US-F08)', () => {
  const ana: Usuario = {
    id: 1,
    nombre: 'Ana Pérez',
    email: 'ana@example.com',
    creado_en: '2026-01-01T00:00:00Z',
  }

  const fetchMock = vi.fn<FetchFn>()

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

  function guardarSesion(token: string) {
    localStorage.setItem(CLAVE_SESION, JSON.stringify({ token, usuario: ana }))
  }

  beforeEach(() => {
    localStorage.clear()
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('adjunta Authorization Bearer cuando hay un token en la sesión', async () => {
    guardarSesion('jwt-vigente')
    fetchMock.mockResolvedValue(respuestaMock({ body: '{}' }))

    await api.get('/proyectos')

    const [, init] = fetchMock.mock.calls[0]
    const headers = init?.headers as Record<string, string>
    expect(headers.Authorization).toBe('Bearer jwt-vigente')
  })

  it('no envía Authorization cuando no hay sesión', async () => {
    fetchMock.mockResolvedValue(respuestaMock({ body: '{}' }))

    await api.get('/proyectos')

    const [, init] = fetchMock.mock.calls[0]
    const headers = init?.headers as Record<string, string>
    expect(headers.Authorization).toBeUndefined()
  })

  it('ante un 401 TOKEN_INVALIDO limpia la sesión y avisa con el evento', async () => {
    guardarSesion('jwt-vencido')
    const expirada = vi.fn()
    window.addEventListener(EVENTO_SESION_EXPIRADA, expirada)
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 401,
        body: JSON.stringify({
          error: 'TOKEN_INVALIDO',
          mensaje: 'el token de autenticación no es válido o ha expirado',
        }),
      }),
    )

    const error = await capturarError(api.get('/proyectos'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 401, code: 'TOKEN_INVALIDO' })
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
    expect(expirada).toHaveBeenCalledTimes(1)

    window.removeEventListener(EVENTO_SESION_EXPIRADA, expirada)
  })

  it('ante un 401 TOKEN_REQUERIDO limpia la sesión y avisa con el evento', async () => {
    guardarSesion('jwt-vencido')
    const expirada = vi.fn()
    window.addEventListener(EVENTO_SESION_EXPIRADA, expirada)
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 401,
        body: JSON.stringify({
          error: 'TOKEN_REQUERIDO',
          mensaje: 'la solicitud no incluye un token de autenticación',
        }),
      }),
    )

    const error = await capturarError(api.get('/proyectos'))

    expect(error).toMatchObject({ status: 401, code: 'TOKEN_REQUERIDO' })
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
    expect(expirada).toHaveBeenCalledTimes(1)

    window.removeEventListener(EVENTO_SESION_EXPIRADA, expirada)
  })

  it('un 401 CREDENCIALES_INVALIDAS no toca la sesión', async () => {
    guardarSesion('jwt-vigente')
    const expirada = vi.fn()
    window.addEventListener(EVENTO_SESION_EXPIRADA, expirada)
    fetchMock.mockResolvedValue(
      respuestaMock({
        status: 401,
        body: JSON.stringify({
          error: 'CREDENCIALES_INVALIDAS',
          mensaje: 'email o contraseña incorrectos',
        }),
      }),
    )

    const error = await capturarError(api.post('/auth/login', { email: 'a@b.c', password: 'x' }))

    expect(error).toMatchObject({ status: 401, code: 'CREDENCIALES_INVALIDAS' })
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual({
      token: 'jwt-vigente',
      usuario: ana,
    })
    expect(expirada).not.toHaveBeenCalled()

    window.removeEventListener(EVENTO_SESION_EXPIRADA, expirada)
  })
})
