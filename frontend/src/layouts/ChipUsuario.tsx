import { useEffect, useId, useRef, useState } from 'react'
import type { RolProyecto, Usuario } from '../api/types'

const ETIQUETA_ROL: Record<RolProyecto, string> = {
  LIDER: 'Líder',
  MIEMBRO: 'Miembro',
}

type ChipUsuarioProps = {
  usuario: Usuario
  /** Rol del usuario en el proyecto abierto; sin proyecto o sin rol, solo se muestra el nombre. */
  rol?: RolProyecto
  onSalir: () => void
}

export default function ChipUsuario({ usuario, rol, onSalir }: ChipUsuarioProps) {
  const [abierto, setAbierto] = useState(false)
  const raiz = useRef<HTMLDivElement>(null)
  const menuId = useId()

  useEffect(() => {
    if (!abierto) {
      return
    }
    function alHacerClic(evento: MouseEvent) {
      if (raiz.current && !raiz.current.contains(evento.target as Node)) {
        setAbierto(false)
      }
    }
    function alPresionarTecla(evento: KeyboardEvent) {
      if (evento.key === 'Escape') {
        setAbierto(false)
      }
    }
    document.addEventListener('mousedown', alHacerClic)
    document.addEventListener('keydown', alPresionarTecla)
    return () => {
      document.removeEventListener('mousedown', alHacerClic)
      document.removeEventListener('keydown', alPresionarTecla)
    }
  }, [abierto])

  const iniciales = usuario.nombre
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((parte) => parte[0])
    .join('')

  return (
    <div ref={raiz} className="relative">
      <button
        type="button"
        onClick={() => setAbierto((valor) => !valor)}
        aria-expanded={abierto}
        aria-haspopup="menu"
        aria-controls={abierto ? menuId : undefined}
        title={usuario.email}
        className="flex items-center gap-2.5 rounded-full border border-glass-border bg-glass-2 py-1 pl-1 pr-4 text-left transition-all duration-150 hover:bg-glass"
      >
        <span
          aria-hidden="true"
          className="flex h-8 w-8 items-center justify-center rounded-full bg-acento text-xs font-bold uppercase text-white"
        >
          {iniciales}
        </span>
        <span className="flex flex-col leading-tight">
          <span className="text-sm font-semibold text-ink">{usuario.nombre}</span>
          {rol && <span className="text-xs text-ink-soft">{ETIQUETA_ROL[rol]}</span>}
        </span>
        <i
          className={`bi bi-chevron-down text-xs text-ink-soft transition-transform duration-150 ${abierto ? 'rotate-180' : ''}`}
          aria-hidden="true"
        />
      </button>
      {abierto && (
        <div
          id={menuId}
          role="menu"
          className="glass absolute right-0 top-full z-30 mt-2 w-64 rounded-2xl bg-[var(--page-end)] p-2"
        >
          <div className="px-3 py-2">
            <p className="truncate text-sm font-semibold text-ink">{usuario.nombre}</p>
            <p className="truncate text-xs text-ink-soft">{usuario.email}</p>
          </div>
          <button
            type="button"
            role="menuitem"
            onClick={() => {
              setAbierto(false)
              onSalir()
            }}
            className="flex w-full items-center gap-2 rounded-xl px-3 py-2 text-sm font-medium text-ink transition-colors duration-150 hover:bg-primary-100"
          >
            <i className="bi bi-box-arrow-right" aria-hidden="true" />
            Salir
          </button>
        </div>
      )}
    </div>
  )
}
