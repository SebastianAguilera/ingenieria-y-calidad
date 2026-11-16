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
      className={`glass rounded-2xl p-6 ${className}`.trim()}
    >
      {tieneEncabezado && (
        <header className="mb-5">
          {title && <h2 className="text-xl font-bold text-ink">{title}</h2>}
          {description && <p className="mt-1 text-sm text-ink-soft">{description}</p>}
        </header>
      )}
      {children}
    </section>
  )
}
