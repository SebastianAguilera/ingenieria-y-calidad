import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'
import type { Usuario } from '../api/types'
import { CLAVE_SESION, useSesion } from '../session/sesion'
import { SesionProvider } from '../session/SesionContext'
import RutaProtegida from './RutaProtegida'

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const SESION = { token: 'jwt-vigente', usuario: ana }

function PanelConSalir() {
  const { salir } = useSesion()
  const navegar = useNavigate()

  return (
    <div>
      <p>Panel de proyectos</p>
      <button
        type="button"
        onClick={() => {
          salir()
          navegar('/login')
        }}
      >
        Salir
      </button>
    </div>
  )
}

function renderizarRutas(inicial = '/proyectos') {
  return render(
    <SesionProvider>
      <MemoryRouter initialEntries={[inicial]}>
        <Routes>
          <Route path="/login" element={<p>Pantalla de login</p>} />
          <Route element={<RutaProtegida />}>
            <Route path="/proyectos" element={<PanelConSalir />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </SesionProvider>,
  )
}

describe('RutaProtegida', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('redirige a /login cuando no hay token (E-01)', () => {
    renderizarRutas()

    expect(screen.getByText('Pantalla de login')).toBeInTheDocument()
    expect(screen.queryByText('Panel de proyectos')).not.toBeInTheDocument()
  })

  it('redirige a /login cuando hay usuario pero sin token (E-12)', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))

    renderizarRutas()

    expect(screen.getByText('Pantalla de login')).toBeInTheDocument()
    expect(screen.queryByText('Panel de proyectos')).not.toBeInTheDocument()
  })

  it('muestra el contenido protegido cuando hay token (E-02)', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(SESION))

    renderizarRutas()

    expect(screen.getByText('Panel de proyectos')).toBeInTheDocument()
  })

  it('redirige a /login al salir de la sesión (E-10)', async () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(SESION))
    const usuario = userEvent.setup()

    renderizarRutas()
    await usuario.click(screen.getByRole('button', { name: 'Salir' }))

    expect(screen.getByText('Pantalla de login')).toBeInTheDocument()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })

  it('redirige a /login cuando el token vence en un request (E-11)', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(SESION))
    renderizarRutas()

    expect(screen.getByText('Panel de proyectos')).toBeInTheDocument()

    localStorage.removeItem(CLAVE_SESION)
    act(() => {
      window.dispatchEvent(new Event('sesion:expirada'))
    })

    expect(screen.getByText('Pantalla de login')).toBeInTheDocument()
  })
})
