import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'
import type { Usuario } from '../api/types'
import { CLAVE_SESION } from '../session/sesion'
import { SesionProvider } from '../session/SesionContext'
import ProyectosPage from './ProyectosPage'

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

describe('ProyectosPage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('saluda al usuario activo', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))

    render(
      <SesionProvider>
        <ProyectosPage />
      </SesionProvider>,
    )

    expect(screen.getByRole('heading', { name: 'Hola, Ana Pérez' })).toBeInTheDocument()
  })
})
