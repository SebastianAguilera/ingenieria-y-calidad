import PageHeader from '../../components/PageHeader'
import Card from '../../ui/Card'

export default function EquipoPage() {
  return (
    <section>
      <PageHeader
        title="Equipo"
        description="Miembros del proyecto. Esta sección se implementa en una historia próxima."
      />
      <Card title="Equipo" description="Todavía no hay miembros para mostrar.">
        <p className="text-sm text-ink-soft">
          Los usuarios asociados al proyecto y sus roles aparecerán acá.
        </p>
      </Card>
    </section>
  )
}
