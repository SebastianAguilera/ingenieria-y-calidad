import type { ButtonHTMLAttributes } from 'react'

export type VarianteBoton = 'primary' | 'secondary' | 'ghost' | 'success' | 'danger'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variante?: VarianteBoton
}

const ESTILOS: Record<VarianteBoton, string> = {
  primary: 'bg-primary text-[var(--on-primary)] hover:bg-[var(--primary-700)]',
  secondary: 'bg-[var(--primary-100)] text-primary hover:bg-[var(--primary-300)]',
  ghost: 'bg-transparent text-primary hover:bg-[var(--primary-100)]',
  success: 'bg-[var(--success)] text-white hover:opacity-90',
  danger: 'bg-[var(--danger)] text-white hover:opacity-90',
}

export default function Button({
  variante = 'primary',
  type = 'button',
  className = '',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={`inline-flex items-center justify-center gap-2 rounded-pill px-5 py-2.5 text-sm font-semibold transition focus-visible:shadow-[var(--ring)] disabled:cursor-not-allowed disabled:opacity-60 ${ESTILOS[variante]} ${className}`.trim()}
      {...props}
    />
  )
}
