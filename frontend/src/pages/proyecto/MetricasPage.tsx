import PageHeader from '../../components/PageHeader'
import Card from '../../ui/Card'

export default function MetricasPage() {
  return (
    <section>
      <PageHeader
        title="Métricas"
        description="Indicadores de calidad del proyecto. Esta sección se implementa en una historia próxima."
      />
      <Card title="Métricas" description="Todavía no hay métricas para mostrar.">
        <p className="text-sm text-ink-soft">
          Velocidad, desviaciones y ratios de defectos aparecerán acá.
        </p>
      </Card>
    </section>
  )
}
