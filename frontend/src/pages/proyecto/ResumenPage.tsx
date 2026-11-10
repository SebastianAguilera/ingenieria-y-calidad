import PageHeader from '../../components/PageHeader'
import Card from '../../ui/Card'

export default function ResumenPage() {
  return (
    <section>
      <PageHeader
        title="Resumen"
        description="Vista general del proyecto. Esta sección se implementa en una historia próxima."
      />
      <Card title="Resumen" description="Todavía no hay datos que mostrar.">
        <p className="text-sm text-ink-soft">
          Acá se verán el estado del proyecto y sus indicadores principales.
        </p>
      </Card>
    </section>
  )
}
