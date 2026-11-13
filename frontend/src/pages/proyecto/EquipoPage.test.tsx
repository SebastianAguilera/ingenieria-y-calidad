import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../api/errors'
import {
  agregarIntegrantePorEmail,
  cambiarRolIntegrante,
  listarIntegrantes,
  obtenerProyecto,
  quitarIntegrante,
} from '../../api/proyectos'
import type { Proyecto, Usuario, UsuarioProyecto } from '../../api/types'
import { SesionProvider } from '../../session/SesionContext'
import { CLAVE_SESION } from '../../session/sesion'
import EquipoPage from './EquipoPage'

vi.mock('../../api/proyectos', () => ({
  obtenerProyecto: vi.fn(),
  listarIntegrantes: vi.fn(),
  agregarIntegrantePorEmail: vi.fn(),
  cambiarRolIntegrante: vi.fn(),
  quitarIntegrante: vi.fn(),
}))

const obtenerMock = vi.mocked(obtenerProyecto)
const listarMock = vi.mocked(listarIntegrantes)
const agregarMock = vi.mocked(agregarIntegrantePorEmail)
const cambiarRolMock = vi.mocked(cambiarRolIntegrante)
const quitarMock = vi.mocked(quitarIntegrante)

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@utn.edu.ar',
  creado_en: '2026-01-01T00:00:00Z',
}

const ben: Usuario = {
  id: 7,
  nombre: 'Ben Gómez',
  email: 'ben@utn.edu.ar',
  creado_en: '2026-01-01T00:00:00Z',
}

const anaIntegrante: UsuarioProyecto = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@utn.edu.ar',
  rol: 'LIDER',
}

const benIntegrante: UsuarioProyecto = {
  id: 7,
  nombre: 'Ben Gómez',
  email: 'ben@utn.edu.ar',
  rol: 'MIEMBRO',
}

const crisIntegrante: UsuarioProyecto = {
  id: 9,
  nombre: 'Cris Díaz',
  email: 'cris@utn.edu.ar',
  rol: 'MIEMBRO',
}

function proyectoCon(rol: 'LIDER' | 'MIEMBRO', estado: 'Activo' | 'Cerrado' = 'Activo'): Proyecto {
  return {
    id: 1,
    nombre: 'Métricas',
    descripcion: '',
    fecha_inicio: '2026-01-01',
    fecha_fin: null,
    estado,
    cantidad_usuarios: 2,
    usuarios: [],
    rol_actual: rol,
    creado_en: '2026-01-01T00:00:00Z',
    actualizado_en: '2026-01-01T00:00:00Z',
  }
}

function renderizar() {
  return render(
    <SesionProvider>
      <MemoryRouter initialEntries={['/proyectos/1/equipo']}>
        <Routes>
          <Route path="/proyectos" element={<p>Listado de proyectos</p>} />
          <Route path="/proyectos/:id/equipo" element={<EquipoPage />} />
        </Routes>
      </MemoryRouter>
    </SesionProvider>,
  )
}

function iniciarSesion(usuario: Usuario) {
  localStorage.setItem(CLAVE_SESION, JSON.stringify({ token: 'jwt-vigente', usuario }))
}

async function esperarEquipo() {
  expect(await screen.findByRole('heading', { name: 'Equipo' })).toBeInTheDocument()
  expect(await screen.findByText('Ana Pérez')).toBeInTheDocument()
}

