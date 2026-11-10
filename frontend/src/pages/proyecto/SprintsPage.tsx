import PageHeader from '../../components/PageHeader'
import Card from '../../ui/Card'

export default function SprintsPage() {
  return (
    <section>
      <PageHeader
        title="Sprints"
        description="Iteraciones del proyecto. Esta sección se implementa en una historia próxima."
      />
      <Card title="Sprints" description="Todavía no hay sprints para mostrar.">
        <p className="text-sm text-ink-soft">
          El plan de sprints y su progreso aparecerán acá.
        </p>
      </Card>
    </section>
  )
}
