import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import { descargarPdfReporte, obtenerReporteProyecto, obtenerReporteSprint } from '../../api/reportes'
import { listarSprints } from '../../api/sprints'
import type { ReporteProyecto, ReporteSprint, Sprint } from '../../api/types'
import BarraLateralProyecto from '../../layouts/BarraLateralProyecto'
import ReportePage from './ReportePage'
import { guardarArchivo } from './reporte/guardarArchivo'

vi.mock('../../api/reportes', () => ({
  obtenerReporteProyecto: vi.fn(),
  obtenerReporteSprint: vi.fn(),
  descargarPdfReporte: vi.fn(),
}))
vi.mock('../../api/sprints', () => ({ listarSprints: vi.fn() }))
vi.mock('./reporte/guardarArchivo', () => ({ guardarArchivo: vi.fn() }))

const reporteProyectoMock = vi.mocked(obtenerReporteProyecto)
const reporteSprintMock = vi.mocked(obtenerReporteSprint)
const pdfMock = vi.mocked(descargarPdfReporte)
const listarSprintsMock = vi.mocked(listarSprints)
const guardarMock = vi.mocked(guardarArchivo)

function sprint(id: number, estado: Sprint['estado']): Sprint {
  return {
    id,
    identificador: `Sprint ${id}`,
    sprint_goal: `Objetivo ${id}`,
    estado,
    fecha_inicio: '2026-10-06',
    fecha_fin: '2026-10-20',
    creado_en: '2026-10-01T00:00:00Z',
    actualizado_en: '2026-10-01T00:00:00Z',
  }
}

const sprints = [sprint(1, 'Cerrado'), sprint(2, 'Activo')]

const reporteProyecto: ReporteProyecto = {
  proyecto_id: 1,
  nombre: 'Gestión Ágil',
  descripcion: 'Sistema de métricas',
  fecha_inicio: '2026-09-15',
  fecha_fin: null,
  estado: 'Activo',
  base: 'estado_actual',
  historias: [
    {
      id: 10,
      codigo: 'US-1',
      titulo: 'Login',
      estado: 'Hecho',
      story_points: 5,
      horas_estimadas: 8,
      horas_reales: 6,
    },
    {
      id: 11,
      codigo: 'US-2',
      titulo: 'Logout',
      estado: 'Backlog',
      story_points: null,
      horas_estimadas: null,
      horas_reales: null,
    },
  ],
  metricas: {
    proyecto_id: 1,
    base: 'estado_actual',
    velocidad: null,
    porcentaje_historias_completadas: 0.5,
    horas_estimadas: 8,
    horas_reales: 6,
    desviacion_esfuerzo: -0.25,
    defectos_detectados: 1,
    defectos_resueltos: 0,
    ratio_defectos: 1,
  },
  defectos: [
    {
      id: 3,
      codigo: 'DEF-3',
      descripcion: 'Falla el login',
      severidad: 'major',
      estado: 'open',
      sprint_deteccion_id: 1,
      sprint_resolucion_id: null,
    },
  ],
}

const reporteSprint: ReporteSprint = {
  proyecto_id: 1,
  sprint_id: 2,
  identificador: 'Sprint 2',
  sprint_goal: 'Cerrar autenticación',
  fecha_inicio: '2026-10-21',
  fecha_fin: '2026-11-04',
  estado: 'Activo',
  base: 'estado_actual',
  historias: [
    {
      id: 12,
      codigo: 'US-3',
      titulo: 'Recuperar clave',
      estado: 'En Progreso',
      story_points: 3,
      horas_estimadas: 4,
      horas_reales: null,
    },
  ],
  metricas: {
    sprint_id: 2,
    base: 'estado_actual',
    story_points_planificados: 3,
    story_points_completados: 0,
    porcentaje_completado: 0,
    historias_completadas: 0,
    horas_estimadas: 4,
    horas_reales: 0,
    desviacion_esfuerzo: null,
    defectos_detectados: 0,
    defectos_resueltos: 0,
  },
  defectos: [],
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1/reporte']}>
      <Routes>
        <Route path="/proyectos/:id/reporte" element={<ReportePage />} />
      </Routes>
    </MemoryRouter>,
  )
}

function leerBlob(blob: Blob): Promise<string> {
  return new Promise((resolver) => {
    const lector = new FileReader()
    lector.onload = () => resolver(String(lector.result))
    lector.readAsText(blob)
  })
}

async function esperarReporteDelProyecto() {
  await screen.findByRole('heading', { name: 'Gestión Ágil' })
}

function filaDe(nombreTabla: string, texto: string): HTMLElement {
  const tabla = screen.getByRole('table', { name: nombreTabla })
  const fila = within(tabla).getByText(texto).closest('tr')
  if (!fila) {
    throw new Error(`no hay fila con ${texto}`)
  }
  return fila
}

