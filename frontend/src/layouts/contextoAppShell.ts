/**
 * Contexto que `AppShell` entrega a las páginas por `<Outlet context>`.
 * `slotAcciones` es un hueco en la barra superior donde una página puede
 * montar controles propios (p. ej. la campana de Planning Poker) con un portal.
 */
export type ContextoAppShell = {
  slotAcciones: HTMLElement | null
}
