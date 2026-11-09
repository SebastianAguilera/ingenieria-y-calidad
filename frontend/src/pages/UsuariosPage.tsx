import PageHeader from '../components/PageHeader'
import Card from '../ui/Card'

export default function UsuariosPage() {
  return (
    <section>
      <PageHeader
        title="Usuarios"
        description="Gestión de usuarios del panel. Esta sección se implementa en una historia próxima."
      />
      <Card title="Usuarios" description="Todavía no hay usuarios para mostrar en esta vista.">
        <p className="text-sm text-ink-soft">
          El alta y la edición de usuarios llegarán con el módulo de gestión de usuarios.
        </p>
      </Card>
    </section>
  )
}
