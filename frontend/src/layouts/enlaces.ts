export function claseEnlaceDeMenu({ isActive }: { isActive: boolean }): string {
  return isActive
    ? 'inline-flex items-center gap-2 rounded-pill bg-primary-100 px-3 py-2 text-sm font-medium text-ink'
    : 'inline-flex items-center gap-2 rounded-pill px-3 py-2 text-sm font-medium text-ink-soft transition hover:bg-primary-100 hover:text-ink'
}
