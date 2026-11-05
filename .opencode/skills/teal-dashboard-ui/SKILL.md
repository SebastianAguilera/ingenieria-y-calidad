---
name: teal-dashboard-ui
description: Sistema de diseño de dashboard limpio para un frontend tipo GitHub Projects (sidebar oscura con avatar, fondo azul claro, tarjetas KPI, gráficos donut, área y barras horizontales, buscador y modo claro/oscuro). Tema editable por el usuario con 3 colores (primary, secondary, accent) y 2 tipografías (principal y secundaria). Úsalo al crear o modificar cualquier pantalla, componente, gráfico o el editor de tema.
---

# Teal Dashboard UI

Estilo de referencia: panel de administración sereno y legible. Sidebar de color primario con avatar y nombre, fondo claro azulado, tarjetas casi blancas con borde fino, fila de KPIs con iconos en cajita, un KPI de alerta resaltado en el color de acento, gráficos suaves (donut, áreas con tooltips en etiqueta, barras horizontales con valor al final), buscador en píldora, interruptor claro/oscuro y campana de notificaciones.

Es una **inspiración de estilo**. No copies logos, ilustraciones ni textos de la captura de referencia; usa iconos de una librería abierta (Lucide) y contenido propio del producto.

## 1. Regla de oro: solo 3 colores y 2 fuentes

El usuario controla únicamente:

| Token | Rol | Valor inicial |
|---|---|---|
| `--c-primary` | Sidebar, botones principales, series principales de gráficos, iconos | `#2c5f78` |
| `--c-secondary` | Fondo de página y superficies suaves | `#d3e5ee` |
| `--c-accent` | Alertas, KPI resaltado, serie secundaria de gráficos, etiquetas de máximo | `#e08a5a` |
| `--font-primary` | Texto de interfaz, tablas, etiquetas | `Nunito Sans` |
| `--font-secondary` | Títulos, números grandes de KPI | `Lexend` |

**Prohibido** escribir hex, `rgb()` o nombres de fuente sueltos en componentes o en la configuración de los gráficos. Todo color sale de los tokens derivados; toda fuente, de `--font-primary` o `--font-secondary`. Si necesitas un tono nuevo, derívalo con `color-mix()` o sintaxis de color relativo en `theme.css`. No existe un cuarto color.

El **modo oscuro** también se deriva de esos mismos 3 colores (sección 2). No pidas colores adicionales al usuario.

## 2. Tokens derivados (`src/theme/theme.css`)

