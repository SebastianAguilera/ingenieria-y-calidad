import type { ComponentPropsWithRef } from 'react'

export type VarianteBoton = 'primary' | 'secondary' | 'ghost' | 'success' | 'danger'
export type TamanoBoton = 'md' | 'sm'

type ButtonProps = ComponentPropsWithRef<'button'> & {
  variante?: VarianteBoton
  tamano?: TamanoBoton
}

const ESTILOS: Record<VarianteBoton, string> = {
  primary: 'bg-primary text-[var(--on-primary)] hover:bg-[var(--primary-700)]',
  secondary: 'bg-[var(--primary-100)] text-primary hover:bg-[var(--primary-300)]',
  ghost: 'bg-transparent text-primary hover:bg-[var(--primary-100)]',
  success: 'border border-[var(--success)] bg-[var(--primary-100)] text-[var(--success)] hover:opacity-90',
  danger: 'border border-[var(--danger)] bg-[var(--accent-100)] text-[var(--danger)] hover:opacity-90',
}

const TAMANOS: Record<TamanoBoton, string> = {
  md: 'px-5 py-2.5 text-sm',
  sm: 'px-3 py-1.5 text-xs',
}

export default function Button({
  variante = 'primary',
  tamano = 'md',
  type = 'button',
  className = '',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={`inline-flex items-center justify-center gap-2 rounded-pill font-semibold transition focus-visible:shadow-[var(--ring)] disabled:cursor-not-allowed disabled:opacity-60 ${TAMANOS[tamano]} ${ESTILOS[variante]} ${className}`.trim()}
      {...props}
    />
  )
}