describe('EquipoPage', () => {
  beforeEach(() => {
    localStorage.clear()
    iniciarSesion(ana)
    obtenerMock.mockReset()
    listarMock.mockReset()
    agregarMock.mockReset()
    cambiarRolMock.mockReset()
    quitarMock.mockReset()
  })

  it('E-F12-01 muestra los integrantes con su rol', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])

    renderizar()
    await esperarEquipo()

    const filaBen = screen.getByRole('row', { name: /Ben Gómez/ })
    expect(within(filaBen).getByText('ben@utn.edu.ar')).toBeInTheDocument()
    expect(within(filaBen).getByText('MIEMBRO')).toBeInTheDocument()
  })

  it('E-F12-02 un líder ve alta, cambio de rol y baja', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])

    renderizar()
    await esperarEquipo()

    expect(screen.getByRole('button', { name: 'Agregar integrante' })).toBeInTheDocument()
    expect(screen.getByLabelText('Rol de Ben Gómez')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Quitar a Ben Gómez' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salir del proyecto' })).toBeInTheDocument()
  })

  it('E-F12-03 un miembro no ve acciones de gestión sobre otros', async () => {
    iniciarSesion(ben)
    obtenerMock.mockResolvedValue(proyectoCon('MIEMBRO'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])

    renderizar()
    await esperarEquipo()

    expect(screen.queryByRole('button', { name: 'Agregar integrante' })).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Rol de Ben Gómez')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Quitar a Ana Pérez' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salir del proyecto' })).toBeInTheDocument()
  })

  it('E-F12-05 el líder agrega por email y refresca la tabla', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock
      .mockResolvedValueOnce([anaIntegrante, benIntegrante])
      .mockResolvedValueOnce([anaIntegrante, benIntegrante, crisIntegrante])
    agregarMock.mockResolvedValue(crisIntegrante)
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.click(screen.getByRole('button', { name: 'Agregar integrante' }))
    const dialogo = screen.getByRole('dialog')
    await usuario.type(within(dialogo).getByLabelText('Email'), 'cris@utn.edu.ar')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Agregar' }))

    expect(agregarMock).toHaveBeenCalledWith(1, 'cris@utn.edu.ar', 'MIEMBRO')
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(await screen.findByText('Cris Díaz')).toBeInTheDocument()
  })

  it('E-F12-06 muestra el error de email inexistente junto al campo', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante])
    agregarMock.mockRejectedValue(
      new ApiError({
        status: 404,
        code: 'USUARIO_NO_ENCONTRADO',
        message: 'No existe un usuario registrado con ese email',
        campo: 'email',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.click(screen.getByRole('button', { name: 'Agregar integrante' }))
    const dialogo = screen.getByRole('dialog')
    await usuario.type(within(dialogo).getByLabelText('Email'), 'dario@utn.edu.ar')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Agregar' }))

    const campo = within(dialogo).getByLabelText('Email')
    await waitFor(() =>
      expect(campo).toHaveAccessibleDescription('No existe un usuario registrado con ese email'),
    )
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('E-F12-08 no envía un email vacío', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante])
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.click(screen.getByRole('button', { name: 'Agregar integrante' }))
    await usuario.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Agregar' }))

    expect(await screen.findByText('El email es obligatorio')).toBeInTheDocument()
    expect(agregarMock).not.toHaveBeenCalled()
  })

  it('E-F12-09 el líder cambia el rol y la fila se actualiza', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])
    cambiarRolMock.mockResolvedValue({ id: 7, rol: 'LIDER' })
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.selectOptions(screen.getByLabelText('Rol de Ben Gómez'), 'LIDER')

    expect(cambiarRolMock).toHaveBeenCalledWith(1, 7, 'LIDER')
    await waitFor(() =>
      expect(screen.getByLabelText('Rol de Ben Gómez')).toHaveValue('LIDER'),
    )
  })

  it('E-F12-10 cambiar al mismo rol no genera una petición', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.selectOptions(screen.getByLabelText('Rol de Ben Gómez'), 'MIEMBRO')

    expect(cambiarRolMock).not.toHaveBeenCalled()
  })

  it('E-F12-11 el último líder no puede degradarse', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante])
    cambiarRolMock.mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'ULTIMO_LIDER',
        message: 'el proyecto debe conservar al menos un líder',
      }),
    )
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.selectOptions(screen.getByLabelText('Rol de Ana Pérez'), 'MIEMBRO')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'el proyecto debe conservar al menos un líder',
    )
    expect(screen.getByLabelText('Rol de Ana Pérez')).toHaveValue('LIDER')
  })

  it('E-F12-12 el líder quita a otro integrante tras confirmar', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])
    quitarMock.mockResolvedValue(undefined)
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.click(screen.getByRole('button', { name: 'Quitar a Ben Gómez' }))
    const dialogo = screen.getByRole('dialog')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Quitar' }))

    expect(quitarMock).toHaveBeenCalledWith(1, 7)
    await waitFor(() => expect(screen.queryByText('Ben Gómez')).not.toBeInTheDocument())
  })

  it('E-F12-04 salir del proyecto vuelve al listado', async () => {
    iniciarSesion(ben)
    obtenerMock.mockResolvedValue(proyectoCon('MIEMBRO'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])
    quitarMock.mockResolvedValue(undefined)
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.click(screen.getByRole('button', { name: 'Salir del proyecto' }))
    const dialogo = screen.getByRole('dialog')
    await usuario.click(within(dialogo).getByRole('button', { name: 'Salir' }))

    expect(quitarMock).toHaveBeenCalledWith(1, 7)
    expect(await screen.findByText('Listado de proyectos')).toBeInTheDocument()
  })

  it('E-F12-13 un proyecto cerrado es de solo lectura', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER', 'Cerrado'))
    listarMock.mockResolvedValue([anaIntegrante, benIntegrante])

    renderizar()
    await esperarEquipo()

    expect(screen.getByText(/Reabrí el proyecto para gestionar el equipo/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Agregar integrante' })).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Rol de Ben Gómez')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Quitar a Ben Gómez' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Salir del proyecto' })).not.toBeInTheDocument()
  })

  it('E-F12-14 sin acceso muestra el mensaje de proyecto inexistente', async () => {
    obtenerMock.mockRejectedValue(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'no existe' }),
    )
    listarMock.mockRejectedValue(
      new ApiError({ status: 404, code: 'PROYECTO_NO_ENCONTRADO', message: 'no existe' }),
    )

    renderizar()

    expect(
      await screen.findByText('No tenés acceso a este proyecto o no existe'),
    ).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Reintentar' })).not.toBeInTheDocument()
  })

  it('E-F12-15 un error de red conserva el email del formulario', async () => {
    obtenerMock.mockResolvedValue(proyectoCon('LIDER'))
    listarMock.mockResolvedValue([anaIntegrante])
    agregarMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'Failed to fetch' }),
    )
    const usuario = userEvent.setup()

    renderizar()
    await esperarEquipo()

    await usuario.click(screen.getByRole('button', { name: 'Agregar integrante' }))
    const dialogo = screen.getByRole('dialog')
    await usuario.type(within(dialogo).getByLabelText('Email'), 'cris@utn.edu.ar')
    fireEvent.submit(within(dialogo).getByRole('button', { name: 'Agregar' }).closest('form')!)

    expect(await within(dialogo).findByRole('alert')).toHaveTextContent('No se pudo conectar')
    expect(within(dialogo).getByLabelText('Email')).toHaveValue('cris@utn.edu.ar')
  })
})
