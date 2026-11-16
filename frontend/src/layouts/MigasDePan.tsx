import { Link } from 'react-router-dom'
import type { Sprint } from '../api/types'

type Miga = { etiqueta: string; a?: string }

const SECCIONES: Record<string, string> = {
  backlog: 'Backlog',
  sprints: 'Sprints',
  equipo: 'Equipo',
  defectos: 'Defectos',
  metricas: 'Métricas',
}

/** Migas de la sección actual dentro de un proyecto (sin "Proyectos" ni el nombre). */
function migasDeSeccion(proyectoId: number, resto: string[]): Miga[] {
  const [seccion, , tercero] = resto
  if (seccion === undefined) {
    return [{ etiqueta: 'Información general' }]
  }
  if (seccion === 'editar') {
    return [{ etiqueta: 'Información general', a: `/proyectos/${proyectoId}` }, { etiqueta: 'Editar proyecto' }]
  }
  const etiqueta = SECCIONES[seccion] ?? 'Proyecto'
  if (resto.length === 1) {
    return [{ etiqueta }]
  }
  const padre: Miga = { etiqueta, a: `/proyectos/${proyectoId}/${seccion}` }
  if (seccion === 'backlog' && tercero === 'poker') {
    return [padre, { etiqueta: 'Planning Poker' }]
  }
  if (seccion === 'sprints') {
    return [padre, { etiqueta: 'Detalle de sprint' }]
  }
  return [padre]
}

function migasDeRuta(pathname: string, proyectoId: number | null, nombre: string | null): Miga[] {
  if (proyectoId !== null) {
    const resto = pathname.split('/').filter(Boolean).slice(2)
    return [
      { etiqueta: 'Proyectos', a: '/proyectos' },
      { etiqueta: nombre ?? 'Proyecto', a: `/proyectos/${proyectoId}` },
      ...migasDeSeccion(proyectoId, resto),
    ]
  }
  if (pathname.startsWith('/usuarios')) {
    return [{ etiqueta: 'Usuarios' }]
  }
  if (pathname.startsWith('/proyectos/nuevo')) {
    return [{ etiqueta: 'Proyectos', a: '/proyectos' }, { etiqueta: 'Nuevo proyecto' }]
  }
  return [{ etiqueta: 'Proyectos' }]
}

function formatearCierre(fechaFin: string): string | null {
  const coincide = /^(\d{4})-(\d{2})-(\d{2})/.exec(fechaFin)
  return coincide ? `${coincide[3]}/${coincide[2]}` : null
}

type MigasDePanProps = {
  pathname: string
  proyectoId: number | null
  nombreProyecto: string | null
  sprintActivo: Sprint | null
  /** En pantallas chicas se muestra solo la sección actual y se oculta el sprint. */
  compacto: boolean
}

export default function MigasDePan({
  pathname,
  proyectoId,
  nombreProyecto,
  sprintActivo,
  compacto,
}: MigasDePanProps) {
  const migas = migasDeRuta(pathname, proyectoId, nombreProyecto)
  const visibles = compacto ? migas.slice(-1) : migas
  const cierre = sprintActivo ? formatearCierre(sprintActivo.fecha_fin) : null

  return (
    <div className="flex min-w-0 flex-1 items-center gap-3">
      <nav aria-label="Ruta de navegación" className="min-w-0">
        <ol className="flex min-w-0 items-center gap-2.5 text-lg">
          {visibles.map((miga, indice) => {
            const esUltima = indice === visibles.length - 1
            return (
              <li key={`${miga.etiqueta}-${indice}`} className="flex min-w-0 items-center gap-2">
                {indice > 0 && (
                  <i className="bi bi-chevron-right text-sm text-ink-soft" aria-hidden="true" />
                )}
                {miga.a && !esUltima ? (
                  <Link
                    to={miga.a}
                    className="max-w-[12rem] truncate text-ink-soft transition-colors duration-150 hover:text-ink"
                  >
                    {miga.etiqueta}
                  </Link>
                ) : (
                  <span
                    aria-current={esUltima ? 'page' : undefined}
                    className={`max-w-[14rem] truncate ${esUltima ? 'font-semibold text-ink' : 'text-ink-soft'}`}
                  >
                    {miga.etiqueta}
                  </span>
                )}
              </li>
            )
          })}
        </ol>
      </nav>
      {!compacto && sprintActivo && (
        <span
          className="inline-flex shrink-0 items-center gap-2 rounded-full border border-glass-border bg-glass-2 px-3 py-1 text-xs font-medium text-ink"
          title="Sprint activo"
        >
          <span
            aria-hidden="true"
            className="h-2 w-2 rounded-full bg-success shadow-[0_0_8px_var(--success)]"
          />
          <span>
            {sprintActivo.identificador}
            {cierre && <span className="text-ink-soft"> · cierra {cierre}</span>}
          </span>
        </span>
      )}
    </div>
  )
}