describe('ReportePage (US-F23)', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date(2026, 10, 16, 10, 0))
    reporteProyectoMock.mockReset()
    reporteSprintMock.mockReset()
    pdfMock.mockReset()
    listarSprintsMock.mockReset()
    guardarMock.mockReset()
    listarSprintsMock.mockResolvedValue(sprints)
    reporteProyectoMock.mockResolvedValue(reporteProyecto)
    reporteSprintMock.mockResolvedValue(reporteSprint)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('E-01 al entrar muestra el reporte del proyecto completo con su encabezado', async () => {
    renderizar()

    await esperarReporteDelProyecto()
    expect(reporteProyectoMock).toHaveBeenCalledWith(1, expect.any(AbortSignal))
    expect(screen.getByLabelText('Alcance')).toHaveValue('proyecto')
    expect(screen.getByText('Sistema de métricas')).toBeInTheDocument()
    expect(screen.getByText('15/09/2026')).toBeInTheDocument()
    expect(screen.getByText('Sin fecha de fin')).toBeInTheDocument()
    expect(screen.getByText('Activo')).toBeInTheDocument()
  })

  it('E-02 muestra métricas, historias y defectos del reporte', async () => {
    renderizar()
    await esperarReporteDelProyecto()

    expect(screen.getByRole('group', { name: 'Historias completadas' })).toHaveTextContent('50%')
    expect(screen.getByRole('group', { name: 'Desviación de esfuerzo' })).toHaveTextContent('−25%')

    const historia = filaDe('Historias del reporte', 'US-1')
    expect(historia).toHaveTextContent('Login')
    expect(historia).toHaveTextContent('Hecho')
    expect(historia).toHaveTextContent('5')
    expect(historia).toHaveTextContent('8 h')
    expect(historia).toHaveTextContent('6 h')

    const sinValores = filaDe('Historias del reporte', 'US-2')
    expect(within(sinValores).getAllByText('—')).toHaveLength(3)

    const defecto = filaDe('Defectos del reporte', 'DEF-3')
    expect(defecto).toHaveTextContent('Falla el login')
    expect(defecto).toHaveTextContent('Mayor')
    expect(defecto).toHaveTextContent('Abierto')
  })

  it('E-03 al elegir un sprint carga su reporte', async () => {
    const usuario = userEvent.setup()
    renderizar()
    await esperarReporteDelProyecto()

    await usuario.selectOptions(screen.getByLabelText('Alcance'), '2')

    await screen.findByRole('heading', { name: 'Sprint 2' })
    expect(reporteSprintMock).toHaveBeenCalledWith(1, 2, expect.any(AbortSignal))
    expect(screen.getByText('Cerrar autenticación')).toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'SP planificados' })).toHaveTextContent('3')
    expect(filaDe('Historias del reporte', 'US-3')).toHaveTextContent('Recuperar clave')
    expect(screen.queryByText('US-1')).not.toBeInTheDocument()
  })

  it('E-04 con listas vacías las tablas muestran su estado vacío', async () => {
    reporteProyectoMock.mockResolvedValue({ ...reporteProyecto, historias: [], defectos: [] })
    renderizar()
    await esperarReporteDelProyecto()

    expect(screen.getByText('Este reporte no tiene historias')).toBeInTheDocument()
    expect(screen.getByText('Este reporte no tiene defectos')).toBeInTheDocument()
  })

  it('E-05 una métrica null se muestra como "Sin datos"', async () => {
    renderizar()
    await esperarReporteDelProyecto()

    expect(screen.getByRole('group', { name: 'Velocidad' })).toHaveTextContent('Sin datos')
  })

  it('E-06 y E-13 "Descargar JSON" guarda la respuesta de la API con el nombre del proyecto', async () => {
    const usuario = userEvent.setup()
    renderizar()
    await esperarReporteDelProyecto()

    await usuario.click(screen.getByRole('button', { name: 'Descargar JSON' }))

    expect(guardarMock).toHaveBeenCalledTimes(1)
    const [blob, nombre] = guardarMock.mock.calls[0]
    expect(nombre).toBe('reporte-gestion-agil-proyecto-2026-11-16.json')
    expect(blob.type).toBe('application/json')
    expect(JSON.parse(await leerBlob(blob))).toEqual(reporteProyecto)
  })

  it('E-07 un error de red muestra el mensaje y permite reintentar', async () => {
    const usuario = userEvent.setup()
    reporteProyectoMock.mockRejectedValueOnce(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
    )
    renderizar()

    expect(
      await screen.findByText('No se pudo cargar el reporte. Intentá nuevamente.'),
    ).toBeInTheDocument()

    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    await esperarReporteDelProyecto()
    expect(reporteProyectoMock).toHaveBeenCalledTimes(2)
  })

  it('E-08 un 404 PROYECTO_NO_ENCONTRADO muestra el mensaje de acceso', async () => {
    listarSprintsMock.mockRejectedValue(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'el proyecto no existe' }),
    )
    reporteProyectoMock.mockRejectedValue(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'el proyecto no existe' }),
    )
    renderizar()

    expect(
      await screen.findByText('No tenés acceso a este proyecto o no existe'),
    ).toBeInTheDocument()
  })

  it('E-09 "Descargar PDF" del proyecto pide el PDF y lo guarda', async () => {
    const usuario = userEvent.setup()
    const pdf = new Blob(['%PDF'], { type: 'application/pdf' })
    pdfMock.mockResolvedValue(pdf)
    renderizar()
    await esperarReporteDelProyecto()

    await usuario.click(screen.getByRole('button', { name: 'Descargar PDF' }))

    await waitFor(() => {
      expect(guardarMock).toHaveBeenCalledWith(pdf, 'reporte-gestion-agil-proyecto-2026-11-16.pdf')
    })
    expect(pdfMock).toHaveBeenCalledWith(1, null)
  })

  it('E-10 y E-14 "Descargar PDF" de un sprint pide su PDF con el nombre del sprint', async () => {
    const usuario = userEvent.setup()
    const pdf = new Blob(['%PDF'], { type: 'application/pdf' })
    pdfMock.mockResolvedValue(pdf)
    renderizar()
    await esperarReporteDelProyecto()
    await usuario.selectOptions(screen.getByLabelText('Alcance'), '2')
    await screen.findByRole('heading', { name: 'Sprint 2' })

    await usuario.click(screen.getByRole('button', { name: 'Descargar PDF' }))

    await waitFor(() => {
      expect(guardarMock).toHaveBeenCalledWith(pdf, 'reporte-proyecto-1-sprint-sprint-2-2026-11-16.pdf')
    })
    expect(pdfMock).toHaveBeenCalledWith(1, 2)
  })

  it('mientras se genera el PDF el botón queda deshabilitado', async () => {
    const usuario = userEvent.setup()
    let resolver: (blob: Blob) => void = () => {}
    pdfMock.mockReturnValue(new Promise((r) => (resolver = r)))
    renderizar()
    await esperarReporteDelProyecto()

    await usuario.click(screen.getByRole('button', { name: 'Descargar PDF' }))

    expect(screen.getByRole('button', { name: 'Generando PDF…' })).toBeDisabled()
    resolver(new Blob([]))
    expect(await screen.findByRole('button', { name: 'Descargar PDF' })).toBeEnabled()
  })

  it('E-11 si falla el PDF se ve el error y el reporte sigue visible', async () => {
    const usuario = userEvent.setup()
    pdfMock.mockRejectedValue(new ApiError({ status: 500, code: 'HTTP_500', message: 'error' }))
    renderizar()
    await esperarReporteDelProyecto()

    await usuario.click(screen.getByRole('button', { name: 'Descargar PDF' }))

    expect(
      await screen.findByText('No se pudo descargar el PDF. Intentá nuevamente.'),
    ).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Gestión Ágil' })).toBeInTheDocument()
    expect(guardarMock).not.toHaveBeenCalled()
  })

  it('E-12 los sprints de los defectos aparecen por nombre y la falta de sprint como —', async () => {
    renderizar()
    await esperarReporteDelProyecto()

    const celdas = within(filaDe('Defectos del reporte', 'DEF-3')).getAllByRole('cell')
    expect(celdas[4]).toHaveTextContent('Sprint 1')
    expect(celdas[5]).toHaveTextContent('—')
  })

  it('E-16 sin sprints el selector solo ofrece "Proyecto completo"', async () => {
    listarSprintsMock.mockResolvedValue([])
    renderizar()
    await esperarReporteDelProyecto()

    const opciones = within(screen.getByLabelText('Alcance')).getAllByRole('option')
    expect(opciones.map((opcion) => opcion.textContent)).toEqual(['Proyecto completo'])
  })
})

describe('Barra lateral del proyecto (US-F23)', () => {
  it('E-15 tiene la entrada "Reporte"', () => {
    render(
      <MemoryRouter initialEntries={['/proyectos/1/reporte']}>
        <BarraLateralProyecto />
      </MemoryRouter>,
    )

    expect(screen.getByRole('link', { name: 'Reporte' })).toHaveAttribute(
      'href',
      '/proyectos/1/reporte',
    )
  })
})
