import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { actualizarHistoria, crearHistoria, listarHistorias } from '../../api/historias'
import { ApiError } from '../../api/errors'
import { listarSesionesActivas } from '../../api/poker'
import type { Historia, SesionPokerActiva } from '../../api/types'
import BacklogPage from './BacklogPage'

vi.mock('../../api/historias', () => ({
  listarHistorias: vi.fn(),
  crearHistoria: vi.fn(),
  actualizarHistoria: vi.fn(),
}))

vi.mock('../../api/poker', () => ({
  listarSesionesActivas: vi.fn(),
}))

const listarMock = vi.mocked(listarHistorias)
const crearMock = vi.mocked(crearHistoria)
const actualizarMock = vi.mocked(actualizarHistoria)
const activasMock = vi.mocked(listarSesionesActivas)

const historiaBase: Historia = {
  id: 10,
  codigo: 'US-01',
  titulo: 'Registrar horas reales',
  descripcion: 'Permitir registrar horas por usuario.',
  prioridad: 'Alta',
  estado: 'Backlog',
  story_points: 5,
  esfuerzo_real: null,
  creado_en: '2026-10-01T09:15:00Z',
  actualizado_en: '2026-10-01T09:15:00Z',
}

const segundaHistoria: Historia = {
  ...historiaBase,
  id: 11,
  codigo: 'US-02',
  titulo: 'Cerrar el sprint',
  descripcion: 'Permitir cerrar un sprint activo.',
  prioridad: 'Media',
  estado: 'En Progreso',
}

const sesionActiva: SesionPokerActiva = {
  sesion_id: 7,
  estado: 'Votando',
  ronda_actual: 1,
  creado_por: { id: 2, nombre: 'Ana' },
  creado_en: '2026-11-13T10:00:00Z',
  votos_registrados: 1,
  total_miembros: 3,
  ya_vote: true,
  historia: { id: 10, codigo: 'US-01', titulo: 'Registrar horas reales' },
}

function renderizar() {
  return render(
    <MemoryRouter initialEntries={['/proyectos/1/backlog']}>
      <Routes>
        <Route path="/proyectos/:id/backlog" element={<BacklogPage />} />
        <Route
          path="/proyectos/:id/backlog/:historiaId/poker"
          element={<div>Pantalla de Planning Poker</div>}
        />
      </Routes>
    </MemoryRouter>,
  )
}

