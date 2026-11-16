import { render } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import GraficoSeveridad from './GraficoSeveridad'

// jsdom no tiene ResizeObserver: se simula un ancho para que se dibuje el SVG.
class ResizeObserverFalso {
  private readonly alCambiar: (entradas: Array<{ contentRect: { width: number } }>) => void
  constructor(alCambiar: (entradas: Array<{ contentRect: { width: number } }>) => void) {
    this.alCambiar = alCambiar
  }
  observe() {
    this.alCambiar([{ contentRect: { width: 600 } }])
  }
  disconnect() {}
}

describe('GraficoSeveridad - degradé de las barras', () => {
  beforeEach(() => {
    vi.stubGlobal('ResizeObserver', ResizeObserverFalso)
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('define un degradé por severidad y lo usa como relleno de cada barra', () => {
    const { container } = render(
      <GraficoSeveridad defectos={{ minor: 1, major: 2, critical: 3, blocker: 4 }} />,
    )

    const degrades = container.querySelectorAll('linearGradient')
    expect(degrades).toHaveLength(4)

    const ids = [...degrades].map((degrade) => degrade.getAttribute('id'))
    const rellenos = [...container.querySelectorAll('.recharts-bar-rectangle path')].map((barra) =>
      barra.getAttribute('fill'),
    )
    expect(rellenos).toHaveLength(4)
    for (const [indice, relleno] of rellenos.entries()) {
      expect(relleno).toBe(`url(#${ids[indice]})`)
    }
  })

  it('cada degradé va del color a una versión más transparente del mismo color', () => {
    const { container } = render(
      <GraficoSeveridad defectos={{ minor: 1, major: 2, critical: 3, blocker: 4 }} />,
    )

    expect(container.querySelectorAll('linearGradient')).toHaveLength(4)
    for (const degrade of container.querySelectorAll('linearGradient')) {
      const paradas = degrade.querySelectorAll('stop')
      expect(paradas).toHaveLength(2)
      const opacidades = [...paradas].map((parada) => Number(parada.getAttribute('stop-opacity')))
      expect(opacidades[0]).toBeLessThan(opacidades[1])
    }
  })
})
