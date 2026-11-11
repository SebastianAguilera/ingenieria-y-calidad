import { useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ActualizacionHistoria, EstadoHistoria, Historia } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalHistoria from './backlog/ModalHistoria'
import TablaBacklog from './backlog/TablaBacklog'
import TableroBacklog from './backlog/TableroBacklog'
import { useHistorias } from './backlog/useHistorias'

type Vista = 'tabla' | 'tablero'

type ModalAbierto = { tipo: 'crear' } | { tipo: 'editar'; historia: Historia }

function parsearProyectoId(id: string | undefined): number | null {
  if (id === undefined) {
    return null
  }
  const numero = Number(id)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

function SelectorVista({
  vista,
  onChange,
}: {
  vista: Vista
  onChange: (vista: Vista) => void
}) {
  const opciones: Array<{ valor: Vista; etiqueta: string; icono: string }> = [
    { valor: 'tabla', etiqueta: 'Tabla', icono: 'bi-table' },
    { valor: 'tablero', etiqueta: 'Tablero', icono: 'bi-kanban' },
  ]

  return (
    <div
      role="group"
      aria-label="Vista del backlog"
      className="inline-flex rounded-pill border border-border bg-surface p-1"
    >
      {opciones.map(({ valor, etiqueta, icono }) => (
        <button
          key={valor}
          type="button"
          aria-pressed={vista === valor}
          onClick={() => onChange(valor)}
          className={`inline-flex items-center gap-2 rounded-pill px-3 py-1.5 text-sm font-semibold transition focus-visible:shadow-[var(--ring)] ${
            vista === valor
              ? 'bg-primary text-[var(--on-primary)]'
              : 'text-ink-soft hover:bg-[var(--primary-100)] hover:text-ink'
          }`}
        >
          <i className={`bi ${icono}`} aria-hidden="true" />
          {etiqueta}
        </button>
      ))}
    </div>
  )
}

export default function BacklogPage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const { estado, reintentar, crear, actualizar, guardandoId } = useHistorias(proyectoId)
  const [vista, setVista] = useState<Vista>('tabla')
  const [modal, setModal] = useState<ModalAbierto | null>(null)
  const [errorOperacion, setErrorOperacion] = useState<string | null>(null)

  function cambiar(historia: Historia, cambios: ActualizacionHistoria) {
    setErrorOperacion(null)
    void actualizar(historia, cambios).then((resultado) => {
      if (!resultado.ok) {
        setErrorOperacion(resultado.mensaje)
      }
    })
  }

  function mover(historia: Historia, nuevoEstado: EstadoHistoria) {
    cambiar(historia, { estado: nuevoEstado })
  }

  return (
    <section>
      <PageHeader
        title="Backlog"
        description="Historias de usuario del proyecto, con su estado y prioridad."
      />

      {errorOperacion && (
        <div className="mb-5">
          <Alert tono="error">{errorOperacion}</Alert>
        </div>
      )}

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando historias…
        </p>
      )}

      {estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{estado.mensaje}</Alert>
          <Button variante="secondary" onClick={reintentar}>
            Reintentar
          </Button>
        </div>
      )}

      {estado.tipo === 'listo' && estado.historias.length === 0 && (
        <Card
          title="Todavía no hay historias"
          description="Creá la primera historia para empezar a planificar el proyecto."
        >
          <Button onClick={() => setModal({ tipo: 'crear' })}>Nueva historia</Button>
        </Card>
      )}

      {estado.tipo === 'listo' && estado.historias.length > 0 && (
        <>
          <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
            <SelectorVista vista={vista} onChange={setVista} />
            <Button onClick={() => setModal({ tipo: 'crear' })}>Nueva historia</Button>
          </div>

          {vista === 'tabla' ? (
            <TablaBacklog
              historias={estado.historias}
              onCambiar={cambiar}
              onEditar={(historia) => setModal({ tipo: 'editar', historia })}
              guardandoId={guardandoId}
            />
          ) : (
            <TableroBacklog
              historias={estado.historias}
              onMover={mover}
              onEditar={(historia) => setModal({ tipo: 'editar', historia })}
              guardandoId={guardandoId}
            />
          )}
        </>
      )}

      {modal && modal.tipo === 'crear' && (
        <ModalHistoria
          titulo="Nueva historia"
          textoBoton="Crear historia"
          onCerrar={() => setModal(null)}
          onGuardar={async (datos) => {
            const resultado = await crear(datos)
            if (resultado.ok) {
              setModal(null)
            }
            return resultado
          }}
        />
      )}

      {modal && modal.tipo === 'editar' && (
        <ModalHistoria
          titulo="Editar historia"
          textoBoton="Guardar cambios"
          historia={modal.historia}
          onCerrar={() => setModal(null)}
          onGuardar={async (datos) => {
            const resultado = await actualizar(modal.historia, datos)
            if (resultado.ok) {
              setModal(null)
            }
            return resultado
          }}
        />
      )}
    </section>
  )
}
