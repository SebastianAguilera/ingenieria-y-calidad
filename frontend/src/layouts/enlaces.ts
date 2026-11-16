export function claseEnlaceDeMenu(
  { isActive }: { isActive: boolean },
  compacto = false,
): string {
  const base = compacto
    ? 'inline-flex items-center justify-center rounded-xl p-2.5 text-sm font-medium transition-all duration-150'
    : 'inline-flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-sm font-medium transition-all duration-150'
  return isActive
    ? `${base} bg-acento brillo-acento`
    : `${base} text-ink-soft hover:bg-primary-100 hover:text-ink`
}
