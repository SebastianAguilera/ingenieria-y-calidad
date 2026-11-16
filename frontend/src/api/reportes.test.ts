import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api, descargarArchivo } from './client'
import { descargarPdfReporte, obtenerReporteProyecto, obtenerReporteSprint } from './reportes'

vi.mock('./client', () => ({
  api: { get: vi.fn() },
  descargarArchivo: vi.fn(),
}))

const getMock = vi.mocked(api.get)
const descargarMock = vi.mocked(descargarArchivo)

describe('API de reportes (US-22 / US-24)', () => {
  beforeEach(() => {
    getMock.mockReset()
    descargarMock.mockReset()
  })

  it('obtenerReporteProyecto consulta GET /proyectos/:id/reporte', async () => {
    getMock.mockResolvedValue({ proyecto_id: 1 })
    const controlador = new AbortController()

    const reporte = await obtenerReporteProyecto(1, controlador.signal)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/reporte', { signal: controlador.signal })
    expect(reporte).toEqual({ proyecto_id: 1 })
  })

  it('obtenerReporteSprint consulta GET /proyectos/:id/sprints/:sprintId/reporte', async () => {
    getMock.mockResolvedValue({ sprint_id: 2 })

    await obtenerReporteSprint(1, 2)

    expect(getMock).toHaveBeenCalledWith('/proyectos/1/sprints/2/reporte', { signal: undefined })
  })

  it('descargarPdfReporte del proyecto pide /proyectos/:id/reporte.pdf', async () => {
    const pdf = new Blob(['%PDF'])
    descargarMock.mockResolvedValue(pdf)

    const resultado = await descargarPdfReporte(1, null)

    expect(descargarMock).toHaveBeenCalledWith('/proyectos/1/reporte.pdf', 'application/pdf', undefined)
    expect(resultado).toBe(pdf)
  })

  it('descargarPdfReporte de un sprint pide /proyectos/:id/sprints/:sprintId/reporte.pdf', async () => {
    descargarMock.mockResolvedValue(new Blob([]))

    await descargarPdfReporte(1, 2)

    expect(descargarMock).toHaveBeenCalledWith(
      '/proyectos/1/sprints/2/reporte.pdf',
      'application/pdf',
      undefined,
    )
  })
})
