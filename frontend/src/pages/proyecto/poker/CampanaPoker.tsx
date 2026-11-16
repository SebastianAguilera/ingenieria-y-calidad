import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ETIQUETA_CAMPANA,
  ETIQUETA_LISTA,
  ID_LISTA_SESIONES,
  MENSAJE_SIN_SESIONES,
  useSesionesPoker,
} from './sesionesPoker'

function etiquetaBadge(cantidad: number): string {
  return cantidad === 1 ? '1 sesión sin votar' : `${cantidad} sesiones sin votar`
}

export default function CampanaPoker() {
  const { proyectoId, sesiones } = useSesionesPoker()
  const navegar = useNavigate()
  const [abierta, setAbierta] = useState(false)
  const contenedorRef = useRef<HTMLDivElement>(null)
  const botonRef = useRef<HTMLButtonElement>(null)
  const primerItemRef = useRef<HTMLButtonElement>(null)
  const pendientes = sesiones.filter(
    (sesion) => sesion.estado === 'Votando' && !sesion.ya_vote,
  ).length

  useEffect(() => {
    if (!abierta) {
      return
    }

    primerItemRef.current?.focus()

    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape') {
        setAbierta(false)
        botonRef.current?.focus()
      }
    }

    function alHacerClic(evento: MouseEvent) {
      if (
        contenedorRef.current &&
        !contenedorRef.current.contains(evento.target as Node)
      ) {
        setAbierta(false)
      }
    }

    document.addEventListener('keydown', alPresionarTecla)
    document.addEventListener('mousedown', alHacerClic)
    return () => {
      document.removeEventListener('keydown', alPresionarTecla)
      document.removeEventListener('mousedown', alHacerClic)
    }
  }, [abierta])

  function abrirSesion(historiaId: number) {
    setAbierta(false)
    if (proyectoId === null) {
      return
    }
    navegar(`/proyectos/${proyectoId}/backlog/${historiaId}/poker`)
  }

  return (
    <div ref={contenedorRef} className="relative flex justify-end">
      <button
        ref={botonRef}
        type="button"
        aria-label={ETIQUETA_CAMPANA}
        aria-expanded={abierta}
        aria-controls={ID_LISTA_SESIONES}
        onClick={() => setAbierta((actual) => !actual)}
        className="relative inline-flex h-10 w-10 items-center justify-center rounded-pill border border-border bg-surface text-ink shadow-card transition hover:bg-[var(--primary-100)] focus-visible:shadow-[var(--ring)]"
      >
        <i className="bi bi-bell text-lg" aria-hidden="true" />
        {pendientes > 0 && (
          <span
            aria-label={etiquetaBadge(pendientes)}
            className="absolute -right-1 -top-1 inline-flex h-5 min-w-5 items-center justify-center rounded-pill bg-[var(--danger)] px-1 text-xs font-bold text-white"
          >
            {pendientes}
          </span>
        )}
      </button>

      {abierta && (
        <div
          id={ID_LISTA_SESIONES}
          role="region"
          aria-label={ETIQUETA_LISTA}
          className="flotante absolute right-0 top-12 z-40 w-80 p-1"
        >
          {sesiones.length === 0 ? (
            <p className="px-3 py-2 text-sm text-ink-soft">{MENSAJE_SIN_SESIONES}</p>
          ) : (
            <ul className="flex flex-col gap-1">
              {sesiones.map((sesion, indice) => (
                <li key={sesion.sesion_id}>
                  <button
                    ref={indice === 0 ? primerItemRef : undefined}
                    type="button"
                    onClick={() => abrirSesion(sesion.historia.id)}
                    className="w-full rounded-tile px-3 py-2 text-left text-sm text-ink transition hover:bg-[var(--primary-100)] focus-visible:shadow-[var(--ring)]"
                  >
                    <span className="flex items-center justify-between gap-2">
                      <span className="font-semibold text-primary">{sesion.historia.codigo}</span>
                      <span className="inline-flex items-center rounded-pill border border-border bg-[var(--surface-2)] px-2 py-0.5 text-xs font-semibold text-ink-soft">
                        {sesion.estado}
                      </span>
                    </span>
                    <span className="mt-0.5 block font-medium">{sesion.historia.titulo}</span>
                    <span className="mt-1 block text-xs text-ink-soft">
                      Ronda {sesion.ronda_actual} · Creada por {sesion.creado_por.nombre}
                    </span>
                    {sesion.estado === 'Votando' && (
                      <span className="mt-1 block text-xs font-semibold text-ink-soft">
                        {sesion.ya_vote ? 'Ya votaste' : 'Falta tu voto'}
                      </span>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
