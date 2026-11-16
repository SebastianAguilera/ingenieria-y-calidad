import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { descargarArchivo } from './client'
import { ApiError } from './errors'
import { CLAVE_SESION } from '../session/sesion'

// US-F23 (D-02): descarga de archivos binarios (PDF de US-24) con el mismo
// manejo de token y errores que `request`.

type FetchFn = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

const fetchMock = vi.fn<FetchFn>()

function respuestaArchivo(blob: Blob): Response {
  return { ok: true, status: 200, statusText: 'OK', blob: async () => blob } as unknown as Response
}

function respuestaError(status: number, body: string): Response {
  return { ok: false, status, statusText: '', text: async () => body } as unknown as Response
}

async function capturarError(promesa: Promise<unknown>): Promise<unknown> {
  try {
    await promesa
  } catch (error) {
    return error
  }
  throw new Error('se esperaba un error y la promesa se resolvio')
}

describe('descargarArchivo', () => {
  beforeEach(() => {
    localStorage.clear()
    fetchMock.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('hace GET a /api + path y devuelve el Blob de la respuesta', async () => {
    const pdf = new Blob(['%PDF-1.4'], { type: 'application/pdf' })
    fetchMock.mockResolvedValue(respuestaArchivo(pdf))

    const resultado = await descargarArchivo('/proyectos/1/reporte.pdf', 'application/pdf')

    const [url, init] = fetchMock.mock.calls[0]
    const headers = (init ?? {}).headers as Record<string, string>
    expect(url).toBe('/api/proyectos/1/reporte.pdf')
    expect(init?.method).toBe('GET')
    expect(headers.Accept).toBe('application/pdf')
    expect(resultado).toBe(pdf)
  })

  it('adjunta el token de la sesión y pasa la señal de aborto', async () => {
    localStorage.setItem(
      CLAVE_SESION,
      JSON.stringify({
        token: 'jwt-vigente',
        usuario: { id: 1, nombre: 'Ana', email: 'ana@example.com', creado_en: '2026-01-01T00:00:00Z' },
      }),
    )
    fetchMock.mockResolvedValue(respuestaArchivo(new Blob([])))
    const controlador = new AbortController()

    await descargarArchivo('/x.pdf', 'application/pdf', controlador.signal)

    const [, init] = fetchMock.mock.calls[0]
    const headers = (init ?? {}).headers as Record<string, string>
    expect(headers.Authorization).toBe('Bearer jwt-vigente')
    expect(init?.signal).toBe(controlador.signal)
  })

  it('traduce el cuerpo de error del backend a ApiError', async () => {
    fetchMock.mockResolvedValue(
      respuestaError(
        404,
        JSON.stringify({ error: 'PROYECTO_NO_ENCONTRADO', mensaje: 'el proyecto no existe' }),
      ),
    )

    const error = await capturarError(descargarArchivo('/x.pdf', 'application/pdf'))

    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ status: 404, code: 'PROYECTO_NO_ENCONTRADO' })
  })

  it('un fallo de red se traduce a ApiError con status 0', async () => {
    fetchMock.mockRejectedValue(new TypeError('sin conexion'))

    const error = await capturarError(descargarArchivo('/x.pdf', 'application/pdf'))

    expect(error).toMatchObject({ status: 0, code: 'NETWORK_ERROR' })
  })
})
