import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'
import type { Usuario } from '../api/types'
import { SesionProvider } from './SesionContext'
import { useSesion } from './sesion'

const CLAVE_SESION = 'app.sesion.v1'

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

function wrapper({ children }: { children: ReactNode }) {
  return <SesionProvider>{children}</SesionProvider>
}

describe('SesionContext', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('inicia sin usuario activo cuando no hay sesión guardada', () => {
    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
  })

  it('guarda el usuario activo en localStorage al ingresar', () => {
    const { result } = renderHook(() => useSesion(), { wrapper })

    act(() => {
      result.current.ingresar(ana)
    })

    expect(result.current.usuarioActivo).toEqual(ana)
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual(ana)
  })

  it('limpia el usuario activo y el almacenamiento al salir', () => {
    const { result } = renderHook(() => useSesion(), { wrapper })

    act(() => {
      result.current.ingresar(ana)
    })
    act(() => {
      result.current.salir()
    })

    expect(result.current.usuarioActivo).toBeNull()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })

  it('hidrata la sesión desde localStorage al montar', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toEqual(ana)
  })

  it('ignora un JSON inválido en localStorage', () => {
    localStorage.setItem(CLAVE_SESION, '{no-es-json')

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
  })

  it('ignora una sesión guardada con forma inválida', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify({ id: 1, nombre: 'Ana' }))

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
  })

  it('falla si useSesion se usa fuera del provider', () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => renderHook(() => useSesion())).toThrow(
      'useSesion debe usarse dentro de SesionProvider',
    )

    consoleError.mockRestore()
  })
})
