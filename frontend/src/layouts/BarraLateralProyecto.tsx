import { NavLink, useMatch } from 'react-router-dom'
import { claseEnlaceDeMenu } from './enlaces'

const SECCIONES_PROYECTO = [
  { etiqueta: 'Resumen', ruta: '' },
  { etiqueta: 'Backlog', ruta: 'backlog' },
  { etiqueta: 'Sprints', ruta: 'sprints' },
  { etiqueta: 'Equipo', ruta: 'equipo' },
  { etiqueta: 'Defectos', ruta: 'defectos' },
  { etiqueta: 'Métricas', ruta: 'metricas' },
]

export default function BarraLateralProyecto() {
  const proyecto = useMatch({ path: '/proyectos/:id', end: false })
  const id = proyecto?.params.id

  if (!id) {
    return null
  }

  return (
    <nav aria-label="Secciones del proyecto" className="mt-6 flex flex-col gap-1">
      {SECCIONES_PROYECTO.map(({ etiqueta, ruta }) => (
        <NavLink
          key={etiqueta}
          to={ruta ? `/proyectos/${id}/${ruta}` : `/proyectos/${id}`}
          end={ruta === ''}
          className={claseEnlaceDeMenu}
        >
          {etiqueta}
        </NavLink>
      ))}
    </nav>
  )
}
