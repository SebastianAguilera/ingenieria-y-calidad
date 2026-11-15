import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'
import type { Usuario } from '../api/types'
import { SesionProvider } from './SesionContext'
import { CLAVE_SESION, EVENTO_SESION_EXPIRADA, useSesion } from './sesion'

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const SESION = { token: 'jwt-vigente', usuario: ana }

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

  it('inicia sin sesión cuando no hay nada guardado', () => {
    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
    expect(result.current.token).toBeNull()
  })

  it('guarda token y usuario en localStorage al ingresar', () => {
    const { result } = renderHook(() => useSesion(), { wrapper })

    act(() => {
      result.current.ingresar('jwt-vigente', ana)
    })

    expect(result.current.usuarioActivo).toEqual(ana)
    expect(result.current.token).toBe('jwt-vigente')
    expect(JSON.parse(localStorage.getItem(CLAVE_SESION) ?? 'null')).toEqual(SESION)
  })

  it('limpia token, usuario y almacenamiento al salir', () => {
    const { result } = renderHook(() => useSesion(), { wrapper })

    act(() => {
      result.current.ingresar('jwt-vigente', ana)
    })
    act(() => {
      result.current.salir()
    })

    expect(result.current.usuarioActivo).toBeNull()
    expect(result.current.token).toBeNull()
    expect(localStorage.getItem(CLAVE_SESION)).toBeNull()
  })

  it('hidrata token y usuario desde localStorage al montar', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(SESION))

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toEqual(ana)
    expect(result.current.token).toBe('jwt-vigente')
  })

  it('ignora un JSON inválido en localStorage', () => {
    localStorage.setItem(CLAVE_SESION, '{no-es-json')

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
    expect(result.current.token).toBeNull()
  })

  it('ignora una sesión con la forma anterior de US-F02 (sin token)', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(ana))

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
    expect(result.current.token).toBeNull()
  })

  it('ignora una sesión guardada con forma inválida', () => {
    localStorage.setItem(
      CLAVE_SESION,
      JSON.stringify({ token: '', usuario: { id: 1, nombre: 'Ana' } }),
    )

    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.usuarioActivo).toBeNull()
    expect(result.current.token).toBeNull()
  })

  it('limpia el estado cuando llega el evento de sesión expirada', () => {
    localStorage.setItem(CLAVE_SESION, JSON.stringify(SESION))
    const { result } = renderHook(() => useSesion(), { wrapper })

    expect(result.current.token).toBe('jwt-vigente')

    act(() => {
      window.dispatchEvent(new Event(EVENTO_SESION_EXPIRADA))
    })

    expect(result.current.usuarioActivo).toBeNull()
    expect(result.current.token).toBeNull()
  })

  it('falla si useSesion se usa fuera del provider', () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => renderHook(() => useSesion())).toThrow(
      'useSesion debe usarse dentro de SesionProvider',
    )

    consoleError.mockRestore()
  })
})
