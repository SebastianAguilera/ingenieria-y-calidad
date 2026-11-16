import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { obtenerProyecto } from '../../api/proyectos'
import { esApiError } from '../../api/errors'
import type { Proyecto } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import ModalEditarProyecto from './ModalEditarProyecto'

type EstadoCarga =
  | { tipo: 'cargando' }
  | { tipo: 'error'; es404: boolean }
  | { tipo: 'listo'; proyecto: Proyecto }

export default function EditarProyectoPage() {
  const { id } = useParams()
  const proyectoId = Number(id)
  const navegar = useNavigate()
  const [estado, setEstado] = useState<EstadoCarga>({ tipo: 'cargando' })
  const [intento, setIntento] = useState(0)

  useEffect(() => {
    const controlador = new AbortController()
    obtenerProyecto(proyectoId, controlador.signal)
      .then((proyecto) => setEstado({ tipo: 'listo', proyecto }))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        setEstado({ tipo: 'error', es404: esApiError(error) && error.status === 404 })
      })
    return () => controlador.abort()
  }, [proyectoId, intento])

  const volver = (
    <Link
      to={`/proyectos/${proyectoId}`}
      className="mb-3 inline-flex items-center gap-2 text-sm text-ink-soft transition-colors duration-150 hover:text-ink"
    >
      <i className="bi bi-arrow-left" aria-hidden="true" />
      Volver a información general
    </Link>
  )

  if (estado.tipo === 'cargando') {
    return (
      <section>
        {volver}
        <PageHeader title="Editar proyecto" description="Actualizá los datos del proyecto." />
        <p role="status" className="text-ink-soft">
          Cargando proyecto…
        </p>
      </section>
    )
  }

  if (estado.tipo === 'error') {
    return (
      <section>
        {volver}
        <PageHeader title="Editar proyecto" description="Actualizá los datos del proyecto." />
        {estado.es404 ? (
          <Alert tono="error">No tenés acceso a este proyecto o no existe</Alert>
        ) : (
          <div className="flex flex-col items-start gap-4">
            <Alert tono="error">No se pudo cargar el proyecto. Intentá nuevamente.</Alert>
            <Button
              variante="secondary"
              onClick={() => {
                setEstado({ tipo: 'cargando' })
                setIntento((anterior) => anterior + 1)
              }}
            >
              Reintentar
            </Button>
          </div>
        )}
      </section>
    )
  }

  return (
    <ModalEditarProyecto
      proyecto={estado.proyecto}
      onCerrar={() => navegar(`/proyectos/${proyectoId}`)}
      onGuardado={(proyecto) => setEstado({ tipo: 'listo', proyecto })}
    />
  )
}
