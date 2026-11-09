import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { CLAVE_TEMA, TemaContext, leerModoGuardado, resolverModo } from './tema'
import type { Modo, ModoResuelto } from './tema'

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [modo, setModo] = useState<Modo>(() => leerModoGuardado())
  const [esOscuro, setEsOscuro] = useState<boolean>(
    () => resolverModo(leerModoGuardado()) === 'dark',
  )

  useEffect(() => {
    const aplicar = () => {
      const resuelto: ModoResuelto = resolverModo(modo)
      document.documentElement.dataset.mode = resuelto
      setEsOscuro(resuelto === 'dark')
    }

    aplicar()
    try {
      localStorage.setItem(CLAVE_TEMA, JSON.stringify(modo))
    } catch {
      // Sin almacenamiento disponible el tema sigue funcionando en memoria.
    }

    if (modo !== 'system') {
      return
    }

    const consulta = window.matchMedia('(prefers-color-scheme: dark)')
    consulta.addEventListener('change', aplicar)
    return () => consulta.removeEventListener('change', aplicar)
  }, [modo])

  const alternarModo = useCallback(() => {
    setModo(resolverModo(modo) === 'dark' ? 'light' : 'dark')
  }, [modo])

  const valor = useMemo(
    () => ({ modo, esOscuro, setModo, alternarModo }),
    [modo, esOscuro, alternarModo],
  )

  return <TemaContext.Provider value={valor}>{children}</TemaContext.Provider>
}
