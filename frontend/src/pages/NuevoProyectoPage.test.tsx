import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { asociarUsuario, crearProyecto } from '../api/proyectos'
import { ApiError } from '../api/errors'
import type { Proyecto, Usuario } from '../api/types'
import { CLAVE_SESION } from '../session/sesion'
import { SesionProvider } from '../session/SesionContext'
import NuevoProyectoPage from './NuevoProyectoPage'

vi.mock('../api/proyectos', () => ({
  crearProyecto: vi.fn(),
  asociarUsuario: vi.fn(),
}))

const crearMock = vi.mocked(crearProyecto)
const asociarMock = vi.mocked(asociarUsuario)

const ana: Usuario = {
  id: 3,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const creado: Proyecto = {
  id: 9,
  nombre: 'Proyecto Alfa',
  fecha_inicio: '2026-01-01',
  fecha_fin: null,
  estado: 'Activo',
  cantidad_usuarios: 0,
  usuarios: [],
  creado_en: '2026-01-01T00:00:00Z',
  actualizado_en: '2026-01-01T00:00:00Z',
}

function renderizar() {
  return render(
    <SesionProvider>
      <MemoryRouter initialEntries={['/proyectos/nuevo']}>
        <Routes>
          <Route path="/proyectos/nuevo" element={<NuevoProyectoPage />} />
          <Route path="/proyectos/:id" element={<p>Detalle del proyecto</p>} />
        </Routes>
      </MemoryRouter>
    </SesionProvider>,
  )
}

function completar({ nombre = 'Proyecto Alfa', inicio = '2026-01-01', fin = '' } = {}) {
  if (nombre) {
    fireEvent.change(screen.getByLabelText('Nombre'), { target: { value: nombre } })
  }
  fireEvent.change(screen.getByLabelText('Fecha de inicio'), { target: { value: inicio } })
  fireEvent.change(screen.getByLabelText('Fecha de fin'), { target: { value: fin } })
}

describe('NuevoProyectoPage', () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem(CLAVE_SESION, JSON.stringify({ token: 'jwt-vigente', usuario: ana }))
    crearMock.mockReset()
    asociarMock.mockReset()
  })

  it('no envía y muestra el error cuando el nombre está vacío', async () => {
    const usuario = userEvent.setup()
    renderizar()
    completar({ nombre: '' })

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('El nombre es obligatorio')).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('no envía cuando el nombre supera los 150 caracteres', async () => {
    const usuario = userEvent.setup()
    renderizar()
    completar({ nombre: 'a'.repeat(151) })

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(
      await screen.findByText('El nombre no puede superar los 150 caracteres'),
    ).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('no envía cuando la fecha de fin es anterior a la de inicio', async () => {
    const usuario = userEvent.setup()
    renderizar()
    completar({ inicio: '2026-02-01', fin: '2026-01-01' })

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(
      await screen.findByText('La fecha de fin debe ser igual o posterior a la de inicio'),
    ).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('permite fechas iguales, crea el proyecto, asocia al usuario activo y navega', async () => {
    crearMock.mockResolvedValue(creado)
    asociarMock.mockResolvedValue({ id: ana.id, nombre: ana.nombre, email: ana.email })
    const usuario = userEvent.setup()
    renderizar()
    completar({ inicio: '2026-03-01', fin: '2026-03-01' })

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Detalle del proyecto')).toBeInTheDocument()
    expect(crearMock).toHaveBeenCalledWith({
      nombre: 'Proyecto Alfa',
      fecha_inicio: '2026-03-01',
      fecha_fin: '2026-03-01',
    })
    expect(asociarMock).toHaveBeenCalledWith(9, ana.id, 'Facilitador Ágil')
  })

  it('envía fecha_fin null cuando no se completa y recorta el nombre', async () => {
    crearMock.mockResolvedValue(creado)
    asociarMock.mockResolvedValue({ id: ana.id, nombre: ana.nombre, email: ana.email })
    const usuario = userEvent.setup()
    renderizar()
    completar({ nombre: '  Proyecto Alfa  ' })

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    await screen.findByText('Detalle del proyecto')
    expect(crearMock).toHaveBeenCalledWith({
      nombre: 'Proyecto Alfa',
      fecha_inicio: '2026-01-01',
      fecha_fin: null,
    })
  })

  it('deshabilita el botón y evita un segundo envío con doble clic', async () => {
    let resolver: (p: Proyecto) => void = () => {}
    crearMock.mockReturnValue(
      new Promise<Proyecto>((resolve) => {
        resolver = resolve
      }),
    )
    asociarMock.mockResolvedValue({ id: ana.id, nombre: ana.nombre, email: ana.email })
    const usuario = userEvent.setup()
    renderizar()
    completar()

    await usuario.dblClick(screen.getByRole('button', { name: 'Guardar' }))

    expect(screen.getByRole('button', { name: 'Guardando…' })).toBeDisabled()
    expect(crearMock).toHaveBeenCalledTimes(1)

    resolver(creado)
    expect(await screen.findByText('Detalle del proyecto')).toBeInTheDocument()
    expect(crearMock).toHaveBeenCalledTimes(1)
  })

  it('muestra el error del backend junto al campo indicado en "campo"', async () => {
    crearMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'VALIDACION',
        message: 'el nombre del proyecto ya existe',
        campo: 'nombre',
      }),
    )
    const usuario = userEvent.setup()
    renderizar()
    completar()

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    const campo = await screen.findByLabelText('Nombre')
    await waitFor(() => expect(campo).toHaveAccessibleDescription('el nombre del proyecto ya existe'))
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeEnabled()
  })

  it('conserva los datos del formulario ante un error de red', async () => {
    crearMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'Failed to fetch' }),
    )
    const usuario = userEvent.setup()
    renderizar()
    completar({ nombre: 'Proyecto Alfa', inicio: '2026-01-01', fin: '2026-02-01' })

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('No se pudo conectar')
    expect(screen.getByLabelText('Nombre')).toHaveValue('Proyecto Alfa')
    expect(screen.getByLabelText('Fecha de inicio')).toHaveValue('2026-01-01')
    expect(screen.getByLabelText('Fecha de fin')).toHaveValue('2026-02-01')
  })

  it('si falla la asociación permite reintentarla sin crear otro proyecto', async () => {
    crearMock.mockResolvedValue(creado)
    asociarMock.mockRejectedValueOnce(
      new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'falla' }),
    )
    asociarMock.mockResolvedValueOnce({ id: ana.id, nombre: ana.nombre, email: ana.email })
    const usuario = userEvent.setup()
    renderizar()
    completar()

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('no se pudo asociar')

    await usuario.click(screen.getByRole('button', { name: 'Reintentar asociación' }))

    expect(await screen.findByText('Detalle del proyecto')).toBeInTheDocument()
    expect(crearMock).toHaveBeenCalledTimes(1)
    expect(asociarMock).toHaveBeenCalledTimes(2)
  })

  it('trata como éxito que el usuario ya esté asociado al reintentar', async () => {
    crearMock.mockResolvedValue(creado)
    asociarMock.mockRejectedValueOnce(
      new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'falla' }),
    )
    asociarMock.mockRejectedValueOnce(
      new ApiError({
        status: 409,
        code: 'USUARIO_DUPLICADO',
        message: 'el usuario ya está asociado',
        campo: 'email',
      }),
    )
    const usuario = userEvent.setup()
    renderizar()
    completar()

    await usuario.click(screen.getByRole('button', { name: 'Guardar' }))
    await usuario.click(await screen.findByRole('button', { name: 'Reintentar asociación' }))

    expect(await screen.findByText('Detalle del proyecto')).toBeInTheDocument()
  })
})