describe('BacklogPage', () => {
  beforeEach(() => {
    listarMock.mockReset()
    crearMock.mockReset()
    actualizarMock.mockReset()
    activasMock.mockReset()
    activasMock.mockResolvedValue([])
  })

  it('muestra el indicador de carga y no la tabla mientras carga', () => {
    listarMock.mockReturnValue(new Promise(() => {}))

    renderizar()

    expect(screen.getByText('Cargando historias…')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('muestra la tabla con las columnas Código, Título, Estado, Prioridad y Acciones', async () => {
    listarMock.mockResolvedValue([historiaBase])

    renderizar()

    expect(await screen.findByRole('table')).toBeInTheDocument()
    for (const encabezado of ['Código', 'Título', 'Estado', 'Prioridad', 'Acciones']) {
      expect(screen.getByRole('columnheader', { name: encabezado })).toBeInTheDocument()
    }
    expect(screen.getByText('US-01')).toBeInTheDocument()
    expect(screen.getByText('Registrar horas reales')).toBeInTheDocument()
  })

  it('muestra el estado vacío y abre el modal de alta', async () => {
    listarMock.mockResolvedValue([])
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByText('Todavía no hay historias')).toBeInTheDocument()
    await usuario.click(screen.getByRole('button', { name: 'Nueva historia' }))

    expect(screen.getByRole('dialog', { name: 'Nueva historia' })).toBeInTheDocument()
  })

  it('muestra el error de listado y vuelve a cargar con Reintentar', async () => {
    listarMock.mockRejectedValueOnce(new Error('falla'))
    listarMock.mockResolvedValueOnce([historiaBase])
    const usuario = userEvent.setup()

    renderizar()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'No se pudieron cargar las historias',
    )
    await usuario.click(screen.getByRole('button', { name: 'Reintentar' }))

    expect(await screen.findByRole('table')).toBeInTheDocument()
    expect(listarMock).toHaveBeenCalledTimes(2)
  })

  it('E-15: proyecto ajeno muestra "No tenés acceso a este proyecto o no existe" y oculta el backlog', async () => {
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
    expect(screen.queryByRole('button', { name: 'Nueva historia' })).not.toBeInTheDocument()
  })

  it('cambia el estado inline y envía el PATCH', async () => {
    listarMock.mockResolvedValue([historiaBase])
    actualizarMock.mockResolvedValue({ ...historiaBase, estado: 'Hecho' })
    const usuario = userEvent.setup()

    renderizar()

    const selector = await screen.findByRole('combobox', { name: 'Estado de US-01' })
    await usuario.selectOptions(selector, 'Hecho')

    expect(actualizarMock).toHaveBeenCalledWith(1, 10, { estado: 'Hecho' })
    await waitFor(() => expect(selector).toHaveValue('Hecho'))
  })

  it('cambia la prioridad inline y envía el PATCH', async () => {
    listarMock.mockResolvedValue([historiaBase])
    actualizarMock.mockResolvedValue({ ...historiaBase, prioridad: 'Baja' })
    const usuario = userEvent.setup()

    renderizar()

    const selector = await screen.findByRole('combobox', { name: 'Prioridad de US-01' })
    await usuario.selectOptions(selector, 'Baja')

    expect(actualizarMock).toHaveBeenCalledWith(1, 10, { prioridad: 'Baja' })
    await waitFor(() => expect(selector).toHaveValue('Baja'))
  })

  it('revierte el cambio inline cuando el PATCH falla', async () => {
    listarMock.mockResolvedValue([historiaBase])
    actualizarMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
    )
    const usuario = userEvent.setup()

    renderizar()

    const selector = await screen.findByRole('combobox', { name: 'Estado de US-01' })
    await usuario.selectOptions(selector, 'Hecho')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'No se pudo actualizar la historia',
    )
    expect(selector).toHaveValue('Backlog')
  })

  it('crea una historia desde el modal y la agrega a la tabla', async () => {
    listarMock.mockResolvedValue([])
    crearMock.mockResolvedValue({
      ...historiaBase,
      id: 12,
      codigo: 'US-01',
      titulo: 'Nueva historia creada',
      descripcion: 'Descripción de la nueva historia.',
      prioridad: 'Baja',
      estado: 'En Progreso',
    })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Nueva historia' }))
    await usuario.type(screen.getByLabelText('Título'), 'Nueva historia creada')
    await usuario.type(screen.getByLabelText('Descripción'), 'Descripción de la nueva historia.')
    await usuario.selectOptions(screen.getByLabelText('Prioridad'), 'Baja')
    await usuario.selectOptions(screen.getByLabelText('Estado'), 'En Progreso')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(crearMock).toHaveBeenCalledWith(1, {
      titulo: 'Nueva historia creada',
      descripcion: 'Descripción de la nueva historia.',
      prioridad: 'Baja',
      estado: 'En Progreso',
    })
    expect(await screen.findByText('Nueva historia creada')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('edita una historia desde el modal y actualiza la fila', async () => {
    listarMock.mockResolvedValue([historiaBase])
    actualizarMock.mockResolvedValue({
      ...historiaBase,
      titulo: 'Registrar horas corregidas',
      descripcion: 'Descripción corregida.',
    })
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Editar US-01' }))
    const dialogo = screen.getByRole('dialog', { name: 'Editar historia' })
    const titulo = within(dialogo).getByLabelText('Título')
    await usuario.clear(titulo)
    await usuario.type(titulo, 'Registrar horas corregidas')
    const descripcion = within(dialogo).getByLabelText('Descripción')
    await usuario.clear(descripcion)
    await usuario.type(descripcion, 'Descripción corregida.')
    await usuario.click(screen.getByRole('button', { name: 'Guardar cambios' }))

    expect(actualizarMock).toHaveBeenCalledWith(1, 10, {
      titulo: 'Registrar horas corregidas',
      descripcion: 'Descripción corregida.',
      prioridad: 'Alta',
      estado: 'Backlog',
    })
    expect(await screen.findByText('Registrar horas corregidas')).toBeInTheDocument()
  })

  it('muestra el tablero con las tres columnas y las tarjetas agrupadas', async () => {
    listarMock.mockResolvedValue([historiaBase, segundaHistoria])
    const usuario = userEvent.setup()

    renderizar()

    await screen.findByRole('table')
    await usuario.click(screen.getByRole('button', { name: 'Tablero' }))

    expect(screen.getByRole('region', { name: 'Columna Backlog' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Columna En Progreso' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Columna Hecho' })).toBeInTheDocument()

    const columnaBacklog = screen.getByRole('region', { name: 'Columna Backlog' })
    const columnaEnProgreso = screen.getByRole('region', { name: 'Columna En Progreso' })
    expect(within(columnaBacklog).getByText('Registrar horas reales')).toBeInTheDocument()
    expect(within(columnaEnProgreso).getByText('Cerrar el sprint')).toBeInTheDocument()
  })

  it('mueve una tarjeta con el selector accesible por teclado', async () => {
    listarMock.mockResolvedValue([historiaBase])
    actualizarMock.mockResolvedValue({ ...historiaBase, estado: 'En Progreso' })
    const usuario = userEvent.setup()

    renderizar()

    await screen.findByRole('table')
    await usuario.click(screen.getByRole('button', { name: 'Tablero' }))
    await usuario.selectOptions(screen.getByRole('combobox', { name: 'Mover US-01 a' }), 'En Progreso')

    expect(actualizarMock).toHaveBeenCalledWith(1, 10, { estado: 'En Progreso' })
  })

  it('muestra el error del modal cuando el backend rechaza un campo', async () => {
    listarMock.mockResolvedValue([])
    crearMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'VALIDACION',
        message: 'la prioridad de la historia debe ser "Alta", "Media" o "Baja"',
        campo: 'prioridad',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(await screen.findByRole('button', { name: 'Nueva historia' }))
    await usuario.type(screen.getByLabelText('Título'), 'Historia con error')
    await usuario.click(screen.getByRole('button', { name: 'Crear historia' }))

    expect(
      await screen.findByText('la prioridad de la historia debe ser "Alta", "Media" o "Baja"'),
    ).toBeInTheDocument()
    expect(screen.getByRole('dialog', { name: 'Nueva historia' })).toBeInTheDocument()
  })

  it('muestra "Planning Poker en curso" en historias con sesión activa', async () => {
    listarMock.mockResolvedValue([historiaBase])
    activasMock.mockResolvedValue([sesionActiva])

    renderizar()

    const tabla = await screen.findByRole('table')
    expect(await within(tabla).findByText('Planning Poker en curso')).toBeInTheDocument()
  })

  it('si falla la consulta de activas el backlog sigue visible sin distintivos', async () => {
    listarMock.mockResolvedValue([historiaBase])
    activasMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexión' }),
    )

    renderizar()

    expect(await screen.findByRole('table')).toBeInTheDocument()
    expect(screen.queryByText('Planning Poker en curso')).not.toBeInTheDocument()
  })

  it('navega a la entrada de Planning Poker desde la tabla', async () => {
    listarMock.mockResolvedValue([historiaBase])
    const usuario = userEvent.setup()

    renderizar()

    await usuario.click(
      await screen.findByRole('button', { name: 'Planning Poker de US-01' }),
    )

    expect(await screen.findByText('Pantalla de Planning Poker')).toBeInTheDocument()
  })

  it('navega a la entrada de Planning Poker desde el tablero', async () => {
    listarMock.mockResolvedValue([historiaBase])
    const usuario = userEvent.setup()

    renderizar()

    await screen.findByRole('table')
    await usuario.click(screen.getByRole('button', { name: 'Tablero' }))
    await usuario.click(screen.getByRole('button', { name: 'Planning Poker de US-01' }))

    expect(await screen.findByText('Pantalla de Planning Poker')).toBeInTheDocument()
  })
})
