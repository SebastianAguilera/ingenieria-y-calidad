export function claseEnlaceDeMenu(
  { isActive }: { isActive: boolean },
  compacto = false,
): string {
  const base = compacto
    ? 'inline-flex items-center justify-center rounded-pill p-2 text-sm font-medium'
    : 'inline-flex items-center gap-2 rounded-pill px-3 py-2 text-sm font-medium'
  return isActive
    ? `${base} bg-primary-100 text-ink [&_.bi]:text-primary`
    : `${base} text-ink-soft transition hover:bg-primary-100 hover:text-ink [&_.bi]:transition-colors hover:[&_.bi]:text-primary`
}
