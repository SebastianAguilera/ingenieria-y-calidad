import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { login } from '../api/auth'
import { ApiError } from '../api/errors'
import type { Usuario } from '../api/types'
import { CLAVE_SESION } from '../session/sesion'
import { SesionProvider } from '../session/SesionContext'
import { ThemeProvider } from '../theme/ThemeProvider'
import LoginPage from './LoginPage'

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}))

const loginMock = vi.mocked(login)

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const SESION = { token: 'jwt-vigente', usuario: ana }

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

function escribir(email: string, password: string) {
  const usuario = userEvent.setup()
  return usuario.type(screen.getByLabelText('Email'), email).then(() =>
    password === '' ? Promise.resolve() : usuario.type(screen.getByLabelText('Contraseña'), password),
  )
}

function enviarIngreso() {
  return userEvent.setup().click(screen.getByRole('button', { name: /Ingresar/ }))
}

describe('LoginPage (US-F08 - login real con JWT)', () => {
  beforeEach(() => {
    localStorage.clear()
    loginMock.mockReset()
  })

  it('E-03: con credenciales válidas guarda token y usuario y redirige al panel', async () => {
    loginMock.mockResolvedValue(SESION)

    renderizarLogin()
    await escribir('ana@example.com', 'Secreto123')
    await enviarIngreso()

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
    expect(loginMock).toHaveBeenCalledWith('ana@example.com', 'Secreto123')
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual(SESION)
  })

  it('E-04: con credenciales inválidas muestra el mensaje y no guarda token', async () => {
    loginMock.mockRejectedValue(
      new ApiError({
        status: 401,
        code: 'CREDENCIALES_INVALIDAS',
        message: 'email o contraseña incorrectos',
      }),
    )

    renderizarLogin()
    await escribir('ana@example.com', 'mala')
    await enviarIngreso()

    expect(await screen.findByText('Email o contraseña incorrectos')).toBeInTheDocument()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
    expect(screen.queryByText('Panel de proyectos')).not.toBeInTheDocument()
  })

  it('E-06: sin email muestra la validación y no consulta la API', async () => {
    renderizarLogin()

    const usuario = userEvent.setup()
    await usuario.type(screen.getByLabelText('Contraseña'), 'Secreto123')
    await enviarIngreso()

    expect(screen.getByText('Ingresá tu email')).toBeInTheDocument()
    expect(loginMock).not.toHaveBeenCalled()
  })

  it('E-06: sin contraseña muestra la validación y no consulta la API', async () => {
    renderizarLogin()

    const usuario = userEvent.setup()
    await usuario.type(screen.getByLabelText('Email'), 'ana@example.com')
    await enviarIngreso()

    expect(screen.getByText('Ingresá tu contraseña')).toBeInTheDocument()
    expect(loginMock).not.toHaveBeenCalled()
  })

  it('E-07: ante un 422 del backend muestra su mensaje y resalta el campo', async () => {
    loginMock.mockRejectedValue(
      new ApiError({
        status: 422,
        code: 'VALIDACION',
        message: 'el email no tiene un formato válido',
        campo: 'email',
      }),
    )

    renderizarLogin()
    await escribir('no-es-email', 'Secreto123')
    await enviarIngreso()

    expect(await screen.findByText('el email no tiene un formato válido')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toHaveAttribute('aria-invalid', 'true')
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })

  it('E-08: ante un error de red muestra el mensaje y permite reintentar', async () => {
    loginMock.mockRejectedValue(
      new ApiError({ status: 0, code: 'NETWORK_ERROR', message: 'sin conexion' }),
    )

    renderizarLogin()
    await escribir('ana@example.com', 'Secreto123')
    await enviarIngreso()

    expect(await screen.findByText('No se pudo conectar con el servidor.')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toHaveValue('ana@example.com')
    expect(screen.getByLabelText('Contraseña')).toHaveValue('Secreto123')

    loginMock.mockResolvedValue(SESION)
    await enviarIngreso()

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
  })

  it('E-09: con una sesión vigente redirige al panel', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(SESION))

    renderizarLogin()

    expect(screen.getByText('Panel de proyectos')).toBeInTheDocument()
  })

  it('E-13: ignora una sesión con el formato anterior de US-F02', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))

    renderizarLogin()

    expect(screen.getByText('Ingresá al panel')).toBeInTheDocument()
    expect(screen.queryByText('Panel de proyectos')).not.toBeInTheDocument()
  })

  it('E-14: recorta el email antes de enviarlo', async () => {
    loginMock.mockResolvedValue(SESION)

    renderizarLogin()
    await escribir('  ANA@Example.com  ', 'Secreto123')
    await enviarIngreso()

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
    expect(loginMock).toHaveBeenCalledWith('ANA@Example.com', 'Secreto123')
  })

  it('E-15: un doble envío genera una sola petición', async () => {
    loginMock.mockReturnValue(new Promise(() => {}))

    renderizarLogin()
    await escribir('ana@example.com', 'Secreto123')

    const boton = screen.getByRole('button', { name: /Ingresar/ })
    await userEvent.setup().click(boton)
    await userEvent.setup().click(boton)

    expect(loginMock).toHaveBeenCalledTimes(1)
    expect(boton).toBeDisabled()
  })

  it('E-16: la contraseña no se guarda en el almacenamiento', async () => {
    loginMock.mockResolvedValue(SESION)

    renderizarLogin()
    await escribir('ana@example.com', 'Secreto123')
    await enviarIngreso()

    expect(await screen.findByText('Panel de proyectos')).toBeInTheDocument()
    const guardado = localStorage.getItem(CLAVE_SESION) ?? ''
    expect(guardado).not.toContain('Secreto123')
    expect(guardado).not.toContain('password')
  })

  it('ante un error inesperado muestra el mensaje genérico', async () => {
    loginMock.mockRejectedValue(
      new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'boom' }),
    )

    renderizarLogin()
    await escribir('ana@example.com', 'Secreto123')
    await enviarIngreso()

    expect(await screen.findByText('boom')).toBeInTheDocument()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })
})
