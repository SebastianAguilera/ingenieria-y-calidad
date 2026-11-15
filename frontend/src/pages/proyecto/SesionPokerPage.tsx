import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { esApiError } from '../../api/errors'
import { listarHistorias } from '../../api/historias'
import {
  abrirNuevaRondaSesionPoker,
  acordarSesionPoker,
  cancelarSesionPoker,
  crearSesionPoker,
  listarSesionesPoker,
  obtenerDetalleSesionPoker,
  revelarSesionPoker,
  votarSesionPoker,
} from '../../api/poker'
import type {
  DetalleSesionPoker,
  EstadoSesionPoker,
  Historia,
  SesionPokerCreada,
  SesionPokerHistorial,
} from '../../api/types'
import PageHeader from '../../components/PageHeader'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalConfirmacion from '../../ui/ModalConfirmacion'
import CartasVoto from './poker/CartasVoto'
import HistorialRondas from './poker/HistorialRondas'
import ModalAcuerdo from './poker/ModalAcuerdo'
import PanelDispersion from './poker/PanelDispersion'
import { sugerirAcuerdo } from './poker/escala'

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
const MENSAJE_SIN_VOTOS = 'No se puede revelar sin votos.'
const MENSAJE_ERROR_REVELAR = 'No se pudieron revelar los votos. Intentá nuevamente.'
const MENSAJE_ERROR_NUEVA_RONDA = 'No se pudo abrir la nueva ronda. Intentá nuevamente.'
const MENSAJE_SESION_NO_REVELADA = 'La sesión ya no está revelada.'
const MENSAJE_HISTORIA_NO_ENCONTRADA =
  'La historia no existe o no pertenece a este proyecto.'
const MENSAJE_PROYECTO_INEXISTENTE = 'No tenés acceso a este proyecto o no existe'
const MENSAJE_SESION_NO_ENCONTRADA = 'La sesión de Planning Poker no existe.'
const MENSAJE_HISTORIA_NO_ESTIMABLE =
  'La historia está en estado Hecho y no se puede estimar.'
const MENSAJE_SESION_ACTIVA = 'Ya hay una sesión activa. Te unimos a ella.'
const MENSAJE_PUNTOS_INVALIDOS = 'El valor elegido no es válido. Elegí otra carta.'
const MENSAJE_PUNTOS_ACUERDO_INVALIDOS =
  'El valor elegido no es válido. Elegí otro de la escala.'
const MENSAJE_SESION_YA_CERRADA = 'La sesión ya fue cerrada o cancelada.'
const MENSAJE_ERROR_ACUERDO = 'No se pudo registrar el acuerdo. Intentá nuevamente.'
const MENSAJE_ERROR_CANCELAR = 'No se pudo cancelar la sesión. Intentá nuevamente.'

const formatoPuntos = new Intl.NumberFormat('es-AR', {
  maximumFractionDigits: 2,
})

function etiquetaPuntos(puntos: number): string {
  return `${formatoPuntos.format(puntos)} ${puntos === 1 ? 'punto' : 'puntos'}`
}

function marcaDeVoto(puntos: number, minimo: number, maximo: number): string | null {
  const esMinimo = puntos === minimo
  const esMaximo = puntos === maximo
  if (esMinimo && esMaximo) {
    return 'Mínimo y máximo'
  }
  if (esMinimo) {
    return 'Mínimo'
  }
  if (esMaximo) {
    return 'Máximo'
  }
  return null
}

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

