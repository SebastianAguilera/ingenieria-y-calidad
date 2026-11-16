import { useEffect, useState } from 'react'
import type { KeyboardEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { listarProyectos } from '../api/proyectos'
import type { Proyecto, ProyectoListado } from '../api/types'
import PageHeader from '../components/PageHeader'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import ModalCrearProyecto from './ModalCrearProyecto'

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
  const [altaAbierta, setAltaAbierta] = useState(false)

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

  function alCrearProyecto(proyecto: Proyecto) {
    setAltaAbierta(false)
    navegar(`/proyectos/${proyecto.id}`)
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
          <Button onClick={() => setAltaAbierta(true)}>Crear proyecto</Button>
        </Card>
      )}

      {estado.tipo === 'listo' && estado.proyectos.length > 0 && (
        <div className="mb-5 flex justify-end">
          <Button onClick={() => setAltaAbierta(true)}>Agregar proyecto</Button>
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
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {altaAbierta && (
        <ModalCrearProyecto
          onCerrar={() => setAltaAbierta(false)}
          onCreado={alCrearProyecto}
        />
      )}
    </section>
  )
}
