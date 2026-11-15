import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import handler, { type RespuestaProxy, type SolicitudProxy } from './api/[...path].ts'

const fetchMock = vi.fn<typeof fetch>()

function respuestaUpstream({
  status = 200,
  body = '{"ok":true}',
  headers = { 'content-type': 'application/json' },
}: {
  status?: number
  body?: string
  headers?: Record<string, string>
} = {}): Response {
  return {
    status,
    headers: {
      forEach: (callback: (valor: string, nombre: string) => void) => {
        for (const [nombre, valor] of Object.entries(headers)) {
          callback(valor, nombre)
        }
      },
    },
    arrayBuffer: async () => Buffer.from(body),
  } as unknown as Response
}

function solicitud({
  method = 'GET',
  url = '/api/protegida',
  headers = {},
  body,
}: {
  method?: string
  url?: string
  headers?: Record<string, string>
  body?: string
} = {}): SolicitudProxy {
  return {
    method,
    url,
    headers,
    async *[Symbol.asyncIterator]() {
      if (body !== undefined) {
        yield Buffer.from(body)
      }
    },
  }
}

class RespuestaFalsa implements RespuestaProxy {
  statusCode = 200
  cabeceras: Record<string, string> = {}
  cuerpo: Buffer | string | undefined

  setHeader(nombre: string, valor: string): void {
    this.cabeceras[nombre.toLowerCase()] = valor
  }

  end(cuerpo?: Buffer | string): void {
    this.cuerpo = cuerpo
  }
}

describe('proxy /api', () => {
  beforeEach(() => {
    vi.stubEnv('API_ORIGIN', 'https://api.onrender.com/')
    vi.stubEnv('FRONTEND_KEY', 'clave-de-prueba')
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
  })

  it('reenvia la solicitud y agrega la clave del frontend', async () => {
    fetchMock.mockResolvedValue(respuestaUpstream())
    const respuesta = new RespuestaFalsa()

    await handler(
      solicitud({
        method: 'POST',
        url: '/api/proyectos?page=2',
        headers: {
          authorization: 'Bearer token',
          'content-type': 'application/json',
          host: 'app.vercel.app',
          'x-frontend-key': 'intruso',
        },
        body: '{"nombre":"x"}',
      }),
      respuesta,
    )

    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [destino, init] = fetchMock.mock.calls[0]
    expect(destino).toBe('https://api.onrender.com/api/proyectos?page=2')
    expect(init?.method).toBe('POST')

    const cabeceras = init?.headers as Record<string, string>
    expect(cabeceras.authorization).toBe('Bearer token')
    expect(cabeceras['content-type']).toBe('application/json')
    expect(cabeceras['X-Frontend-Key']).toBe('clave-de-prueba')
    expect(cabeceras.host).toBeUndefined()
    expect(init?.body).toEqual(Buffer.from('{"nombre":"x"}'))
  })

  it('devuelve estado, cabeceras y cuerpo de la API', async () => {
    fetchMock.mockResolvedValue(
      respuestaUpstream({
        status: 201,
        body: '{"id":1}',
        headers: {
          'content-type': 'application/json',
          'x-trace': 'abc',
          connection: 'keep-alive',
          'content-length': '999',
        },
      }),
    )
    const respuesta = new RespuestaFalsa()

    await handler(solicitud(), respuesta)

    expect(respuesta.statusCode).toBe(201)
    expect(respuesta.cabeceras['content-type']).toBe('application/json')
    expect(respuesta.cabeceras['x-trace']).toBe('abc')
    expect(respuesta.cabeceras.connection).toBeUndefined()
    expect(respuesta.cabeceras['content-length']).toBeUndefined()
    expect(respuesta.cuerpo).toEqual(Buffer.from('{"id":1}'))
  })

  it('no envia cuerpo en GET', async () => {
    fetchMock.mockResolvedValue(respuestaUpstream())
    const respuesta = new RespuestaFalsa()

    await handler(solicitud({ body: 'ignorado' }), respuesta)

    const [, init] = fetchMock.mock.calls[0]
    expect(init?.body).toBeUndefined()
  })

  it.each(['API_ORIGIN', 'FRONTEND_KEY'] as const)('falla cerrado si falta %s', async (variable) => {
    vi.stubEnv(variable, '')
    const respuesta = new RespuestaFalsa()

    await handler(solicitud(), respuesta)

    expect(respuesta.statusCode).toBe(500)
    expect(respuesta.cabeceras['content-type']).toBe('application/json')
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('responde 504 cuando la API no responde a tiempo', async () => {
    const error = new Error('timeout')
    error.name = 'TimeoutError'
    fetchMock.mockRejectedValue(error)
    const respuesta = new RespuestaFalsa()

    await handler(solicitud(), respuesta)

    expect(respuesta.statusCode).toBe(504)
    expect(respuesta.cabeceras['content-type']).toBe('application/json')
  })
})
