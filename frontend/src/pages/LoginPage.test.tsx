import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../api/errors'
import type { Usuario } from '../api/types'
import { buscarUsuario, crearUsuario } from '../api/usuarios'
import { CLAVE_SESION } from '../session/sesion'
import { SesionProvider } from '../session/SesionContext'
import { ThemeProvider } from '../theme/ThemeProvider'
import LoginPage from './LoginPage'

vi.mock('../api/usuarios', () => ({
  buscarUsuario: vi.fn(),
  crearUsuario: vi.fn(),
}))

const buscarMock = vi.mocked(buscarUsuario)
const crearMock = vi.mocked(crearUsuario)

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const otroBruno: Usuario = {
  id: 2,
  nombre: 'Bruno Díaz',
  email: 'bruno@example.com',
  creado_en: '2026-01-02T00:00:00Z',
}

function renderizarLogin(inicial = '/login') {
  return render(
    <ThemeProvider>
      <SesionProvider>
        <MemoryRouter initialEntries={[inicial]}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/proyectos" element={<p>Panel de proyectos</p>} />
          </Routes>
        </MemoryRouter>
      </SesionProvider>
    </ThemeProvider>,
  )
}

function ingresar(texto: string) {
  return userEvent.setup().type(screen.getByLabelText('Email o nombre'), texto)
}

async function enviarIngreso() {
  await userEvent.setup().click(screen.getByRole('button', { name: 'Ingresar' }))
}

describe('LoginPage', () => {
  beforeEach(() => {
    localStorage.clear()
    buscarMock.mockReset()
    crearMock.mockReset()
  })

  it('E-06: con el campo vacío muestra el mensaje y no consulta la API', async () => {
    renderizarLogin()

    await enviarIngreso()

    expect(screen.getByText('Ingresá tu email')).toBeInTheDocument()
    expect(buscarMock).not.toHaveBeenCalled()
  })

  it('E-03: ingresa con un email existente y redirige al panel', async () => {
    buscarMock.mockResolvedValue({ estado: 'encontrado', usuario: ana })

    renderizarLogin()
    await ingresar('ana@example.com')
    await enviarIngreso()

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
    expect(buscarMock).toHaveBeenCalledWith('ana@example.com')
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual(ana)
  })

  it('E-04: ingresa con un nombre que identifica a un único usuario', async () => {
    buscarMock.mockResolvedValue({ estado: 'encontrado', usuario: ana })

    renderizarLogin()
    await ingresar('Ana Pérez')
    await enviarIngreso()

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
    expect(buscarMock).toHaveBeenCalledWith('Ana Pérez')
  })

  it('E-07: con un email inexistente muestra la tarjeta de registro', async () => {
    buscarMock.mockResolvedValue({ estado: 'no_encontrado' })

    renderizarLogin()
    await ingresar('nuevo@example.com')
    await enviarIngreso()

    expect(await screen.findByText('Usuario no encontrado')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toHaveValue('nuevo@example.com')
  })

  it('E-08: registra un usuario nuevo y redirige al panel', async () => {
    const creado: Usuario = {
      id: 9,
      nombre: 'Nuevo Usuario',
      email: 'nuevo@example.com',
      creado_en: '2026-01-03T00:00:00Z',
    }
    buscarMock.mockResolvedValue({ estado: 'no_encontrado' })
    crearMock.mockResolvedValue(creado)

    renderizarLogin()
    await ingresar('nuevo@example.com')
    await enviarIngreso()

    const usuario = userEvent.setup()
    await usuario.type(screen.getByLabelText('Nombre'), 'Nuevo Usuario')
    await usuario.click(screen.getByRole('button', { name: 'Crear usuario' }))

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
    expect(crearMock).toHaveBeenCalledWith({
      nombre: 'Nuevo Usuario',
      email: 'nuevo@example.com',
    })
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual(creado)
  })

  it('E-09: muestra un mensaje cuando el email ya está registrado', async () => {
    buscarMock.mockResolvedValue({ estado: 'no_encontrado' })
    crearMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'USUARIO_DUPLICADO',
        message: 'el email ya existe',
        campo: 'email',
      }),
    )

    renderizarLogin()
    await ingresar('ana@example.com')
    await enviarIngreso()

    const usuario = userEvent.setup()
    await usuario.type(screen.getByLabelText('Nombre'), 'Otra Ana')
    await usuario.click(screen.getByRole('button', { name: 'Crear usuario' }))

    expect(await screen.findByText('Ese email ya está registrado. Volvé a ingresar.')).toBeInTheDocument()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })

  it('E-10: no registra si faltan datos obligatorios', async () => {
    buscarMock.mockResolvedValue({ estado: 'no_encontrado' })

    renderizarLogin()
    await ingresar('nuevo@example.com')
    await enviarIngreso()

    await userEvent.setup().click(screen.getByRole('button', { name: 'Crear usuario' }))

    expect(screen.getByText('Completá tu nombre y tu email')).toBeInTheDocument()
    expect(crearMock).not.toHaveBeenCalled()
  })

  it('E-11: muestra un mensaje ante un error de red', async () => {
    buscarMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexion' }),
    )

    renderizarLogin()
    await ingresar('ana@example.com')
    await enviarIngreso()

    expect(await screen.findByText('No se pudo conectar con el servidor.')).toBeInTheDocument()
  })

  it('E-12: pide el email cuando el nombre es ambiguo', async () => {
    buscarMock.mockResolvedValue({
      estado: 'ambiguo',
      usuarios: [ana, otroBruno],
    })

    renderizarLogin()
    await ingresar('Ana Pérez')
    await enviarIngreso()

    expect(
      await screen.findByText('Hay más de un usuario con ese nombre. Ingresá tu email.'),
    ).toBeInTheDocument()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })

  it('redirige al panel si ya hay una sesión activa', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))

    renderizarLogin()

    expect(screen.getByText('Panel de proyectos')).toBeInTheDocument()
  })
})
