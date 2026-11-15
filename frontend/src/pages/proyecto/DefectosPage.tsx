import PageHeader from '../../components/PageHeader'
import Card from '../../ui/Card'

export default function DefectosPage() {
  return (
    <section>
      <PageHeader
        title="Defectos"
        description="Bugs del proyecto. Esta sección se implementa en una historia próxima."
      />
      <Card title="Defectos" description="Todavía no hay defectos para mostrar.">
        <p className="text-sm text-ink-soft">
          El seguimiento de bugs del sprint aparecerá acá.
        </p>
      </Card>
    </section>
  )
}
