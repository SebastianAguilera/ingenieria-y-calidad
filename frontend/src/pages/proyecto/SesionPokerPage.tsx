import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { esApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import { crearSesionPoker, listarSesionesPoker } from '../../api/poker'
import type { Historia, SesionPokerCreada, SesionPokerHistorial } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'

type EstadoCarga =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'noEncontrada'; mensaje: string }
  | { tipo: 'listo'; historia: Historia; sesiones: SesionPokerHistorial[] }

const MENSAJE_ERROR_CARGA = 'No se pudo cargar el Planning Poker. Intentá nuevamente.'
const MENSAJE_ERROR_INICIO = 'No se pudo iniciar la sesión. Intentá nuevamente.'
const MENSAJE_HISTORIA_NO_ENCONTRADA =
  'La historia no existe o no pertenece a este proyecto.'
const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'
const MENSAJE_HISTORIA_NO_ESTIMABLE =
  'La historia está en estado Hecho y no se puede estimar.'
const MENSAJE_SESION_ACTIVA = 'Ya hay una sesión activa. Te unimos a ella.'

function parsearId(valor: string | undefined): number | null {
  if (valor === undefined) {
    return null
  }
  const numero = Number(valor)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

function esSesionActiva(sesion: SesionPokerHistorial): boolean {
  return sesion.estado === 'Votando' || sesion.estado === 'Revelada'
}

function formatearFecha(fecha: string | null | undefined): string {
  if (!fecha) {
    return '—'
  }
  const [anio, mes, dia] = fecha.slice(0, 10).split('-')
  return `${dia}/${mes}/${anio}`
}

function aHistorial(creada: SesionPokerCreada): SesionPokerHistorial {
  return {
    id: creada.id,
    estado: creada.estado,
    ronda_actual: creada.ronda_actual,
    puntos_acordados: creada.puntos_acordados,
    creado_por: creada.creado_por,
    creado_en: creada.creado_en,
    cerrado_en: creada.cerrado_en ?? null,
  }
}

function ResultadoCerrado({ sesion }: { sesion: SesionPokerHistorial }) {
  return (
    <Card
      title="Última sesión cerrada"
      description={`Ronda ${sesion.ronda_actual} · Cerrada el ${formatearFecha(sesion.cerrado_en)}`}
      className="mb-5"
    >
      {sesion.puntos_acordados === null ? (
        <p className="text-ink-soft">Sin puntos acordados</p>
      ) : (
        <p className="text-ink">Puntos acordados: {sesion.puntos_acordados}</p>
      )}
    </Card>
  )
}

export default function SesionPokerPage() {
  const { id, historiaId: historiaParam } = useParams()
  const proyectoId = parsearId(id)
  const historiaId = parsearId(historiaParam)
  const [intento, setIntento] = useState(0)
  const [consulta, setConsulta] = useState<{ clave: string; estado: EstadoCarga } | null>(null)
  const [creando, setCreando] = useState(false)
  const [errorInicio, setErrorInicio] = useState<string | null>(null)
  const [avisoInicio, setAvisoInicio] = useState<string | null>(null)

  const parametrosValidos = proyectoId !== null && historiaId !== null
  const clave = parametrosValidos ? `${proyectoId}:${historiaId}:${intento}` : 'invalido'

  const estado: EstadoCarga = !parametrosValidos
    ? { tipo: 'noEncontrada', mensaje: MENSAJE_HISTORIA_NO_ENCONTRADA }
    : consulta?.clave === clave
      ? consulta.estado
      : { tipo: 'cargando' }

  useEffect(() => {
    if (proyectoId === null || historiaId === null) {
      return
    }

    const controlador = new AbortController()

    Promise.all([
      listarHistorias(proyectoId, controlador.signal),
      listarSesionesPoker(proyectoId, historiaId, controlador.signal),
    ])
      .then(([historias, sesiones]) => {
        const historia = historias.find((actual) => actual.id === historiaId)
        if (!historia) {
          setConsulta({
            clave,
            estado: { tipo: 'noEncontrada', mensaje: MENSAJE_HISTORIA_NO_ENCONTRADA },
          })
          return
        }
        setConsulta({ clave, estado: { tipo: 'listo', historia, sesiones } })
      })
      .catch((error: unknown) => {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }
        if (esApiError(error) && error.status === 404) {
          const mensaje =
            error.code === 'PROYECTO_NO_ENCONTRADO'
              ? MENSAJE_PROYECTO_INEXISTENTE
              : MENSAJE_HISTORIA_NO_ENCONTRADA
          setConsulta({ clave, estado: { tipo: 'noEncontrada', mensaje } })
          return
        }
        setConsulta({ clave, estado: { tipo: 'error', mensaje: MENSAJE_ERROR_CARGA } })
      })

    return () => controlador.abort()
  }, [clave, historiaId, proyectoId])

  async function iniciar() {
    if (estado.tipo !== 'listo' || proyectoId === null || historiaId === null) {
      return
    }

    setCreando(true)
    setErrorInicio(null)
    setAvisoInicio(null)

    try {
      const creada = await crearSesionPoker(proyectoId, historiaId)
      setConsulta({
        clave,
        estado: {
          tipo: 'listo',
          historia: estado.historia,
          sesiones: [aHistorial(creada), ...estado.sesiones],
        },
      })
    } catch (error) {
      if (esApiError(error)) {
        if (error.code === 'HISTORIA_NO_ESTIMABLE') {
          setErrorInicio(MENSAJE_HISTORIA_NO_ESTIMABLE)
          return
        }
        if (error.code === 'SESION_POKER_ACTIVA') {
          try {
            const sesiones = await listarSesionesPoker(proyectoId, historiaId)
            setConsulta({ clave, estado: { tipo: 'listo', historia: estado.historia, sesiones } })
            setAvisoInicio(MENSAJE_SESION_ACTIVA)
          } catch {
            setErrorInicio(MENSAJE_ERROR_INICIO)
          }
          return
        }
        if (error.code === 'PROYECTO_NO_ENCONTRADO') {
          setErrorInicio(MENSAJE_PROYECTO_INEXISTENTE)
          return
        }
        if (error.code === 'HISTORIA_NO_ENCONTRADA') {
          setErrorInicio(MENSAJE_HISTORIA_NO_ENCONTRADA)
          return
        }
        if (error.status !== 0 && error.message !== '') {
          setErrorInicio(error.message)
          return
        }
      }
      setErrorInicio(MENSAJE_ERROR_INICIO)
    } finally {
      setCreando(false)
    }
  }

  const urlBacklog = `/proyectos/${proyectoId ?? ''}/backlog`

  if (estado.tipo === 'cargando') {
    return (
      <section>
        <PageHeader title="Planning Poker" description="Estimación en equipo por historia." />
        <p role="status" className="text-ink-soft">
          Cargando Planning Poker…
        </p>
      </section>
    )
  }

  if (estado.tipo === 'error') {
    return (
      <section>
        <PageHeader title="Planning Poker" description="Estimación en equipo por historia." />
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{estado.mensaje}</Alert>
          <Button variante="secondary" onClick={() => setIntento((anterior) => anterior + 1)}>
            Reintentar
          </Button>
        </div>
      </section>
    )
  }

  if (estado.tipo === 'noEncontrada') {
    return (
      <section>
        <PageHeader title="Planning Poker" description="Estimación en equipo por historia." />
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{estado.mensaje}</Alert>
          <Link
            to={urlBacklog}
            className="text-sm font-semibold text-primary hover:underline focus-visible:shadow-[var(--ring)]"
          >
            Volver al backlog
          </Link>
        </div>
      </section>
    )
  }

  const { historia, sesiones } = estado
  const sesionActiva = sesiones.find(esSesionActiva)
  const ultimaSesion = sesiones[0]

  return (
    <section>
      <PageHeader
        title="Planning Poker"
        description={`${historia.codigo} · ${historia.titulo}`}
      />

      <div className="mb-5">
        <Link
          to={urlBacklog}
          className="text-sm font-semibold text-primary hover:underline focus-visible:shadow-[var(--ring)]"
        >
          Volver al backlog
        </Link>
      </div>

      {errorInicio && (
        <div className="mb-5">
          <Alert tono="error">{errorInicio}</Alert>
        </div>
      )}

      {avisoInicio && (
        <div className="mb-5">
          <Alert tono="info">{avisoInicio}</Alert>
        </div>
      )}

      {sesionActiva ? (
        <Card
          title="Sesión en curso"
          description={`Creada por ${sesionActiva.creado_por.nombre} el ${formatearFecha(sesionActiva.creado_en)}.`}
        >
          <dl className="flex flex-wrap items-center gap-x-8 gap-y-3 text-sm">
            <div className="flex items-center gap-2">
              <dt className="text-ink-soft">Estado</dt>
              <dd>
                <span className="inline-flex items-center rounded-pill border border-border bg-[var(--surface-2)] px-3 py-1.5 text-xs font-semibold text-ink">
                  {sesionActiva.estado}
                </span>
              </dd>
            </div>
            <div className="flex items-center gap-2">
              <dt className="text-ink-soft">Ronda</dt>
              <dd className="font-medium text-ink">Ronda {sesionActiva.ronda_actual}</dd>
            </div>
          </dl>
          <p className="mt-4 text-sm text-ink-soft">
            {sesionActiva.estado === 'Votando'
              ? 'El equipo está votando. Los votos se revelan al finalizar la ronda.'
              : 'Los votos fueron revelados. La sesión sigue activa.'}
          </p>
        </Card>
      ) : (
        <>
          {ultimaSesion?.estado === 'Cerrada' && <ResultadoCerrado sesion={ultimaSesion} />}

          {ultimaSesion?.estado === 'Cancelada' && (
            <div className="mb-5">
              <Alert tono="info">La última sesión fue cancelada.</Alert>
            </div>
          )}

          <Card
            title="Iniciar Planning Poker"
            description="La sesión se crea recién cuando confirmás."
          >
            <p className="text-sm text-ink-soft">
              Estado de la historia: <span className="font-semibold text-ink">{historia.estado}</span>
            </p>
            {ultimaSesion === undefined && (
              <p className="mt-2 text-sm text-ink-soft">
                Todavía no hay sesiones de Planning Poker para esta historia.
              </p>
            )}
            <div className="mt-4">
              <Button onClick={() => void iniciar()} disabled={creando} aria-busy={creando}>
                Iniciar Planning Poker
              </Button>
            </div>
          </Card>
        </>
      )}
    </section>
  )
}
