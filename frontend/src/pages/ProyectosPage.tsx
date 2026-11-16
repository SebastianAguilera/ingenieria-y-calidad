import { useEffect, useState } from 'react'
import type { KeyboardEvent, MouseEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { eliminarProyecto, listarProyectos } from '../api/proyectos'
import { esApiError } from '../api/errors'
import type { ProyectoListado } from '../api/types'
import PageHeader from '../components/PageHeader'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import ModalConfirmacion from '../ui/ModalConfirmacion'

type EstadoCarga =
  | { tipo: 'cargando' }
  | { tipo: 'error' }
  | { tipo: 'listo'; proyectos: ProyectoListado[] }

function formatearFecha(fecha: string | null): string {
  if (!fecha) {
    return 'Sin fecha'
  }
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
}

export default function ProyectosPage() {
  const navegar = useNavigate()
  const [estado, setEstado] = useState<EstadoCarga>({ tipo: 'cargando' })

  const [intento, setIntento] = useState(0)
  const [proyectoAEliminar, setProyectoAEliminar] = useState<ProyectoListado | null>(null)
  const [errorEliminar, setErrorEliminar] = useState<string | null>(null)
  const [eliminando, setEliminando] = useState(false)

  useEffect(() => {
    const controlador = new AbortController()
    listarProyectos(controlador.signal)
      .then((proyectos) => setEstado({ tipo: 'listo', proyectos }))
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        setEstado({ tipo: 'error' })
      })
    return () => controlador.abort()
  }, [intento])

  function reintentar() {
    setEstado({ tipo: 'cargando' })
    setIntento((anterior) => anterior + 1)
  }

  function irAlDetalle(id: number) {
    navegar(`/proyectos/${id}`)
  }

  function alPresionarTecla(evento: KeyboardEvent<HTMLTableRowElement>, id: number) {
    if (evento.key === 'Enter') {
      irAlDetalle(id)
    }
  }

  function pedirEliminar(evento: MouseEvent<HTMLButtonElement>, proyecto: ProyectoListado) {
    evento.stopPropagation()
    setErrorEliminar(null)
    setProyectoAEliminar(proyecto)
  }

  function cancelarEliminar() {
    if (eliminando) {
      return
    }
    setProyectoAEliminar(null)
    setErrorEliminar(null)
  }

  async function confirmarEliminar() {
    if (!proyectoAEliminar || eliminando) {
      return
    }

    setEliminando(true)
    setErrorEliminar(null)
    try {
      await eliminarProyecto(proyectoAEliminar.id)
      setEstado((anterior) =>
        anterior.tipo === 'listo'
          ? {
              tipo: 'listo',
              proyectos: anterior.proyectos.filter((p) => p.id !== proyectoAEliminar.id),
            }
          : anterior,
      )
      setProyectoAEliminar(null)
    } catch (error) {
      setErrorEliminar(mensajeDeEliminacion(error))
    } finally {
      setEliminando(false)
    }
  }

  return (
    <section>
      <PageHeader title="Proyectos" description="Elegí el proyecto en el que querés trabajar." />

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando proyectos…
        </p>
      )}

      {estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">No se pudieron cargar los proyectos. Intentá nuevamente.</Alert>
          <Button variante="secondary" onClick={reintentar}>
            Reintentar
          </Button>
        </div>
      )}

      {estado.tipo === 'listo' && estado.proyectos.length === 0 && (
        <Card title="Todavía no participás en ningún proyecto">
          <Button onClick={() => navegar('/proyectos/nuevo')}>Crear proyecto</Button>
        </Card>
      )}

      {estado.tipo === 'listo' && estado.proyectos.length > 0 && (
        <div className="mb-5 flex justify-end">
          <Button onClick={() => navegar('/proyectos/nuevo')}>Agregar proyecto</Button>
        </div>
      )}

      {estado.tipo === 'listo' && estado.proyectos.length > 0 && (
        <div className="tabla-contenedor">
          <table className="tabla min-w-[32rem]">
            <thead>
              <tr>
                <th scope="col" className="text-center">
                  Nombre
                </th>
                <th scope="col" className="text-center">
                  Estado
                </th>
                <th scope="col" className="text-center">
                  Fecha de inicio
                </th>
                <th scope="col" className="text-center">
                  Fecha de fin
                </th>
                <th scope="col" className="col-acciones">
                  Acciones
                </th>
              </tr>
            </thead>
            <tbody>
              {estado.proyectos.map((proyecto) => (
                <tr
                  key={proyecto.id}
                  tabIndex={0}
                  onClick={() => irAlDetalle(proyecto.id)}
                  onKeyDown={(evento) => alPresionarTecla(evento, proyecto.id)}
                  className="cursor-pointer focus-visible:shadow-[var(--ring)]"
                >
                  <td className="text-center font-medium text-ink">{proyecto.nombre}</td>
                  <td className="text-center">{proyecto.estado}</td>
                  <td className="text-center">{formatearFecha(proyecto.fecha_inicio)}</td>
                  <td className="text-center">{formatearFecha(proyecto.fecha_fin)}</td>
                  <td className="col-acciones">
                    <div className="tabla-acciones">
                      <button
                        type="button"
                        aria-label={`Eliminar ${proyecto.nombre}`}
                        title="Eliminar proyecto"
                        className="inline-flex items-center gap-1 rounded-pill px-3 py-1.5 text-sm font-semibold text-[var(--danger)] transition hover:bg-[var(--accent-100)] focus-visible:shadow-[var(--ring)]"
                        onClick={(evento) => pedirEliminar(evento, proyecto)}
                        onKeyDown={(evento) => evento.stopPropagation()}
                      >
                        <i className="bi bi-trash" aria-hidden="true" />
                        Eliminar
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {proyectoAEliminar && (
        <ModalConfirmacion
          titulo="Eliminar proyecto"
          textoConfirmar="Eliminar"
          procesando={eliminando}
          onConfirmar={() => void confirmarEliminar()}
          onCancelar={cancelarEliminar}
        >
          {errorEliminar && <Alert tono="error">{errorEliminar}</Alert>}
          <p>
            ¿Querés eliminar <strong className="text-ink">{proyectoAEliminar.nombre}</strong>?
          </p>
          <p>Esta acción no se puede deshacer desde la interfaz.</p>
        </ModalConfirmacion>
      )}
    </section>
  )
}

function mensajeDeEliminacion(error: unknown): string {
  if (esApiError(error)) {
    if (error.status === 0) {
      return 'No se pudo conectar con el servidor. Intentá nuevamente.'
    }
    if (error.status === 404) {
      return 'No tenés acceso a este proyecto o no existe'
    }
    if (error.code === 'PROYECTO_CON_HISTORIAL') {
      return error.message
    }
  }
  return 'No se pudo eliminar el proyecto. Intentá nuevamente.'
}
