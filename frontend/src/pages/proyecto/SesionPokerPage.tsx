import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { esApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import {
  crearSesionPoker,
  listarSesionesPoker,
  obtenerDetalleSesionPoker,
  votarSesionPoker,
} from '../../api/poker'
import type {
  DetalleSesionPoker,
  Historia,
  SesionPokerCreada,
  SesionPokerHistorial,
} from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import CartasVoto from './poker/CartasVoto'

type EstadoCarga =
  | { tipo: 'cargando' }
  | { tipo: 'error'; mensaje: string }
  | { tipo: 'noEncontrada'; mensaje: string }
  | {
      tipo: 'listo'
      historia: Historia
      sesiones: SesionPokerHistorial[]
      detalle: DetalleSesionPoker | null
    }

const MENSAJE_ERROR_CARGA = 'No se pudo cargar el Planning Poker. Intentá nuevamente.'
const MENSAJE_ERROR_INICIO = 'No se pudo iniciar la sesión. Intentá nuevamente.'
const MENSAJE_ERROR_VOTO = 'No se pudo registrar el voto. Intentá nuevamente.'
const MENSAJE_VOTACION_CERRADA = 'La votación ya cerró'
const MENSAJE_ERROR_POLLING = 'No se pudo actualizar la sesión. Reintentando…'
const MENSAJE_HISTORIA_NO_ENCONTRADA =
  'La historia no existe o no pertenece a este proyecto.'
const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'
const MENSAJE_SESION_NO_ENCONTRADA = 'La sesión de Planning Poker no existe.'
const MENSAJE_HISTORIA_NO_ESTIMABLE =
  'La historia está en estado Hecho y no se puede estimar.'
const MENSAJE_SESION_ACTIVA = 'Ya hay una sesión activa. Te unimos a ella.'
const MENSAJE_PUNTOS_INVALIDOS = 'El valor elegido no es válido. Elegí otra carta.'

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
  const [enviandoVoto, setEnviandoVoto] = useState(false)
  const [errorVoto, setErrorVoto] = useState<string | null>(null)
  const [avisoVoto, setAvisoVoto] = useState<string | null>(null)
  const [errorPolling, setErrorPolling] = useState<string | null>(null)

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

    async function cargar() {
      try {
        const [historias, sesiones] = await Promise.all([
          listarHistorias(proyectoId as number, controlador.signal),
          listarSesionesPoker(proyectoId as number, historiaId as number, controlador.signal),
        ])

        if (controlador.signal.aborted) {
          return
        }

        const historia = historias.find((actual) => actual.id === historiaId)
        if (!historia) {
          setConsulta({
            clave,
            estado: { tipo: 'noEncontrada', mensaje: MENSAJE_HISTORIA_NO_ENCONTRADA },
          })
          return
        }

        const sesionActiva = sesiones.find(esSesionActiva)
        const detalle = sesionActiva
          ? await obtenerDetalleSesionPoker(
              proyectoId as number,
              sesionActiva.id,
              controlador.signal,
            )
          : null

        if (controlador.signal.aborted) {
          return
        }

        setErrorPolling(null)
        setConsulta({
          clave,
          estado: { tipo: 'listo', historia, sesiones, detalle },
        })
      } catch (error) {
        if (
          controlador.signal.aborted ||
          (error instanceof Error && error.name === 'AbortError')
        ) {
          return
        }
        if (esApiError(error) && error.status === 404) {
          const mensaje =
            error.code === 'PROYECTO_NO_ENCONTRADO'
              ? MENSAJE_PROYECTO_INEXISTENTE
              : error.code === 'SESION_POKER_NO_ENCONTRADA'
                ? MENSAJE_SESION_NO_ENCONTRADA
                : MENSAJE_HISTORIA_NO_ENCONTRADA
          setConsulta({ clave, estado: { tipo: 'noEncontrada', mensaje } })
          return
        }
        setConsulta({ clave, estado: { tipo: 'error', mensaje: MENSAJE_ERROR_CARGA } })
      }
    }

    void cargar()

    return () => controlador.abort()
  }, [clave, historiaId, proyectoId])

  function actualizarDetalle(detalle: DetalleSesionPoker) {
    setConsulta((actual) => {
      if (!actual || actual.estado.tipo !== 'listo') {
        return actual
      }
      return {
        clave: actual.clave,
        estado: { ...actual.estado, detalle },
      }
    })
  }

  async function iniciar() {
    if (estado.tipo !== 'listo' || proyectoId === null || historiaId === null) {
      return
    }

    setCreando(true)
    setErrorInicio(null)
    setAvisoInicio(null)

    try {
      const creada = await crearSesionPoker(proyectoId, historiaId)
      const detalle = await obtenerDetalleSesionPoker(proyectoId, creada.id)
      setConsulta({
        clave,
        estado: {
          tipo: 'listo',
          historia: estado.historia,
          sesiones: [aHistorial(creada), ...estado.sesiones],
          detalle,
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
            const sesionActiva = sesiones.find(esSesionActiva)
            const detalle = sesionActiva
              ? await obtenerDetalleSesionPoker(proyectoId, sesionActiva.id)
              : null
            setConsulta({
              clave,
              estado: { tipo: 'listo', historia: estado.historia, sesiones, detalle },
            })
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

  async function votar(puntos: number) {
    if (
      estado.tipo !== 'listo' ||
      proyectoId === null ||
      estado.detalle === null ||
      estado.detalle.estado !== 'Votando' ||
      enviandoVoto
    ) {
      return
    }

    const sesionId = estado.detalle.id
    setEnviandoVoto(true)
    setErrorVoto(null)
    setAvisoVoto(null)

    try {
      const resultado = await votarSesionPoker(proyectoId, sesionId, puntos)
      setConsulta((actual) => {
        if (
          !actual ||
          actual.estado.tipo !== 'listo' ||
          actual.estado.detalle === null ||
          actual.estado.detalle.id !== sesionId
        ) {
          return actual
        }
        return {
          clave: actual.clave,
          estado: {
            ...actual.estado,
            detalle: {
              ...actual.estado.detalle,
              mi_voto: resultado.mi_voto,
              ya_vote: true,
              votos_registrados: resultado.votos_registrados,
              total_miembros: resultado.total_miembros,
            },
          },
        }
      })
    } catch (error) {
      if (esApiError(error) && error.code === 'SESION_POKER_NO_VOTANDO') {
        try {
          const detalle = await obtenerDetalleSesionPoker(proyectoId, sesionId)
          actualizarDetalle(detalle)
          setAvisoVoto(MENSAJE_VOTACION_CERRADA)
        } catch {
          setErrorVoto(MENSAJE_ERROR_VOTO)
        }
        return
      }
      if (esApiError(error)) {
        if (error.code === 'PUNTOS_POKER_INVALIDOS') {
          setErrorVoto(MENSAJE_PUNTOS_INVALIDOS)
          return
        }
        if (error.code === 'PROYECTO_NO_ENCONTRADO') {
          setErrorVoto(MENSAJE_PROYECTO_INEXISTENTE)
          return
        }
        if (error.code === 'SESION_POKER_NO_ENCONTRADA') {
          setErrorVoto(MENSAJE_SESION_NO_ENCONTRADA)
          return
        }
        if (error.status !== 0 && error.message !== '') {
          setErrorVoto(error.message)
          return
        }
      }
      setErrorVoto(MENSAJE_ERROR_VOTO)
    } finally {
      setEnviandoVoto(false)
    }
  }

  const sesionActiva = estado.tipo === 'listo' ? estado.sesiones.find(esSesionActiva) : undefined
  const detalle = estado.tipo === 'listo' ? estado.detalle : null
  const debeActualizar = proyectoId !== null && sesionActiva !== undefined && detalle?.estado === 'Votando'
  const sesionActivaId = sesionActiva?.id

  useEffect(() => {
    if (!debeActualizar || proyectoId === null || sesionActivaId === undefined) {
      return
    }

    const temporizadores = new Set<number>()
    const controlador = new AbortController()

    function programar() {
      const temporizador = window.setTimeout(() => {
        temporizadores.delete(temporizador)
        void actualizar()
      }, 3000)
      temporizadores.add(temporizador)
    }

    async function actualizar() {
      try {
        const nuevoDetalle = await obtenerDetalleSesionPoker(
          proyectoId as number,
          sesionActivaId as number,
          controlador.signal,
        )
        if (controlador.signal.aborted) {
          return
        }
        setConsulta((actual) => {
          if (!actual || actual.estado.tipo !== 'listo') {
            return actual
          }
          return {
            clave: actual.clave,
            estado: { ...actual.estado, detalle: nuevoDetalle },
          }
        })
        setErrorPolling(null)
      } catch (error) {
        if (
          controlador.signal.aborted ||
          (error instanceof Error && error.name === 'AbortError')
        ) {
          return
        }
        setErrorPolling(MENSAJE_ERROR_POLLING)
      } finally {
        if (!controlador.signal.aborted) {
          programar()
        }
      }
    }

    programar()

    return () => {
      controlador.abort()
      for (const temporizador of temporizadores) {
        window.clearTimeout(temporizador)
      }
    }
  }, [debeActualizar, proyectoId, sesionActivaId])

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
  const ultimaSesion = sesiones[0]
  const estadoVisible = detalle?.estado ?? sesionActiva?.estado
  const rondaVisible = detalle?.ronda_actual ?? sesionActiva?.ronda_actual

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

      {errorPolling && (
        <div className="mb-5">
          <Alert tono="info">{errorPolling}</Alert>
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
                  {estadoVisible}
                </span>
              </dd>
            </div>
            <div className="flex items-center gap-2">
              <dt className="text-ink-soft">Ronda</dt>
              <dd className="font-medium text-ink">Ronda {rondaVisible}</dd>
            </div>
          </dl>
          <p className="mt-4 text-sm text-ink-soft">
            {estadoVisible === 'Votando'
              ? 'El equipo está votando. Los votos se revelan al finalizar la ronda.'
              : estadoVisible === 'Revelada'
                ? 'Los votos fueron revelados. La sesión sigue activa.'
                : estadoVisible === 'Cerrada'
                  ? 'La sesión fue cerrada.'
                  : 'La sesión fue cancelada.'}
          </p>

          {detalle && (
            <div className="mt-5 border-t border-border pt-5">
              <div className="flex flex-wrap items-center justify-between gap-3 text-sm">
                <p className="font-medium text-ink" aria-live="polite">
                  {detalle.votos_registrados} de {detalle.total_miembros} votaron
                </p>
                <p className="text-ink-soft" aria-live="polite">
                  {detalle.mi_voto === null ? 'Todavía no votaste' : `Tu voto: ${detalle.mi_voto}`}
                </p>
              </div>

              {errorVoto && (
                <div className="mt-4">
                  <Alert tono="error">{errorVoto}</Alert>
                </div>
              )}

              {avisoVoto && (
                <div className="mt-4">
                  <Alert tono="info">{avisoVoto}</Alert>
                </div>
              )}

              <div className="mt-4">
                <CartasVoto
                  valor={detalle.mi_voto}
                  onSeleccionar={(puntos) => void votar(puntos)}
                  deshabilitado={detalle.estado !== 'Votando' || enviandoVoto}
                />
              </div>
            </div>
          )}
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