```css
:root {
  /* Entrada del usuario */
  --c-primary: #2c5f78;
  --c-secondary: #d3e5ee;
  --c-accent: #e08a5a;
  --font-primary: "Nunito Sans", system-ui, sans-serif;
  --font-secondary: "Lexend", system-ui, sans-serif;

  /* Texto sobre color (lo recalcula theme.ts según contraste) */
  --on-primary: #ffffff;
  --on-accent: #2b1608;

  /* Fondos y superficies (claro) */
  --page:        color-mix(in oklab, var(--c-secondary) 70%, white);
  --bg:          color-mix(in oklab, var(--c-secondary) 45%, white);
  --surface:     color-mix(in oklab, var(--c-secondary) 12%, white);
  --surface-2:   color-mix(in oklab, var(--c-secondary) 40%, white);
  --border:      color-mix(in oklab, var(--c-secondary) 70%, var(--c-primary) 18%);
  --border-strong: color-mix(in oklab, var(--c-secondary) 45%, var(--c-primary));

  /* Texto */
  --ink:      color-mix(in oklab, var(--c-primary) 40%, black);
  --ink-soft: color-mix(in oklab, var(--ink) 62%, var(--surface));

  /* Primario */
  --primary-700: color-mix(in oklab, var(--c-primary) 80%, black);
  --primary-500: var(--c-primary);
  --primary-300: color-mix(in oklab, var(--c-primary) 45%, var(--surface));
  --primary-100: color-mix(in oklab, var(--c-primary) 14%, var(--surface));
  --sidebar-bg:  var(--c-primary);
  --sidebar-active-bg: var(--bg);
  --sidebar-active-ink: var(--ink);

  /* Acento */
  --accent-600: color-mix(in oklab, var(--c-accent) 82%, black);
  --accent-500: var(--c-accent);
  --accent-300: color-mix(in oklab, var(--c-accent) 45%, var(--surface));
  --accent-100: color-mix(in oklab, var(--c-accent) 18%, var(--surface));

  /* Estados (derivados, no son colores nuevos) */
  --success: color-mix(in oklab, var(--c-primary) 60%, #2e9e5b);
  --warning: var(--c-accent);
  --danger:  color-mix(in oklab, var(--c-accent) 50%, #c0392b);

  /* Paleta de gráficos: series 1 y 2 son primary y accent; el resto rota el matiz */
  --chart-1: var(--primary-500);
  --chart-2: var(--accent-500);
  --chart-3: var(--primary-300);
  --chart-4: oklch(from var(--c-primary) l c calc(h + 70));
  --chart-5: oklch(from var(--c-accent) l c calc(h - 70));
  --chart-grid: color-mix(in oklab, var(--ink) 30%, transparent);

  /* Forma */
  --radius-card: 14px;
  --radius-tile: 12px;
  --radius-pill: 999px;
  --shadow-card: 0 1px 2px color-mix(in oklab, var(--c-primary) 12%, transparent);
  --ring: 0 0 0 3px color-mix(in oklab, var(--c-accent) 45%, transparent);
}

/* Modo oscuro derivado de los mismos 3 colores */
:root[data-mode="dark"] {
  --page:      color-mix(in oklab, var(--c-primary) 26%, black);
  --bg:        color-mix(in oklab, var(--c-primary) 34%, black);
  --surface:   color-mix(in oklab, var(--c-primary) 42%, black);
  --surface-2: color-mix(in oklab, var(--c-primary) 50%, black);
  --border:    color-mix(in oklab, var(--c-primary) 62%, black);
  --border-strong: color-mix(in oklab, var(--c-primary) 78%, white);
  --ink:       color-mix(in oklab, var(--c-secondary) 85%, white);
  --ink-soft:  color-mix(in oklab, var(--ink) 65%, var(--surface));
  --sidebar-bg: color-mix(in oklab, var(--c-primary) 70%, black);
  --sidebar-active-bg: var(--surface);
  --sidebar-active-ink: var(--ink);
  --primary-100: color-mix(in oklab, var(--c-primary) 35%, var(--surface));
  --accent-100:  color-mix(in oklab, var(--c-accent) 22%, var(--surface));
  --chart-1: color-mix(in oklab, var(--c-primary) 55%, white);
  --chart-3: color-mix(in oklab, var(--c-primary) 35%, white);
  --chart-grid: color-mix(in oklab, var(--ink) 25%, transparent);
}

body {
  background: var(--page);
  color: var(--ink);
  font-family: var(--font-primary);
  font-size: 14px;
  line-height: 1.5;
}
h1, h2, h3, .display, .kpi-value { font-family: var(--font-secondary); color: var(--ink); }
:focus-visible { outline: none; box-shadow: var(--ring); }
@media (prefers-reduced-motion: reduce) { * { animation: none !important; transition: none !important; } }
```

Si usas Tailwind v4, expón los tokens con `@theme inline { --color-surface: var(--surface); --color-ink: var(--ink); --color-primary: var(--primary-500); --color-accent: var(--accent-500); --font-sans: var(--font-primary); --font-display: var(--font-secondary); }` y úsalos como `bg-surface`, `text-ink`, `font-display`.

Los navegadores sin color relativo (`oklch(from …)`) deben caer a `--chart-4: var(--primary-700)` y `--chart-5: var(--accent-600)` mediante `@supports not (color: oklch(from red l c h))`.

## 3. Motor de tema (`src/theme/theme.ts`)

