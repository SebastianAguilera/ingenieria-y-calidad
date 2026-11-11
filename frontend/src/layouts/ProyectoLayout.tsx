import { Link, Outlet, useParams } from 'react-router-dom'
import { ProyectoProvider } from '../pages/proyecto/ProyectoProvider'
import { useProyecto } from '../pages/proyecto/proyectoContext'
import Alert from '../ui/Alert'
import Button from '../ui/Button'

function DetalleProyecto() {
  const { estado, reintentar } = useProyecto()

  return (
    <section className="flex flex-col gap-6">
      {estado.tipo === 'listo' && (
        <header className="flex flex-wrap items-center gap-3">
          <h2 className="text-2xl font-semibold tracking-tight text-ink">
            {estado.proyecto.nombre}
          </h2>
          {estado.proyecto.estado === 'Cerrado' && (
            <span className="rounded-pill border border-border bg-[var(--primary-100)] px-3 py-1 text-xs font-semibold text-ink-soft">
              Cerrado
            </span>
          )}
        </header>
      )}

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando proyecto…
        </p>
      )}

      {estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">No se pudo cargar el proyecto. Intentá nuevamente.</Alert>
          <Button variante="secondary" onClick={reintentar}>
            Reintentar
          </Button>
        </div>
      )}

      {estado.tipo === 'noEncontrado' && (
        <div className="flex flex-col items-start gap-4">
          <p className="text-ink">Proyecto no encontrado</p>
          <Link
            to="/proyectos"
            className="inline-flex items-center justify-center rounded-pill bg-primary px-5 py-2.5 text-sm font-semibold text-[var(--on-primary)] transition hover:bg-[var(--primary-700)] focus-visible:shadow-[var(--ring)]"
          >
            Volver a proyectos
          </Link>
        </div>
      )}

      {estado.tipo === 'listo' && <Outlet />}
    </section>
  )
}

export default function ProyectoLayout() {
  const { id } = useParams()

  return (
    <ProyectoProvider key={id}>
      <DetalleProyecto />
    </ProyectoProvider>
  )
}
