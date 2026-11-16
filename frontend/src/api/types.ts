export type EstadoProyecto = 'Activo' | 'Cerrado'

export type PrioridadHistoria = 'Alta' | 'Media' | 'Baja'

export type EstadoHistoria = 'Backlog' | 'En Progreso' | 'Hecho'

export type EstadoSprint = 'Planificado' | 'Activo' | 'Cerrado'

export type SeveridadDefecto = 'minor' | 'major' | 'critical' | 'blocker'

export type EstadoDefecto = 'open' | 'in_progress' | 'resolved' | 'closed' | 'rejected'

export type TipoListadoDefecto = 'deteccion' | 'resolucion'

export type RolProyecto = 'LIDER' | 'MIEMBRO'

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
  rol_actual?: RolProyecto
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
  horas_estimadas: number | null
  creado_en: string
  actualizado_en: string
}

export type EstadoSesionPoker = 'Votando' | 'Revelada' | 'Cerrada' | 'Cancelada'

export type ReferenciaUsuario = {
  id: number
  nombre: string
}

export type StoryEstimate = {
  id: number
  historia_id: number
  sesion_id: number
  ronda_id: number
  puntos: number
  fuente: string
  estimado_por: number | null
  estimado_en: string
}

export type SesionPokerHistorial = {
  id: number
  estado: EstadoSesionPoker
  ronda_actual: number
  puntos_acordados: number | null
  creado_por: ReferenciaUsuario
  creado_en: string
  cerrado_en?: string | null
  cancelado_en?: string | null
  story_estimate?: StoryEstimate | null
}

export type SesionPokerCreada = {
  id: number
  historia_id: number
  estado: EstadoSesionPoker
  ronda_actual: number
  creado_por: ReferenciaUsuario
  creado_en: string
  cerrado_en?: string | null
  puntos_acordados: number | null
}

export type SesionPokerActiva = {
  sesion_id: number
  estado: EstadoSesionPoker
  ronda_actual: number
  creado_por: ReferenciaUsuario
  creado_en: string
  votos_registrados: number
  total_miembros: number
  ya_vote: boolean
  historia: {
    id: number
    codigo: string
    titulo: string
  }
}

export type VotoPokerRevelado = {
  usuario_id: number
  usuario_nombre: string
  puntos: number
  ronda: number
}

export type DispersionPoker = {
  promedio: number
  mediana: number
  minimo: number
  maximo: number
  rango: number
  desviacion_estandar: number
  hay_dispersion: boolean
  total_votos: number
  votos_por_valor: Record<string, number>
}

export type HistorialRondaPoker = {
  numero: number
  estado: EstadoSesionPoker
  revelada_en: string | null
  votos: VotoPokerRevelado[]
  dispersion?: DispersionPoker
}

export type DetalleSesionPoker = {
  id: number
  historia_id: number
  estado: EstadoSesionPoker
  ronda_actual: number
  creado_por: ReferenciaUsuario
  creado_en: string
  cerrado_en: string | null
  cancelado_en: string | null
  puntos_acordados: number | null
  story_estimate?: StoryEstimate | null
  total_miembros: number
  votos_registrados: number
  ya_vote: boolean
  mi_voto: number | null
  votos: VotoPokerRevelado[]
  dispersion?: DispersionPoker
  historial_rondas: HistorialRondaPoker[]
}

export type ResultadoVotoPoker = {
  sesion_id: number
  mi_voto: number
  votado_en: string
  votos_registrados: number
  total_miembros: number
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
  horas_estimadas?: number | null
}

export type ActualizacionHistoria = {
  titulo?: string
  descripcion?: string
  prioridad?: PrioridadHistoria
  estado?: EstadoHistoria
  story_points?: number | null
  esfuerzo_real?: number | null
  horas_estimadas?: number | null
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
  descripcion?: string
  severidad?: SeveridadDefecto
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

/** Base de cálculo de las métricas: siempre el estado actual (US-20, RN-19). */
export type BaseMetricas = 'estado_actual'

export type DefectosPorSeveridad = Record<SeveridadDefecto, number>

/** Respuesta de `GET /api/proyectos/{id}/metricas` (US-20 y extensiones de US-21). */
export type MetricasProyecto = {
  proyecto_id: number
  base: BaseMetricas
  velocidad: number | null
  porcentaje_historias_completadas: number | null
  horas_estimadas: number
  horas_reales: number
  desviacion_esfuerzo: number | null
  defectos_detectados: number
  defectos_resueltos: number
  ratio_defectos: number | null
  sprints_restantes: number | null
  wip: number
  tasa_entrega: number | null
  lead_time_estimado: number | null
  lead_time_aproximado: boolean
  defectos_abiertos: number
  defectos_por_severidad: DefectosPorSeveridad
}

/** Respuesta de `GET /api/proyectos/{id}/sprints/{sprintId}/metricas` (US-20). */
export type MetricasSprint = {
  sprint_id: number
  base: BaseMetricas
  story_points_planificados: number
  story_points_completados: number
  porcentaje_completado: number | null
  historias_completadas: number
  horas_estimadas: number
  horas_reales: number
  desviacion_esfuerzo: number | null
  defectos_detectados: number
  defectos_resueltos: number
}