```ts
export type Mode = "light" | "dark" | "system";
export type Theme = {
  primary: string; secondary: string; accent: string;
  fontPrimary: string; fontSecondary: string; mode: Mode;
};

export const DEFAULT_THEME: Theme = {
  primary: "#2c5f78", secondary: "#d3e5ee", accent: "#e08a5a",
  fontPrimary: "Nunito Sans", fontSecondary: "Lexend", mode: "system",
};

const KEY = "app.theme.v1";
const HEX = /^#[0-9a-f]{6}$/i;

function luminance(hex: string) {
  const [r, g, b] = [1, 3, 5].map(i => parseInt(hex.slice(i, i + 2), 16) / 255)
    .map(c => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}
export const contrast = (a: string, b: string) => {
  const [l1, l2] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (l1 + 0.05) / (l2 + 0.05);
};
export const readableOn = (bg: string) =>
  contrast(bg, "#ffffff") >= contrast(bg, "#1b1b1b") ? "#ffffff" : "#1b1b1b";

export function loadFont(family: string) {
  if (!family || document.querySelector(`link[data-font="${family}"]`)) return;
  const link = document.createElement("link");
  link.rel = "stylesheet";
  link.dataset.font = family;
  link.href = `https://fonts.googleapis.com/css2?family=${encodeURIComponent(family)}:wght@400;500;600;700&display=swap`;
  document.head.appendChild(link);
}

