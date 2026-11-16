import { useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { listarSesionesActivas } from '../../../api/poker'
import type { SesionPokerActiva } from '../../../api/types'
import { useSesion } from '../../../session/sesion'
import ToastSesionPoker from './ToastSesionPoker'
import {
  DURACION_TOAST_MS,
  INTERVALO_POLLING_MS,
  SesionesPokerContext,
} from './sesionesPoker'

type ProveedorSesionesPokerProps = {
  proyectoId: number | null
  children: ReactNode
}

export function ProveedorSesionesPoker({
  proyectoId,
  children,
}: ProveedorSesionesPokerProps) {
  const { usuarioActivo } = useSesion()
  const usuarioId = usuarioActivo?.id ?? null
  const usuarioIdRef = useRef<number | null>(usuarioId)
  const sesionesVistas = useRef<Set<number>>(new Set())
  const primeraCarga = useRef(true)
  const [sesiones, setSesiones] = useState<SesionPokerActiva[]>([])
  const [toasts, setToasts] = useState<SesionPokerActiva[]>([])

  useEffect(() => {
    usuarioIdRef.current = usuarioId
  }, [usuarioId])

  useEffect(() => {
    if (proyectoId === null) {
      return
    }

    const idProyecto = proyectoId
    const consulta: { vigente: boolean } = { vigente: true }
    const reloj: { temporizador: number | undefined } = { temporizador: undefined }
    const solicitud: { controlador: AbortController | null } = { controlador: null }

    function aplicarSesiones(activas: readonly SesionPokerActiva[]) {
      setSesiones([...activas])

      const esPrimeraCarga = primeraCarga.current
      const nuevas = activas.filter(
        (sesion) => !sesionesVistas.current.has(sesion.sesion_id),
      )
      for (const sesion of activas) {
        sesionesVistas.current.add(sesion.sesion_id)
      }
      primeraCarga.current = false

      if (esPrimeraCarga) {
        return
      }

      const paraNotificar = nuevas.filter(
        (sesion) =>
          sesion.estado === 'Votando' && sesion.creado_por.id !== usuarioIdRef.current,
      )
      if (paraNotificar.length > 0) {
        setToasts((actuales) => [...actuales, ...paraNotificar])
      }
    }

    function programar() {
      if (reloj.temporizador !== undefined) {
        window.clearTimeout(reloj.temporizador)
        reloj.temporizador = undefined
      }
      if (!consulta.vigente || document.hidden) {
        return
      }
      reloj.temporizador = window.setTimeout(() => {
        void consultar()
      }, INTERVALO_POLLING_MS)
    }

    async function consultar() {
      if (!consulta.vigente || document.hidden) {
        return
      }

      solicitud.controlador?.abort()
      const actual = new AbortController()
      solicitud.controlador = actual

      try {
        const activas = await listarSesionesActivas(idProyecto, actual.signal)
        if (!consulta.vigente || actual.signal.aborted) {
          return
        }
        aplicarSesiones(activas)
      } catch {
        // Se conserva la última lista; el próximo poll reintenta.
      } finally {
        if (consulta.vigente && solicitud.controlador === actual) {
          programar()
        }
      }
    }

    function alCambiarVisibilidad() {
      if (document.hidden) {
        if (reloj.temporizador !== undefined) {
          window.clearTimeout(reloj.temporizador)
          reloj.temporizador = undefined
        }
        return
      }
      void consultar()
    }

    document.addEventListener('visibilitychange', alCambiarVisibilidad)
    void consultar()

    return () => {
      consulta.vigente = false
      solicitud.controlador?.abort()
      if (reloj.temporizador !== undefined) {
        window.clearTimeout(reloj.temporizador)
      }
      document.removeEventListener('visibilitychange', alCambiarVisibilidad)
    }
  }, [proyectoId])

  const toastActual = toasts[0]

  useEffect(() => {
    if (!toastActual) {
      return
    }
    const temporizador = window.setTimeout(() => {
      setToasts((actuales) => actuales.slice(1))
    }, DURACION_TOAST_MS)
    return () => window.clearTimeout(temporizador)
  }, [toastActual])

  const valor = useMemo(
    () => ({
      proyectoId,
      sesiones,
      historiasConPoker: new Set(sesiones.map((sesion) => sesion.historia.id)),
    }),
    [proyectoId, sesiones],
  )

  return (
    <SesionesPokerContext.Provider value={valor}>
      {children}
      {toastActual && <ToastSesionPoker sesion={toastActual} />}
    </SesionesPokerContext.Provider>
  )
}
