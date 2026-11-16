import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import Card from '../../../ui/Card'
import TablaHistorias from './TablaHistorias'
import type { ColumnaHistoria } from './TablaHistorias'
import { claseEstado, clasePrioridad } from './etiquetas'
import { useBacklogPriorizado } from './useBacklogPriorizado'

type VistaBacklogPriorizadoProps = {
  proyectoId: number
}

function formatear(valor: number | null): string {
  return valor === null ? '—' : String(valor)
}

export default function VistaBacklogPriorizado({ proyectoId }: VistaBacklogPriorizadoProps) {
  const { estado, reintentar } = useBacklogPriorizado(proyectoId)

  if (estado.tipo === 'cargando') {
    return <p className="text-sm text-ink-soft" role="status">Cargando backlog priorizado…</p>
  }

  if (estado.tipo === 'error') {
    return (
      <div className="flex flex-col items-start gap-4">
        <Alert tono="error">{estado.mensaje}</Alert>
        <Button variante="secondary" onClick={reintentar}>
          Reintentar
        </Button>
      </div>
    )
  }

  const historias = estado.historias.filter((historia) => historia.estado !== 'Hecho')

  if (historias.length === 0) {
    return (
      <Card
        title="Backlog priorizado"
        description="No hay historias pendientes de priorizar"
      >
        <p className="text-sm text-ink-soft">
          Las historias marcadas como Hecho no se incluyen en esta vista.
        </p>
      </Card>
    )
  }

  const priorizadasPorId = new Map(estado.historias.map((historia) => [historia.id, historia]))

  const columnas: ColumnaHistoria[] = [
    {
      clave: 'codigo',
      encabezado: 'Código',
      render: (historia) => (
        <span className="font-mono text-xs text-ink-soft">{historia.codigo}</span>
      ),
    },
    {
      clave: 'titulo',
      encabezado: 'Título',
      alineacion: 'centro',
      render: (historia) => <span className="font-medium text-ink">{historia.titulo}</span>,
    },
    {
      clave: 'estado',
      encabezado: 'Estado',
      render: (historia) => (
        <span
          className={`inline-flex rounded-pill border px-2.5 py-0.5 text-xs font-semibold ${claseEstado(historia.estado)}`}
        >
          {historia.estado}
        </span>
      ),
    },
    {
      clave: 'prioridad',
      encabezado: 'Prioridad',
      alineacion: 'centro',
      render: (historia) => (
        <span
          className={`inline-flex rounded-pill border px-2.5 py-0.5 text-xs font-semibold ${clasePrioridad(historia.prioridad)}`}
        >
          {historia.prioridad}
        </span>
      ),
    },
    {
      clave: 'story_points',
      encabezado: 'SP',
      alineacion: 'centro',
      render: (historia) => formatear(historia.story_points),
    },
    {
      clave: 'valor_negocio',
      encabezado: 'Valor',
      alineacion: 'centro',
      render: (historia) => formatear(historia.valor_negocio),
    },
    {
      clave: 'criticidad_tiempo',
      encabezado: 'Criticidad',
      alineacion: 'centro',
      render: (historia) => formatear(historia.criticidad_tiempo),
    },
    {
      clave: 'reduccion_riesgo',
      encabezado: 'Riesgo',
      alineacion: 'centro',
      render: (historia) => formatear(historia.reduccion_riesgo),
    },
    {
      clave: 'roi',
      encabezado: 'ROI',
      alineacion: 'centro',
      render: (historia) => formatear(priorizadasPorId.get(historia.id)?.roi ?? null),
    },
    {
      clave: 'wsjf',
      encabezado: 'WSJF',
      alineacion: 'centro',
      render: (historia) => formatear(priorizadasPorId.get(historia.id)?.wsjf ?? null),
    },
  ]

  return <TablaHistorias historias={historias} columnas={columnas} />
}
