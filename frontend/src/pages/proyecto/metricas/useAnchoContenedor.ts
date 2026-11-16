import { useEffect, useRef, useState } from 'react'

/**
 * Mide el ancho disponible para un gráfico. Reemplaza a `ResponsiveContainer`
 * de Recharts: sin `ResizeObserver` (jsdom) el ancho queda en 0 y el gráfico no
 * se dibuja; la tabla accesible sigue mostrando los mismos valores.
 */
export function useAnchoContenedor<T extends HTMLElement>() {
  const ref = useRef<T>(null)
  const [ancho, setAncho] = useState(0)

  useEffect(() => {
    const elemento = ref.current
    if (!elemento || typeof ResizeObserver === 'undefined') {
      return
    }

    const observador = new ResizeObserver(([entrada]) => {
      setAncho(Math.floor(entrada.contentRect.width))
    })
    observador.observe(elemento)
    return () => observador.disconnect()
  }, [])

  return { ref, ancho }
}
