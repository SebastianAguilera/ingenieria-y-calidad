import { describe, expect, it } from 'vitest'

// Todas las tablas comparten `tabla-contenedor` + `tabla` de theme.css: mismo
// encabezado y filas; las que tienen acciones usan `col-acciones`. Esta prueba
// lee los fuentes para detectar una tabla que vuelva a definir su propio estilo.

const FUENTES = import.meta.glob('../**/*.tsx', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

const TABLAS = [
  '../pages/proyecto/backlog/TablaHistorias.tsx',
  '../pages/ProyectosPage.tsx',
  '../pages/proyecto/SprintsPage.tsx',
  '../pages/proyecto/EquipoPage.tsx',
  '../pages/proyecto/DefectosPage.tsx',
]

// ProyectosPage movió sus acciones al detalle del proyecto, así que ya no
// tiene columna de acciones.
const TABLAS_CON_ACCIONES = [
  '../pages/proyecto/SprintsPage.tsx',
  '../pages/proyecto/EquipoPage.tsx',
  '../pages/proyecto/DefectosPage.tsx',
]

function leer(ruta: string): string {
  const fuente = FUENTES[ruta]
  if (fuente === undefined) {
    throw new Error(`No se encontró ${ruta}`)
  }
  return fuente
}

describe('tablas con estilo compartido', () => {
  it.each(TABLAS)('%s usa tabla-contenedor y tabla', (ruta) => {
    const fuente = leer(ruta)

    expect(fuente).toContain('tabla-contenedor')
    expect(fuente).toMatch(/<table\s+className=\{?[`"']tabla[\s`"']/)
  })

  it.each(TABLAS)('%s no define su propio encabezado', (ruta) => {
    const fuente = leer(ruta)

    expect(fuente).not.toContain('px-6 py-4 font-semibold')
    expect(fuente).not.toContain('border-b border-border text-ink-soft')
  })

  it.each(TABLAS_CON_ACCIONES)(
    '%s alinea las acciones con col-acciones y tabla-acciones',
    (ruta) => {
      const fuente = leer(ruta)

      expect(fuente).toContain('col-acciones')
      expect(fuente).toContain('tabla-acciones')
    },
  )

  it('Equipo muestra la tabla sin envolverla en otra tarjeta', () => {
    expect(leer('../pages/proyecto/EquipoPage.tsx')).not.toMatch(/<Card[^>]*title="Integrantes"/s)
  })
})

describe('alineación: el encabezado va como su contenido', () => {
  const CENTRADAS: Array<[string, string]> = [
    ['../pages/ProyectosPage.tsx', 'Nombre'],
    ['../pages/ProyectosPage.tsx', 'Estado'],
    ['../pages/ProyectosPage.tsx', 'Fecha de inicio'],
    ['../pages/ProyectosPage.tsx', 'Fecha de fin'],
    ['../pages/proyecto/SprintsPage.tsx', 'Identificador'],
    ['../pages/proyecto/SprintsPage.tsx', 'Sprint Goal'],
    ['../pages/proyecto/SprintsPage.tsx', 'Fecha de inicio'],
    ['../pages/proyecto/SprintsPage.tsx', 'Fecha de fin'],
    ['../pages/proyecto/SprintsPage.tsx', 'Estado'],
    ['../pages/proyecto/DefectosPage.tsx', 'Código'],
    ['../pages/proyecto/DefectosPage.tsx', 'Descripción'],
    ['../pages/proyecto/DefectosPage.tsx', 'Historia'],
    ['../pages/proyecto/DefectosPage.tsx', 'Severidad'],
    ['../pages/proyecto/DefectosPage.tsx', 'Estado'],
    ['../pages/proyecto/DefectosPage.tsx', 'Sprint de detección'],
    ['../pages/proyecto/DefectosPage.tsx', 'Sprint de resolución'],
    ['../pages/proyecto/EquipoPage.tsx', 'Rol'],
  ]

  it.each(CENTRADAS)('%s centra la columna %s', (ruta, columna) => {
    const encabezado = new RegExp(`<th scope="col" className="text-center">[^<]*${columna}[^<]*</th>`)

    expect(leer(ruta)).toMatch(encabezado)
  })

  it('Equipo centra Rol y Acciones: encabezado, celda y botones', () => {
    const fuente = leer('../pages/proyecto/EquipoPage.tsx')

    expect(fuente).toContain('<td className="text-center">')
    expect(fuente).toContain('<th scope="col" className="col-acciones">')
    expect(fuente).toContain('<td className="col-acciones">')
    expect(fuente).not.toContain('!justify-start')
  })
})