const ESTILO_ESTADO: Record<EstadoSesionPoker, string> = {
  Votando: 'bg-[var(--accent-100)] text-[var(--accent-600)]',
  Revelada: 'bg-[var(--primary-100)] text-primary',
  Cerrada: 'bg-[var(--primary-100)] text-[var(--success)]',
  Cancelada: 'bg-[var(--surface-2)] text-ink-soft',
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
      <div className="flex items-center gap-4 rounded-tile border border-border bg-[var(--primary-100)] p-4">
        <span
          aria-hidden="true"
          className="flex h-11 w-11 shrink-0 items-center justify-center rounded-tile bg-surface text-xl text-[var(--success)]"
        >
          <i className="bi bi-check2-circle" />
        </span>
        {sesion.puntos_acordados === null ? (
          <p className="text-ink-soft">Sin puntos acordados</p>
        ) : (
          <p className="text-2xl font-semibold text-ink">{`Puntos acordados: ${sesion.puntos_acordados}`}</p>
        )}
      </div>
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
  const [revelando, setRevelando] = useState(false)
  const [abriendoRonda, setAbriendoRonda] = useState(false)
  const [errorAccion, setErrorAccion] = useState<string | null>(null)
  const [acuerdoAbierto, setAcuerdoAbierto] = useState(false)
  const [acordando, setAcordando] = useState(false)
  const [errorAcuerdo, setErrorAcuerdo] = useState<string | null>(null)
  const [cancelacionAbierta, setCancelacionAbierta] = useState(false)
  const [cancelando, setCancelando] = useState(false)

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

  function aplicarSesionFinal(detalle: DetalleSesionPoker) {
    setConsulta((actual) => {
      if (!actual || actual.estado.tipo !== 'listo') {
        return actual
      }
      return {
        clave: actual.clave,
        estado: {
          ...actual.estado,
          detalle,
          sesiones: actual.estado.sesiones.map((sesion) =>
            sesion.id === detalle.id
              ? {
                  ...sesion,
                  estado: detalle.estado,
                  puntos_acordados: detalle.puntos_acordados,
                  cerrado_en: detalle.cerrado_en,
                  cancelado_en: detalle.cancelado_en,
                  story_estimate: detalle.story_estimate ?? null,
                }
              : sesion,
          ),
        },
      }
    })
  }

  async function refrescarDetalle(sesionId: number) {
    if (proyectoId === null) {
      return
    }
    try {
      const detalle = await obtenerDetalleSesionPoker(proyectoId, sesionId)
      aplicarSesionFinal(detalle)
    } catch {
      // Se conserva el detalle previo si el refresco falla.
    }
  }

  async function acordar(puntos: number) {
    if (
      estado.tipo !== 'listo' ||
      proyectoId === null ||
      estado.detalle === null ||
      estado.detalle.estado !== 'Revelada' ||
      acordando
    ) {
      return
    }

    const sesionId = estado.detalle.id
    setAcordando(true)
    setErrorAcuerdo(null)

    try {
      const nuevoDetalle = await acordarSesionPoker(proyectoId, sesionId, puntos)
      aplicarSesionFinal(nuevoDetalle)
      setAcuerdoAbierto(false)
      setErrorAccion(null)
    } catch (error) {
      if (esApiError(error)) {
        if (error.code === 'PUNTOS_POKER_INVALIDOS') {
          setErrorAcuerdo(MENSAJE_PUNTOS_ACUERDO_INVALIDOS)
          return
        }
        if (error.code === 'SESION_POKER_NO_REVELADA') {
          setAcuerdoAbierto(false)
          await refrescarDetalle(sesionId)
          setErrorAccion(MENSAJE_SESION_NO_REVELADA)
          return
        }
        if (error.code === 'SESION_POKER_CERRADA') {
          setAcuerdoAbierto(false)
          await refrescarDetalle(sesionId)
          setErrorAccion(MENSAJE_SESION_YA_CERRADA)
          return
        }
        if (error.code === 'PROYECTO_NO_ENCONTRADO') {
          setErrorAcuerdo(MENSAJE_PROYECTO_INEXISTENTE)
          return
        }
        if (error.code === 'SESION_POKER_NO_ENCONTRADA') {
          setErrorAcuerdo(MENSAJE_SESION_NO_ENCONTRADA)
          return
        }
        if (error.status !== 0 && error.message !== '') {
          setErrorAcuerdo(error.message)
          return
        }
      }
      setErrorAcuerdo(MENSAJE_ERROR_ACUERDO)
    } finally {
      setAcordando(false)
    }
  }

  async function cancelarSesion() {
    if (
      estado.tipo !== 'listo' ||
      proyectoId === null ||
      estado.detalle === null ||
      cancelando
    ) {
      return
    }

    const sesionId = estado.detalle.id
    setCancelando(true)
    setErrorAccion(null)

    try {
      const nuevoDetalle = await cancelarSesionPoker(proyectoId, sesionId)
      aplicarSesionFinal(nuevoDetalle)
      setCancelacionAbierta(false)
    } catch (error) {
      setCancelacionAbierta(false)
      if (esApiError(error)) {
        if (error.code === 'SESION_POKER_CERRADA') {
          await refrescarDetalle(sesionId)
          setErrorAccion(MENSAJE_SESION_YA_CERRADA)
          return
        }
        if (error.code === 'PROYECTO_NO_ENCONTRADO') {
          setErrorAccion(MENSAJE_PROYECTO_INEXISTENTE)
          return
        }
        if (error.code === 'SESION_POKER_NO_ENCONTRADA') {
          setErrorAccion(MENSAJE_SESION_NO_ENCONTRADA)
          return
        }
        if (error.status !== 0 && error.message !== '') {
          setErrorAccion(error.message)
          return
        }
      }
      setErrorAccion(MENSAJE_ERROR_CANCELAR)
    } finally {
      setCancelando(false)
    }
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
    setErrorAccion(null)

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

  async function revelar() {
    if (
      estado.tipo !== 'listo' ||
      proyectoId === null ||
      estado.detalle === null ||
      estado.detalle.estado !== 'Votando' ||
      revelando
    ) {
      return
    }

    const sesionId = estado.detalle.id
    setRevelando(true)
    setErrorAccion(null)

    try {
      const nuevoDetalle = await revelarSesionPoker(proyectoId, sesionId)
      actualizarDetalle(nuevoDetalle)
      setErrorVoto(null)
      setAvisoVoto(null)
    } catch (error) {
      if (esApiError(error) && error.code === 'SESION_POKER_NO_VOTANDO') {
        try {
          const detalleActualizado = await obtenerDetalleSesionPoker(proyectoId, sesionId)
          actualizarDetalle(detalleActualizado)
          setAvisoVoto(MENSAJE_VOTACION_CERRADA)
        } catch {
          setErrorAccion(MENSAJE_ERROR_REVELAR)
        }
        return
      }
      if (esApiError(error)) {
        if (error.code === 'SESION_POKER_SIN_VOTOS') {
          setErrorAccion(MENSAJE_SIN_VOTOS)
          return
        }
        if (error.code === 'PROYECTO_NO_ENCONTRADO') {
          setErrorAccion(MENSAJE_PROYECTO_INEXISTENTE)
          return
        }
        if (error.code === 'SESION_POKER_NO_ENCONTRADA') {
          setErrorAccion(MENSAJE_SESION_NO_ENCONTRADA)
          return
        }
        if (error.status !== 0 && error.message !== '') {
          setErrorAccion(error.message)
          return
        }
      }
      setErrorAccion(MENSAJE_ERROR_REVELAR)
    } finally {
      setRevelando(false)
    }
  }

  async function abrirNuevaRonda() {
    if (
      estado.tipo !== 'listo' ||
      proyectoId === null ||
      estado.detalle === null ||
      estado.detalle.estado !== 'Revelada' ||
      abriendoRonda
    ) {
      return
    }

    const sesionId = estado.detalle.id
    setAbriendoRonda(true)
    setErrorAccion(null)
    setErrorVoto(null)
    setAvisoVoto(null)

    try {
      const nuevoDetalle = await abrirNuevaRondaSesionPoker(proyectoId, sesionId)
      actualizarDetalle(nuevoDetalle)
      setErrorPolling(null)
    } catch (error) {
      if (esApiError(error) && error.code === 'SESION_POKER_NO_REVELADA') {
        try {
          const detalleActualizado = await obtenerDetalleSesionPoker(proyectoId, sesionId)
          actualizarDetalle(detalleActualizado)
          setErrorAccion(MENSAJE_SESION_NO_REVELADA)
        } catch {
          setErrorAccion(MENSAJE_ERROR_NUEVA_RONDA)
        }
        return
      }
      if (esApiError(error)) {
        if (error.code === 'PROYECTO_NO_ENCONTRADO') {
          setErrorAccion(MENSAJE_PROYECTO_INEXISTENTE)
          return
        }
        if (error.code === 'SESION_POKER_NO_ENCONTRADA') {
          setErrorAccion(MENSAJE_SESION_NO_ENCONTRADA)
          return
        }
        if (error.status !== 0 && error.message !== '') {
          setErrorAccion(error.message)
          return
        }
      }
      setErrorAccion(MENSAJE_ERROR_NUEVA_RONDA)
    } finally {
      setAbriendoRonda(false)
    }
  }

  const sesionActiva = estado.tipo === 'listo' ? estado.sesiones.find(esSesionActiva) : undefined
  const detalle = estado.tipo === 'listo' ? estado.detalle : null
  const debeActualizar =
    proyectoId !== null &&
    sesionActiva !== undefined &&
    (detalle?.estado === 'Votando' || detalle?.estado === 'Revelada')
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

      {errorAccion && (
        <div className="mb-5">
          <Alert tono="error">{errorAccion}</Alert>
        </div>
      )}

      {sesionActiva ? (
        <Card
          title="Sesión en curso"
          description={`Creada por ${sesionActiva.creado_por.nombre} el ${formatearFecha(sesionActiva.creado_en)}.`}
        >
          <div className="flex flex-wrap gap-x-8 gap-y-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-ink-soft">Estado</p>
              <span
                className={`mt-1.5 inline-flex items-center gap-2 rounded-pill px-3 py-1.5 text-xs font-semibold ${ESTILO_ESTADO[estadoVisible ?? 'Cancelada']}`}
              >
                <i className="bi bi-circle-fill text-[7px]" aria-hidden="true" />
                {estadoVisible}
              </span>
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-ink-soft">Ronda</p>
              <p className="mt-1.5 text-lg font-semibold text-ink">Ronda {rondaVisible}</p>
            </div>
          </div>
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

              <div
                role="progressbar"
                aria-label="Progreso de votos"
                aria-valuemin={0}
                aria-valuemax={detalle.total_miembros}
                aria-valuenow={detalle.votos_registrados}
                className="mt-3 h-2 overflow-hidden rounded-pill bg-[var(--surface-2)]"
              >
                <div
                  className="h-full rounded-pill bg-primary transition-[width]"
                  style={{
                    width: `${
                      detalle.total_miembros > 0
                        ? Math.round((detalle.votos_registrados / detalle.total_miembros) * 100)
                        : 0
                    }%`,
                  }}
                />
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

              {(detalle.estado === 'Votando' || detalle.estado === 'Revelada') && (
                <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-5">
                  <div className="flex flex-wrap gap-3">
                    {detalle.estado === 'Votando' && (
                      <Button
                        onClick={() => void revelar()}
                        disabled={revelando}
                        aria-busy={revelando}
                      >
                        <i className="bi bi-eye" aria-hidden="true" />
                        Revelar votos
                      </Button>
                    )}
                    {detalle.estado === 'Revelada' && (
                      <Button
                        onClick={() => {
                          setErrorAcuerdo(null)
                          setAcuerdoAbierto(true)
                        }}
                        disabled={acordando}
                      >
                        <i className="bi bi-check2-circle" aria-hidden="true" />
                        Registrar acuerdo
                      </Button>
                    )}
                    {detalle.estado === 'Revelada' && (
                      <Button
                        variante="secondary"
                        onClick={() => void abrirNuevaRonda()}
                        disabled={abriendoRonda}
                        aria-busy={abriendoRonda}
                      >
                        <i className="bi bi-arrow-repeat" aria-hidden="true" />
                        Nueva ronda
                      </Button>
                    )}
                  </div>
                  <Button
                    variante="danger"
                    onClick={() => setCancelacionAbierta(true)}
                    disabled={cancelando}
                  >
                    <i className="bi bi-slash-circle" aria-hidden="true" />
                    Cancelar sesión
                  </Button>
                </div>
              )}

              {detalle.estado === 'Cerrada' && detalle.puntos_acordados !== null && (
                <p className="mt-4 text-sm text-ink" aria-live="polite">
                  Puntos acordados: {detalle.puntos_acordados}
                </p>
              )}

              {detalle.estado === 'Revelada' && (
                <div className="mt-5 space-y-4 border-t border-border pt-5">
                  <div>
                    <h3 className="mb-3 text-base font-semibold text-ink">Votos revelados</h3>
                    <ul aria-label="Votos revelados" className="grid gap-2 sm:grid-cols-2">
                      {detalle.votos.map((voto) => {
                        const marca = detalle.dispersion
                          ? marcaDeVoto(
                              voto.puntos,
                              detalle.dispersion.minimo,
                              detalle.dispersion.maximo,
                            )
                          : null

                        return (
                          <li
                            key={voto.usuario_id}
                            className="flex items-center gap-3 rounded-tile border border-border bg-surface px-3 py-2 text-sm text-ink"
                          >
                            <span
                              aria-hidden="true"
                              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-pill bg-[var(--primary-100)] text-xs font-semibold text-primary"
                            >
                              {voto.usuario_nombre.charAt(0).toUpperCase()}
                            </span>
                            <span className="flex-1">
                              {voto.usuario_nombre}: {etiquetaPuntos(voto.puntos)}
                            </span>
                            {marca && (
                              <span className="inline-flex items-center rounded-pill border border-border bg-[var(--surface-2)] px-2.5 py-0.5 text-xs font-semibold text-ink-soft">
                                {marca}
                              </span>
                            )}
                          </li>
                        )
                      })}
                    </ul>
                  </div>

                  {detalle.dispersion && <PanelDispersion dispersion={detalle.dispersion} />}
                </div>
              )}

              <HistorialRondas
                rondas={detalle.historial_rondas}
                rondaActual={detalle.ronda_actual}
              />
            </div>
          )}
        </Card>
      ) : (
        <>
          {ultimaSesion?.estado === 'Cerrada' && <ResultadoCerrado sesion={ultimaSesion} />}

          {ultimaSesion?.estado === 'Cancelada' && (
            <div className="mb-5 flex items-center gap-4 rounded-card border border-border bg-surface p-4 shadow-card">
              <span
                aria-hidden="true"
                className="flex h-11 w-11 shrink-0 items-center justify-center rounded-tile bg-[var(--accent-100)] text-xl text-[var(--warning)]"
              >
                <i className="bi bi-slash-circle" />
              </span>
              <p className="text-ink">La última sesión fue cancelada.</p>
            </div>
          )}

          <Card
            title="Iniciar Planning Poker"
            description="La sesión se crea recién cuando confirmás."
          >
            <div className="flex flex-wrap gap-x-8 gap-y-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-ink-soft">
                  Estado de la historia
                </p>
                <p className="mt-1.5 text-lg font-semibold text-ink">{historia.estado}</p>
              </div>
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-ink-soft">
                  Estimación
                </p>
                <p className="mt-1.5 text-lg font-semibold text-ink">
                  {historia.story_points ?? '–'}
                </p>
              </div>
            </div>
            {ultimaSesion === undefined && (
              <p className="mt-4 text-sm text-ink-soft">
                Todavía no hay sesiones de Planning Poker para esta historia.
              </p>
            )}
            <div className="mt-5">
              <Button onClick={() => void iniciar()} disabled={creando} aria-busy={creando}>
                <i className="bi bi-play-fill" aria-hidden="true" />
                Iniciar Planning Poker
              </Button>
            </div>
          </Card>
        </>
      )}

      {acuerdoAbierto && detalle && (
        <ModalAcuerdo
          valorInicial={sugerirAcuerdo(detalle.dispersion)}
          procesando={acordando}
          error={errorAcuerdo}
          onConfirmar={(puntos) => void acordar(puntos)}
          onCerrar={() => {
            setAcuerdoAbierto(false)
            setErrorAcuerdo(null)
          }}
        />
      )}

      {cancelacionAbierta && (
        <ModalConfirmacion
          titulo="Cancelar sesión"
          textoConfirmar="Sí, cancelar"
          textoProcesando="Cancelando…"
          procesando={cancelando}
          onConfirmar={() => void cancelarSesion()}
          onCancelar={() => setCancelacionAbierta(false)}
        >
          <p>¿Querés cancelar la sesión? No se estimará la historia y podrás iniciar otra.</p>
        </ModalConfirmacion>
      )}
    </section>
  )
}
