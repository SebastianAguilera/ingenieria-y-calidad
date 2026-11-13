export type EstadoProyecto = 'Activo' | 'Cerrado'

export type PrioridadHistoria = 'Alta' | 'Media' | 'Baja'

export type EstadoHistoria = 'Backlog' | 'En Progreso' | 'Hecho'

export type EstadoSprint = 'Planificado' | 'Activo' | 'Cerrado'

export type SeveridadDefecto = 'minor' | 'major' | 'critical' | 'blocker'

export type EstadoDefecto = 'open' | 'in_progress' | 'resolved' | 'closed' | 'rejected'

export type RolProyecto =
  | 'Constructor de Producto'
  | 'Arquitecto de Producto'
  | 'Facilitador Ágil'

export type Usuario = {
  id: number
  nombre: string
  email: string
  creado_en: string
}

export type UsuarioProyecto = {
  id: number
  nombre: string
  email: string
  rol?: RolProyecto
  proyecto_id?: number
  creado_en?: string
}

export type Proyecto = {
  id: number
  nombre: string
  descripcion: string
  fecha_inicio: string
  fecha_fin: string | null
  estado: EstadoProyecto
  cantidad_usuarios: number
  usuarios: UsuarioProyecto[]
  creado_en: string
  actualizado_en: string
}

export type ProyectoListado = {
  id: number
  nombre: string
  descripcion: string
  fecha_inicio: string
  fecha_fin: string | null
  estado: EstadoProyecto
  cantidad_usuarios: number
  creado_en: string
  actualizado_en: string
}

export type Historia = {
  id: number
  codigo: string
  titulo: string
  descripcion: string
  prioridad: PrioridadHistoria
  estado: EstadoHistoria
  story_points: number | null
  esfuerzo_real: number | null
  creado_en: string
  actualizado_en: string
}

export type Sprint = {
  id: number
  identificador: string
  sprint_goal: string
  estado: EstadoSprint
  fecha_inicio: string
  fecha_fin: string
  creado_en: string
  actualizado_en: string
}

export type SprintBacklogItem = {
  id: number
  orden: number
  fecha_asignacion: string
  historia: Historia
}

export type SprintBacklog = {
  sprint: Sprint
  total: number
  items: SprintBacklogItem[]
}

export type Defecto = {
  id: number
  historia_id: number
  sprint_deteccion_id: number
  sprint_resolucion_id: number | null
  reportado_por: number | null
  descripcion: string
  severidad: SeveridadDefecto
  estado: EstadoDefecto
  creado_en: string
  resuelto_en: string | null
}

export type CriterioAceptacion = {
  id: number
  historia_id: number
  descripcion: string
  position: number
  is_met: boolean
}

export type ListadoProyectos = {
  total: number
  proyectos: ProyectoListado[]
}

export type ListadoHistorias = {
  total: number
  historias: Historia[]
}

export type ListadoSprints = {
  total: number
  sprints: Sprint[]
}

export type ListadoDefectos = {
  total: number
  defectos: Defecto[]
}

export type ListadoCriterios = {
  total: number
  criterios: CriterioAceptacion[]
}

export type AltaProyecto = {
  nombre: string
  descripcion?: string
  fecha_inicio: string
  fecha_fin?: string | null
}

export type ActualizacionProyecto = AltaProyecto

export type CambioEstadoProyecto = {
  estado: EstadoProyecto
}

export type AltaUsuario = {
  nombre: string
  email: string
}

export type AltaHistoria = {
  titulo: string
  descripcion: string
  prioridad?: PrioridadHistoria
  estado?: EstadoHistoria
  story_points?: number | null
  esfuerzo_real?: number | null
}

export type ActualizacionHistoria = {
  titulo?: string
  descripcion?: string
  prioridad?: PrioridadHistoria
  estado?: EstadoHistoria
  story_points?: number | null
  esfuerzo_real?: number | null
}

export type AltaSprint = {
  identificador: string
  sprint_goal: string
  fecha_inicio: string
  fecha_fin: string
}

export type CambioEstadoSprint = {
  estado: EstadoSprint
}

export type AltaDefecto = {
  descripcion: string
  severidad: SeveridadDefecto
  historia_id: number
  sprint_deteccion_id: number
  reportado_por?: number | null
}

export type ActualizacionDefecto = {
  estado?: EstadoDefecto
  sprint_resolucion_id?: number | null
}

export type AltaCriterio = {
  descripcion: string
}

export type ActualizacionCriterio = {
  descripcion?: string
  is_met?: boolean
}

export type AsignarHistoriaSprint = {
  historia_id: number
}
