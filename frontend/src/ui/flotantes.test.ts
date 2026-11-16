import { describe, expect, it } from 'vitest'

// Todos los modales y elementos flotantes comparten las clases `modal-fondo`,
// `modal-panel` y `flotante` de theme.css: así se ven iguales y los modales
// futuros las heredan. Esta prueba lee los fuentes para detectar un modal que
// vuelva a definir su propio overlay o panel.

const FUENTES = import.meta.glob('../**/*.tsx', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

const MODALES = [
  '../pages/proyecto/backlog/ModalHistoria.tsx',
  '../pages/proyecto/defectos/MarcoModal.tsx',
  '../pages/proyecto/equipo/ModalIntegrante.tsx',
  '../pages/proyecto/poker/ModalAcuerdo.tsx',
  '../pages/proyecto/sprints/ModalAsignarHistoria.tsx',
  '../pages/proyecto/sprints/ModalConfirmacion.tsx',
  '../pages/proyecto/sprints/ModalSprint.tsx',
  '../pages/ModalCrearProyecto.tsx',
  '../pages/proyecto/ModalEditarProyecto.tsx',
  './ModalConfirmacion.tsx',
]

const FLOTANTES = [
  '../layouts/ChipUsuario.tsx',
  '../pages/proyecto/poker/CampanaPoker.tsx',
  '../pages/proyecto/poker/ToastSesionPoker.tsx',
]

function leer(ruta: string): string {
  const fuente = FUENTES[ruta]
  if (fuente === undefined) {
    throw new Error(`No se encontró ${ruta}`)
  }
  return fuente
}

describe('modales con estilo compartido', () => {
  it.each(MODALES)('%s usa modal-fondo y modal-panel', (ruta) => {
    const fuente = leer(ruta)

    expect(fuente).toContain('modal-fondo')
    expect(fuente).toContain('modal-panel')
  })

  it.each(MODALES)('%s no define su propio overlay ni un panel translúcido', (ruta) => {
    const fuente = leer(ruta)

    expect(fuente).not.toContain('fixed inset-0')
    // El panel del diálogo no debe volver a usar el fondo translúcido de superficie.
    expect(fuente).not.toMatch(/role="dialog"[^>]*bg-surface/s)
    expect(fuente).not.toMatch(/role="dialog"[\s\S]{0,300}className="[^"]*bg-surface/)
  })
})

describe('elementos flotantes con estilo compartido', () => {
  it.each(FLOTANTES)('%s usa la clase flotante', (ruta) => {
    expect(leer(ruta)).toContain('flotante')
  })
})
