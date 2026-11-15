import type { ReactNode } from 'react'

type CardProps = {
  title?: string
  description?: string
  children: ReactNode
  className?: string
}

export default function Card({ title, description, children, className = '' }: CardProps) {
  const tieneEncabezado = Boolean(title || description)

  return (
    <section
      className={`rounded-card border border-border bg-surface p-6 shadow-card ${className}`.trim()}
    >
      {tieneEncabezado && (
        <header className="mb-5">
          {title && <h2 className="text-2xl font-semibold text-ink">{title}</h2>}
          {description && <p className="mt-1 text-sm text-ink-soft">{description}</p>}
        </header>
      )}
      {children}
    </section>
  )
}
