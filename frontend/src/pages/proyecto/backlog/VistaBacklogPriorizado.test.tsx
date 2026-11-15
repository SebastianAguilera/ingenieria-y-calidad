import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../../api/errors'
import { listarBacklogPriorizado } from '../../../api/historias'
import type { HistoriaPriorizada } from '../../../api/types'
import VistaBacklogPriorizado from './VistaBacklogPriorizado'

vi.mock('../../../api/historias', () => ({
  listarBacklogPriorizado: vi.fn(),
}))

const listarMock = vi.mocked(listarBacklogPriorizado)

function crearHistoria(
  id: number,
  titulo: string,
  extra: Partial<HistoriaPriorizada> = {},
): HistoriaPriorizada {
  return {
    id,
    codigo: `US-0${id}`,
    titulo,
    descripcion: '',
    prioridad: 'Media',
    estado: 'Backlog',
    story_points: 5,
    esfuerzo_real: null,
    valor_negocio: 8,
    criticidad_tiempo: 3,
    reduccion_riesgo: 2,
    roi: 1.6,
    wsjf: 2.6,
    creado_en: '2026-11-15T10:00:00Z',
    actualizado_en: '2026-11-15T10:00:00Z',
    ...extra,
  }
}

const prioridadAlta = crearHistoria(1, 'Prioridad alta', {
  valor_negocio: 21,
  criticidad_tiempo: 13,
  reduccion_riesgo: 8,
  roi: 4.2,
  wsjf: 8.4,
})

const prioridadMedia = crearHistoria(2, 'Prioridad media', {
  story_points: 3,
  valor_negocio: 5,
  criticidad_tiempo: 1,
  reduccion_riesgo: 1,
  roi: 1.6666666666666667,
  wsjf: 2.3333333333333335,
})

function renderizar() {
  return render(<VistaBacklogPriorizado proyectoId={1} />)
}

describe('VistaBacklogPriorizado', () => {
  beforeEach(() => {
    listarMock.mockReset()
  })

  it('E-10: muestra el estado de carga mientras consulta', () => {
    listarMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByRole('status')).toHaveTextContent('Cargando backlog priorizado…')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('E-05: muestra las columnas y respeta el orden recibido por WSJF', async () => {
    listarMock.mockResolvedValue([prioridadAlta, prioridadMedia])

    renderizar()

    expect(await screen.findByRole('table')).toBeInTheDocument()
    for (const encabezado of [
      'Código',
      'Título',
      'Estado',
      'Prioridad',
      'SP',
      'Valor',
      'Criticidad',
      'Riesgo',
      'ROI',
      'WSJF',
    ]) {
      expect(screen.getByRole('columnheader', { name: encabezado })).toBeInTheDocument()
    }

    const filas = screen.getAllByRole('row').slice(1)
    expect(within(filas[0]).getByText('Prioridad alta')).toBeInTheDocument()
    expect(within(filas[1]).getByText('Prioridad media')).toBeInTheDocument()
    expect(within(filas[0]).getByText('8.4')).toBeInTheDocument()
    expect(within(filas[1]).getByText('2.3333333333333335')).toBeInTheDocument()
  })

  it('E-06: muestra "—" en ROI, WSJF y SP cuando son null', async () => {
    const sinEstimacion = crearHistoria(3, 'Sin estimar', {
      story_points: null,
      valor_negocio: null,
      criticidad_tiempo: null,
      reduccion_riesgo: null,
      roi: null,
      wsjf: null,
    })
    listarMock.mockResolvedValue([sinEstimacion])

    renderizar()

    const fila = (await screen.findByText('Sin estimar')).closest('tr')
    expect(fila).not.toBeNull()
    expect(within(fila as HTMLElement).getAllByText('—')).toHaveLength(6)
  })

  it('E-07: no muestra las historias en Hecho', async () => {
    const terminada = crearHistoria(4, 'Historia terminada', { estado: 'Hecho' })
    listarMock.mockResolvedValue([prioridadAlta, terminada])

    renderizar()

    expect(await screen.findByText('Prioridad alta')).toBeInTheDocument()
    expect(screen.queryByText('Historia terminada')).not.toBeInTheDocument()
  })

  it('E-09: muestra el error de red y reintenta la consulta', async () => {
    listarMock.mockRejectedValueOnce(new Error('falla'))
    listarMock.mockResolvedValueOnce([prioridadAlta])
    const usuario = userEvent.setup()

    renderizar()

    expect(
      await screen.findByText('No se pudo cargar el backlog priorizado. Intentá nuevamente.'),
    ).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByText('Prioridad alta')).toBeInTheDocument()
    expect(listarMock).toHaveBeenCalledTimes(2)
  })

  it('E-10: muestra el estado vacío cuando no hay historias pendientes', async () => {
    listarMock.mockResolvedValue([])

    renderizar()

    expect(
      await screen.findByText('No hay historias pendientes de priorizar'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('muestra el mensaje de no acceso cuando el backend responde 404', async () => {
    listarMock.mockRejectedValue(
      new ApiError({
        status: 404,
        code: 'PROYECTO_NO_ENCONTRADO',
        message: 'el proyecto no existe',
      }),
    )

    renderizar()

    expect(
      await screen.findByText('No tenés acceso a este proyecto o no existe'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })
})
