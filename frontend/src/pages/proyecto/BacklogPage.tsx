import PageHeader from '../../components/PageHeader'
import Card from '../../ui/Card'

export default function BacklogPage() {
  return (
    <section>
      <PageHeader
        title="Backlog"
        description="Historias de usuario del proyecto. Esta sección se implementa en una historia próxima."
      />
      <Card title="Backlog" description="Todavía no hay historias para mostrar.">
        <p className="text-sm text-ink-soft">
          El listado de historias con sus story points aparecerá acá.
        </p>
      </Card>
    </section>
  )
}
