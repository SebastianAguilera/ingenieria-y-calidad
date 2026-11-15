export type ResultadoOperacion =
  | { ok: true }
  | { ok: false; mensaje: string; campo?: string }
