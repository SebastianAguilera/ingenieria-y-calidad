import type { ReactNode } from 'react'

export type TonoAlerta = 'error' | 'info' | 'success'

type AlertProps = {
  tono?: TonoAlerta
  children: ReactNode
}

const ESTILOS: Record<TonoAlerta, string> = {
  error: 'border-[var(--danger)] bg-[var(--accent-100)] text-ink',
  info: 'border-border bg-[var(--primary-100)] text-ink',
  success: 'border-[var(--success)] bg-[var(--primary-100)] text-ink',
}

export default function Alert({ tono = 'error', children }: AlertProps) {
  return (
    <div
      role={tono === 'error' ? 'alert' : 'status'}
      className={`rounded-tile border px-4 py-3 text-sm ${ESTILOS[tono]}`}
    >
      {children}
    </div>
  )
}
