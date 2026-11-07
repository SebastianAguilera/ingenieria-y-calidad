import PageHeader from '../components/PageHeader'
import { useSesion } from '../session/sesion'
import Card from '../ui/Card'

export default function ProyectosPage() {
  const { usuarioActivo } = useSesion()

  return (
    <section>
      <PageHeader
        title={usuarioActivo ? `Hola, ${usuarioActivo.nombre}` : 'Proyectos'}
        description="Este es el punto de entrada del panel. El listado y la navegación de proyectos se incorporan en las próximas historias."
      />
      <Card title="Proyectos" description="Todavía no hay proyectos para mostrar en esta vista.">
        <p className="text-sm text-ink-soft">
          La sesión es simulada: podés cerrarla desde el botón Salir del encabezado.
        </p>
      </Card>
    </section>
  )
}
