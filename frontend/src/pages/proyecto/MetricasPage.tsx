import { Link, useParams } from 'react-router-dom'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import GraficoEstados from './metricas/GraficoEstados'
import GraficoSprints from './metricas/GraficoSprints'
import PanelSprint from './metricas/PanelSprint'
import ResumenKpi from './metricas/ResumenKpi'
import SeccionFlujo from './metricas/SeccionFlujo'
import { useMetricas } from './metricas/useMetricas'

function parsearProyectoId(id: string | undefined): number | null {
  if (id === undefined) {
    return null
  }
  const numero = Number(id)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

export default function MetricasPage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const { estado, reintentar } = useMetricas(proyectoId)

  return (
    <section className="mx-auto w-full max-w-6xl">
      <PageHeader
        title="Métricas"
        description="Avance del proyecto y de cada sprint: velocidad, completitud, esfuerzo y defectos."
        className="mb-6"
      />

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando métricas…
        </p>
      )}

      {estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{estado.mensaje}</Alert>
          <Button variante="secondary" onClick={reintentar}>
            Reintentar
          </Button>
        </div>
      )}

      {estado.tipo === 'listo' && estado.datos.historias.length === 0 && (
        <Card
          title="Todavía no hay historias"
          description="Las métricas se calculan a partir de las historias del proyecto. Creá la primera en el backlog para empezar a medir."
        >
          <Link
            to={`/proyectos/${proyectoId}/backlog`}
            className="inline-flex items-center gap-2 rounded-pill bg-primary px-5 py-2.5 text-sm font-semibold text-[var(--on-primary)] transition hover:bg-primary-700 focus-visible:shadow-[var(--ring)]"
          >
            <i className="bi bi-list-task" aria-hidden="true" />
            Ir al backlog
          </Link>
        </Card>
      )}

      {estado.tipo === 'listo' && proyectoId !== null && estado.datos.historias.length > 0 && (
        <div className="flex flex-col gap-5">
          <Alert tono="info">
            <strong>Calculado sobre el estado actual de las historias</strong>
            <span className="block text-ink-soft">
              El sistema no guarda histórico: si una historia cambia después del cierre de un sprint, sus
              métricas también cambian.
            </span>
          </Alert>

          <ResumenKpi metricas={estado.datos.metricas} />

          <div className="grid gap-5 lg:grid-cols-12">
            <GraficoSprints sprints={estado.datos.sprintsGrafico} className="lg:col-span-7" />
            <GraficoEstados historias={estado.datos.historias} className="lg:col-span-5" />
          </div>

          <SeccionFlujo metricas={estado.datos.metricas} />

          <PanelSprint proyectoId={proyectoId} sprints={estado.datos.sprints} />
        </div>
      )}
    </section>
  )
}
