import type { ReactNode } from 'react'

export type TonoAlerta = 'error' | 'info' | 'success'

type AlertProps = {
  tono?: TonoAlerta
  children: ReactNode
}

const ESTILOS: Record<TonoAlerta, string> = {
  error: 'border-[var(--danger)]/50 bg-[color-mix(in_oklab,var(--danger)_10%,transparent)] text-ink',
  info: 'border-glass-border bg-[var(--primary-100)] text-ink',
  success: 'border-[var(--success)]/50 bg-[color-mix(in_oklab,var(--success)_10%,transparent)] text-ink',
}

export default function Alert({ tono = 'error', children }: AlertProps) {
  return (
    <div
      role={tono === 'error' ? 'alert' : 'status'}
      className={`rounded-2xl border px-4 py-3 text-sm backdrop-blur-md ${ESTILOS[tono]}`}
    >
      {children}
    </div>
  )
}
