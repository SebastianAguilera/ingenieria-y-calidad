import { useId, useState } from 'react'
import { useParams } from 'react-router-dom'
import { descargarPdfReporte } from '../../api/reportes'
import type { Sprint } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import DatosSprint from './metricas/DatosSprint'
import ResumenKpi from './metricas/ResumenKpi'
import EncabezadoReporte from './reporte/EncabezadoReporte'
import { guardarArchivo } from './reporte/guardarArchivo'
import {
  fechaLocal,
  nombreArchivoProyecto,
  nombreArchivoSprint,
  type ExtensionReporte,
} from './reporte/nombreArchivo'
import { TablaDefectosReporte, TablaHistoriasReporte } from './reporte/TablasReporte'
import { useReporte, useSprintsDelReporte, type Reporte } from './reporte/useReporte'

const ALCANCE_PROYECTO = 'proyecto'
export const MENSAJE_ERROR_PDF = 'No se pudo descargar el PDF. Intentá nuevamente.'

const CLASES_SELECT =
  'appearance-none rounded-pill border border-border bg-[var(--surface-2)] py-2 pl-3.5 pr-9 text-sm text-ink focus:border-primary focus:outline-none'

function parsearProyectoId(id: string | undefined): number | null {
  const numero = Number(id)
  return id !== undefined && Number.isInteger(numero) && numero > 0 ? numero : null
}

function nombreDeArchivo(reporte: Reporte, extension: ExtensionReporte): string {
  const fecha = fechaLocal(new Date())
  return reporte.alcance === 'proyecto'
    ? nombreArchivoProyecto(reporte.datos.nombre, fecha, extension)
    : nombreArchivoSprint(reporte.datos.proyecto_id, reporte.datos.identificador, fecha, extension)
}

function SelectorAlcance({
  sprints,
  valor,
  onCambiar,
}: {
  sprints: Sprint[]
  valor: string
  onCambiar: (valor: string) => void
}) {
  const id = useId()
  return (
    <div className="flex items-center gap-2 text-sm text-ink-soft">
      <label htmlFor={id} className="font-medium">
        Alcance
      </label>
      <span className="relative inline-flex items-center">
        <select
          id={id}
          value={valor}
          onChange={(evento) => onCambiar(evento.target.value)}
          className={CLASES_SELECT}
        >
          <option value={ALCANCE_PROYECTO}>Proyecto completo</option>
          {sprints.map((sprint) => (
            <option key={sprint.id} value={String(sprint.id)}>
              {`${sprint.identificador} (${sprint.estado})`}
            </option>
          ))}
        </select>
        <i
          className="bi bi-chevron-down pointer-events-none absolute right-3 text-xs text-ink-soft"
          aria-hidden="true"
        />
      </span>
    </div>
  )
}

function ContenidoReporte({ reporte, sprints }: { reporte: Reporte; sprints: Sprint[] }) {
  return (
    <div className="flex flex-col gap-5">
      <EncabezadoReporte reporte={reporte} />
      <Alert tono="info">
        <strong>Calculado sobre el estado actual de las historias</strong>
        <span className="block text-ink-soft">
          El sistema no guarda histórico: el reporte refleja los datos al momento de consultarlo.
        </span>
      </Alert>
      {reporte.alcance === 'proyecto' ? (
        <ResumenKpi metricas={reporte.datos.metricas} />
      ) : (
        <Card title="Métricas del sprint">
          <DatosSprint metricas={reporte.datos.metricas} />
        </Card>
      )}
      <TablaHistoriasReporte historias={reporte.datos.historias} />
      <TablaDefectosReporte defectos={reporte.datos.defectos} sprints={sprints} />
    </div>
  )
}

/** Reporte del proyecto o de un sprint, con descarga en JSON y PDF (US-F23). */
export default function ReportePage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const [alcance, setAlcance] = useState(ALCANCE_PROYECTO)
  const sprintId = alcance === ALCANCE_PROYECTO ? null : Number(alcance)
  const sprints = useSprintsDelReporte(proyectoId)
  const reporte = useReporte(proyectoId, sprintId)
  const [generandoPdf, setGenerandoPdf] = useState(false)
  const [errorPdf, setErrorPdf] = useState<string | null>(null)

  const reporteListo = reporte.estado.tipo === 'listo' ? reporte.estado.valor : null

  function cambiarAlcance(valor: string) {
    setAlcance(valor)
    setErrorPdf(null)
  }

  function descargarJson() {
    if (reporteListo === null) {
      return
    }
    const contenido = new Blob([JSON.stringify(reporteListo.datos, null, 2)], {
      type: 'application/json',
    })
    guardarArchivo(contenido, nombreDeArchivo(reporteListo, 'json'))
  }

  async function descargarPdf() {
    if (reporteListo === null || proyectoId === null) {
      return
    }
    setGenerandoPdf(true)
    setErrorPdf(null)
    try {
      const pdf = await descargarPdfReporte(proyectoId, sprintId)
      guardarArchivo(pdf, nombreDeArchivo(reporteListo, 'pdf'))
    } catch {
      setErrorPdf(MENSAJE_ERROR_PDF)
    } finally {
      setGenerandoPdf(false)
    }
  }

  return (
    <section className="mx-auto w-full max-w-6xl">
      <PageHeader
        title="Reporte"
        description="Reporte consolidado del proyecto o de un sprint: datos generales, métricas, historias y defectos."
        className="mb-6"
      />

      {sprints.estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando reporte…
        </p>
      )}

      {sprints.estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{sprints.estado.mensaje}</Alert>
          <Button
            variante="secondary"
            onClick={() => {
              sprints.reintentar()
              reporte.reintentar()
            }}
          >
            Reintentar
          </Button>
        </div>
      )}

      {sprints.estado.tipo === 'listo' && (
        <div className="flex flex-col gap-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <SelectorAlcance sprints={sprints.estado.valor} valor={alcance} onCambiar={cambiarAlcance} />
            <div className="flex flex-wrap gap-2">
              <Button variante="secondary" disabled={reporteListo === null} onClick={descargarJson}>
                <i className="bi bi-filetype-json" aria-hidden="true" />
                Descargar JSON
              </Button>
              <Button disabled={reporteListo === null || generandoPdf} onClick={() => void descargarPdf()}>
                <i className="bi bi-file-earmark-pdf" aria-hidden="true" />
                {generandoPdf ? 'Generando PDF…' : 'Descargar PDF'}
              </Button>
            </div>
          </div>

          {errorPdf && <Alert tono="error">{errorPdf}</Alert>}

          {reporte.estado.tipo === 'cargando' && (
            <p role="status" className="text-ink-soft">
              Cargando reporte…
            </p>
          )}

          {reporte.estado.tipo === 'error' && (
            <div className="flex flex-col items-start gap-4">
              <Alert tono="error">{reporte.estado.mensaje}</Alert>
              <Button variante="secondary" onClick={reporte.reintentar}>
                Reintentar
              </Button>
            </div>
          )}

          {reporteListo !== null && <ContenidoReporte reporte={reporteListo} sprints={sprints.estado.valor} />}
        </div>
      )}
    </section>
  )
}
