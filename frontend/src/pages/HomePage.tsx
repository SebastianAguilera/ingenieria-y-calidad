import { useEffect, useState } from 'react'
import { api } from '../api/client'
import { esApiError } from '../api/errors'
import type { ListadoProyectos } from '../api/types'
import PageHeader from '../components/PageHeader'

type EstadoBackend =
  | { tipo: 'cargando' }
  | { tipo: 'ok'; total: number }
  | { tipo: 'error'; mensaje: string }

export default function HomePage() {
  const [estado, setEstado] = useState<EstadoBackend>({ tipo: 'cargando' })

  useEffect(() => {
    let activo = true

    api
      .get<ListadoProyectos>('/proyectos')
      .then((data) => {
        if (activo) {
          setEstado({ tipo: 'ok', total: data.total })
        }
      })
      .catch((error: unknown) => {
        if (activo) {
          setEstado({
            tipo: 'error',
            mensaje: esApiError(error) ? error.message : 'Error desconocido',
          })
        }
      })

    return () => {
      activo = false
    }
  }, [])

  return (
    <section>
      <PageHeader
        title="Software Metrics & Estimation Engine"
        description="Base del frontend con Vite, React, TypeScript, Tailwind y React Router."
      />
      <div className="grid gap-4 sm:grid-cols-2">
        <article className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold">Backend</h2>
          <p className="mt-2 text-sm text-slate-600">
            {estado.tipo === 'cargando' && 'Consultando el backend...'}
            {estado.tipo === 'ok' && `Backend conectado: ${estado.total} proyecto(s).`}
            {estado.tipo === 'error' && `No se pudo conectar: ${estado.mensaje}`}
          </p>
        </article>
        <article className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold">Rutas</h2>
          <p className="mt-2 text-sm text-slate-600">
            Esta es la ruta principal. Cualquier ruta desconocida muestra la página 404.
          </p>
        </article>
      </div>
    </section>
  )
}
