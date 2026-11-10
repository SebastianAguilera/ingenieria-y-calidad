import { NavLink, useMatch } from 'react-router-dom'
import { claseEnlaceDeMenu } from './enlaces'

const SECCIONES_PROYECTO = [
  { etiqueta: 'Resumen', ruta: '', icono: 'bi-grid-1x2' },
  { etiqueta: 'Backlog', ruta: 'backlog', icono: 'bi-list-task' },
  { etiqueta: 'Sprints', ruta: 'sprints', icono: 'bi-lightning-charge' },
  { etiqueta: 'Equipo', ruta: 'equipo', icono: 'bi-person-workspace' },
  { etiqueta: 'Defectos', ruta: 'defectos', icono: 'bi-bug' },
  { etiqueta: 'Métricas', ruta: 'metricas', icono: 'bi-graph-up-arrow' },
]

export default function BarraLateralProyecto() {
  const proyecto = useMatch({ path: '/proyectos/:id', end: false })
  const id = proyecto?.params.id

  if (!id) {
    return null
  }

  return (
    <nav aria-label="Secciones del proyecto" className="mt-6 flex flex-col gap-1">
      {SECCIONES_PROYECTO.map(({ etiqueta, ruta, icono }) => (
        <NavLink
          key={etiqueta}
          to={ruta ? `/proyectos/${id}/${ruta}` : `/proyectos/${id}`}
          end={ruta === ''}
          className={claseEnlaceDeMenu}
        >
          <i className={`bi ${icono}`} aria-hidden="true" />
          {etiqueta}
        </NavLink>
      ))}
    </nav>
  )
}