export function resolveMode(m: Mode): "light" | "dark" {
  return m === "system" ? (matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light") : m;
}

export function applyTheme(t: Theme) {
  const root = document.documentElement, s = root.style;
  if (HEX.test(t.primary)) s.setProperty("--c-primary", t.primary);
  if (HEX.test(t.secondary)) s.setProperty("--c-secondary", t.secondary);
  if (HEX.test(t.accent)) s.setProperty("--c-accent", t.accent);
  s.setProperty("--on-primary", readableOn(t.primary));
  s.setProperty("--on-accent", readableOn(t.accent));
  loadFont(t.fontPrimary); loadFont(t.fontSecondary);
  s.setProperty("--font-primary", `"${t.fontPrimary}", system-ui, sans-serif`);
  s.setProperty("--font-secondary", `"${t.fontSecondary}", system-ui, sans-serif`);
  root.dataset.mode = resolveMode(t.mode);
}

export const saveTheme = (t: Theme) => { try { localStorage.setItem(KEY, JSON.stringify(t)); } catch {} };
export function loadTheme(): Theme {
  try { return { ...DEFAULT_THEME, ...JSON.parse(localStorage.getItem(KEY) ?? "{}") }; }
  catch { return DEFAULT_THEME; }
}
```

Reglas del motor:
- Llama a `applyTheme(loadTheme())` **antes del primer render** (script inline en `index.html` o al inicio de `main.tsx`) para evitar parpadeo.
- Si `mode` es `system`, escucha `matchMedia` y reaplica al cambiar.
- Avisa (sin bloquear) cuando `--ink` sobre `--surface` o `--on-primary` sobre `--c-primary` baje de 4.5:1, en claro y en oscuro.
- Si hay backend, guarda el tema por usuario; `localStorage` es solo caché.
- Los gráficos leen sus colores con `var(--chart-n)`; no los copies a constantes JS.

## 4. Editor de tema (Ajustes → Apariencia)

Ruta `/settings/appearance`. Controles a la izquierda y vista previa viva a la derecha, que es un mini dashboard real (KPIs, un gráfico y un item de sidebar activo).

- 3 selectores de color (`<input type="color">` + campo hex): "Color principal", "Color secundario", "Color de acento".
- 2 selectores de fuente: "Fuente principal" y "Fuente secundaria". Lista curada (Nunito Sans, Lexend, Inter, DM Sans, Poppins, Manrope, Plus Jakarta Sans, Outfit, Source Sans 3, Merriweather, Space Grotesk, IBM Plex Sans) y un campo "otra fuente de Google Fonts".
- Interruptor de modo: Claro, Oscuro, Sistema.
- Presets (Océano, Bosque, Atardecer, Grafito) que solo rellenan los 3 colores y las 2 fuentes.
- Botones "Restablecer" y "Guardar cambios". La vista previa cambia en vivo; guardar solo persiste.
- Avisos de contraste con texto claro: qué par falla y su ratio.

## 5. Layout

```
┌───────────┬─────────────────────────────────────────────────┐
│  (avatar) │ Bienvenida, Ana            [ Buscar… 🔍 ] ◐ 🔔 │
│  Ana R.   ├─────────────────────────────────────────────────┤
│  ana@…    │ Resumen                                         │
│ ───────── │ [KPI] [KPI] [KPI] [KPI alerta]                  │
│ ▌Dashboard├──────────┬──────────────────┬───────────────────┤
│  Tareas   │ Miembros │ Estado de tareas │ Top 10            │
│  Tablero  ├──────────┴──────────────────┤ colaboradores     │
│  Roadmap  │ Abiertas vs cerradas        │ (barras)          │
│  PRs      │ (área, últimos 6 meses)     │                   │
│  Miembros │                             │                   │
│  Reportes │                             │                   │
│  Ajustes  │                             │                   │
│ ───────── │                             │                   │
│  Salir    │                             │                   │
└───────────┴─────────────────────────────┴───────────────────┘
```

- **Sidebar** de 220 px en `--sidebar-bg`. Arriba, avatar circular de 64 px con borde, nombre y correo en `--on-primary`. El item activo es una píldora redondeada a la izquierda y pegada al borde derecho, con fondo `--sidebar-active-bg` y texto `--sidebar-active-ink`. Iconos de 18 px. "Salir" va fijo abajo.
- **Cabecera**: saludo en `--font-secondary` 24–28 px y a la derecha buscador en píldora con botón de lupa, interruptor de modo y campana con punto de aviso.
- **Contenido**: grid de 12 columnas, gap 16 px, padding 20–24 px. La fila de KPIs vive dentro de una tarjeta "Resumen". Las tarjetas de gráfico pueden ocupar 3, 4, 5 u 8 columnas según el dato.
- **Responsive**: bajo 1024 px la sidebar se reduce a solo iconos; bajo 720 px pasa a un menú deslizante con botón hamburguesa y las tarjetas van en una columna.
- No dibujes el marco redondeado oscuro de la captura: es parte del mockup, no del producto.

## 6. Componentes base (`src/ui/`)

Créalos primero y reutilízalos. Ninguna pantalla estiliza a mano lo que ya existe aquí.

- `Card` (`title`, `subtitle`, `action`, `menu`), fondo `--surface`, borde 1 px `--border`, radio `--radius-card`.
- `KpiTile` con caja de icono 40 px (fondo `--primary-100`, icono `--primary-500`), valor grande en `--font-secondary` y etiqueta pequeña. Variante `alert`: fondo `--accent-100`, borde `--accent-500` al 50 %, icono en acento y un icono de información con tooltip.
- `Sidebar`, `NavItem`, `UserBadge`, `Topbar`, `SearchInput`, `ModeSwitch`, `NotificationBell`.
- `DonutChart`/`PieChart`, `AreaChart`, `HBarChart`, `Sparkline`, `ChartTag` (etiqueta de máximo/mínimo) y `Legend` (cuadrados de color + etiqueta).
- `Button` (primary, accent, soft, ghost), `IconButton`, `Pill`/`Badge`, `Avatar`, `AvatarStack`, `Progress`, `Tabs`, `Select`, `DataTable`, `EmptyState`, `Skeleton`, `Toast`.
- `ThemeProvider` y `useTheme()`.

## 7. Reglas de los gráficos

Usa **Recharts** salvo que el proyecto ya tenga otra librería. Se le pasan siempre `var(--chart-n)`, `var(--chart-grid)` y `var(--ink-soft)`; nunca colores literales.

- **Líneas y áreas**: curva `monotone`, trazo 2 px, relleno con degradado de la serie del 25 % a 0 %. Solo rejilla horizontal, punteada, en `--chart-grid`. Ejes sin línea, texto de 11 px en `--ink-soft`. Marcador en el punto destacado (círculo con anillo de `--surface`). Los valores máximo y mínimo se señalan con `ChartTag`: etiqueta pequeña con el color de la serie y texto `--on-primary`/`--on-accent`.
- **Donut/Pie**: dos o tres segmentos como máximo, porcentaje dentro del segmento, leyenda a un lado con cuadrados. Segmento principal en `--chart-1`, secundario en `--chart-3`.
- **Barras horizontales**: barra de 8–10 px con extremos redondeados, nombre a la izquierda, valor al final de la barra, orden descendente, color único `--chart-1`.
- **Sparklines** sin ejes ni rejilla para los KPIs, si se necesitan.
- Cada gráfico: título, rango temporal si aplica (selector "Últimos 6 meses"), `aria-label` descriptivo y una tabla accesible alternativa (visible con un botón "Ver como tabla").
- Los colores no son el único canal de información: añade etiquetas, valores o patrones.
- Estados de carga (`Skeleton` del tamaño del gráfico), vacío y error en todos.

## 8. Mapeo a "GitHub Project"

La referencia habla de inventario y ventas; el producto gestiona proyectos. Traducción:

| Elemento de la referencia | En este producto |
|---|---|
| Bienvenida con nombre | Saludo y proyecto activo |
| KPIs "Total products / Orders / Total stock" | Tareas abiertas, En curso, PRs en revisión |
| KPI de alerta ("Out of stock") | Tareas vencidas o bloqueadas, resaltado en acento |
| "No of users" | Miembros y colaboradores activos |
| Donut "Inventory values" | Tareas por estado (cerradas vs abiertas) |
| Área "Expense vs Profit" | Tareas abiertas vs cerradas por mes, o velocidad por iteración |
| Barras "Top 10 stores" | Top 10 colaboradores por tareas cerradas |
| Sidebar: Inventory, Orders, Purchase… | Tareas, Tablero, Roadmap, Pull requests, Miembros, Reportes, Ajustes |
| Buscador | Búsqueda global y filtros (`estado:abierto asignado:yo`) |

Pantallas: Dashboard, Tareas (Tabla, Tablero kanban con arrastre accesible por teclado, y Roadmap), Pull requests, Miembros, Reportes, Ajustes (incluye Apariencia). Estados de tarea con `Pill`: abierta, en curso, en revisión, cerrada.

## 9. Contenido y tono

- Español por defecto, con todos los textos en un diccionario i18n (`src/i18n/es.ts`).
- Frases cortas, verbos claros, minúsculas salvo la primera letra. Botones: "Crear tarea", "Guardar cambios", "Ver todo". Una acción mantiene su nombre en todo el flujo.
- Estados vacíos: dicen qué falta y ofrecen la acción. Errores: qué pasó y cómo arreglarlo, sin disculpas.
- Sin datos inventados en producción: capa `src/api/` con adaptadores y mocks en `src/mocks/`.

## 10. Calidad mínima antes de dar algo por terminado

1. Cambiar los 3 colores, las 2 fuentes y el modo en Apariencia actualiza **toda** la interfaz, gráficos incluidos. Comprueba con grep que no quedan `#[0-9a-fA-F]{3,6}` ni `rgb(` fuera de `src/theme/`.
2. Contraste AA (4.5:1) en texto normal con el tema por defecto y con los 4 presets, en claro y en oscuro.
3. Foco visible y navegación completa por teclado, incluidos los gráficos con alternativa en tabla.
4. `prefers-reduced-motion` respetado. Sin animaciones decorativas; solo las que responden a una acción.
5. Responsive verificado a 1440, 1024, 720 y 390 px.
6. Estados de carga, vacío y error en cada tarjeta que consuma datos.
7. `npm run build` y el linter pasan sin errores.

## 11. Estructura de archivos recomendada

```
src/
  theme/       theme.css  theme.ts  ThemeProvider.tsx  presets.ts
  ui/          Card.tsx  KpiTile.tsx  Sidebar.tsx  Topbar.tsx  charts/ ...
  features/
    dashboard/  tasks/  board/  roadmap/  pulls/  members/  reports/  settings/
  api/         adaptadores (REST/GraphQL) y tipos
  mocks/       datos de desarrollo
  i18n/        es.ts
```