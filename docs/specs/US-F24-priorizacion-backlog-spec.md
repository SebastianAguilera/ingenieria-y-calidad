# Especificación: US-F24 - Valor de negocio y backlog priorizado

## Objetivo

Agregar a la pantalla de **Backlog** la carga de los tres valores de
priorización de una historia (`valor_negocio`, `criticidad_tiempo` y
`reduccion_riesgo`) y una vista de **Backlog priorizado** que consume el
endpoint de US-23 y muestra las historias no `Hecho` ordenadas por WSJF, con su
ROI y WSJF.

La historia satisface el issue
`SebastianAguilera/ingenieria-y-calidad#113`
(`US-F24 - Valor de negocio y backlog priorizado`). **No modifica el backend**:
solo consume lo que ya expone US-23.

**Trazabilidad:** Issue #113 → esta especificación → criterios de aceptación
(CA-01 a CA-10) → escenarios BDD
(`docs/features/US-F24-priorizacion-backlog.feature`) → tests Vitest/RTL →
código en `frontend/src`.

| Eslabón | Artefacto |
| :--- | :--- |
| Issue | [SebastianAguilera/ingenieria-y-calidad#113](https://github.com/SebastianAguilera/ingenieria-y-calidad/issues/113) |
| SDD | `docs/specs/US-F24-priorizacion-backlog-spec.md` (este archivo) |
| Criterios de aceptación | CA-01 a CA-10 de este documento |
| BDD | `docs/features/US-F24-priorizacion-backlog.feature` (E-01 a E-10) |
| Tests | `frontend/src/api/historias.test.ts`, `frontend/src/pages/proyecto/backlog/ModalHistoria.test.tsx`, `frontend/src/pages/proyecto/backlog/VistaBacklogPriorizado.test.tsx`, `frontend/src/pages/proyecto/BacklogPage.test.tsx` |
| Código | `frontend/src/api/types.ts`, `frontend/src/api/historias.ts`, `frontend/src/pages/proyecto/backlog/etiquetas.ts`, `frontend/src/pages/proyecto/backlog/tipos.ts`, `frontend/src/pages/proyecto/backlog/ModalHistoria.tsx`, `frontend/src/pages/proyecto/backlog/useHistorias.ts`, `frontend/src/pages/proyecto/backlog/useBacklogPriorizado.ts`, `frontend/src/pages/proyecto/backlog/VistaBacklogPriorizado.tsx`, `frontend/src/pages/proyecto/BacklogPage.tsx` |

## Dependencias y relación con otras historias

| Historia | Relación |
| :--- | :--- |
| US-23 | **Bloqueante**. Aporta los tres campos, el endpoint priorizado, ROI, WSJF y el orden. Ya está implementada en el backend. |
| US-F07 | Aporta la pantalla de Backlog, el selector Tabla/Tablero, el modal de historia y los estados de carga/error/vacío. |
| US-02 / US-05 | Aportan la prioridad manual y `story_points`. La prioridad manual sigue siendo independiente de WSJF. |
| US-F10 / US-11 | Guard de membresía: un no miembro recibe `404 PROYECTO_NO_ENCONTRADO`. |
| US-F21 | No se toca. ROI y WSJF no son métricas de avance y no van al dashboard. |

**Bloquea a:** ninguna historia.

## Alcance

**Incluye**

- Tres selectores opcionales en `ModalHistoria`: `valor_negocio`,
  `criticidad_tiempo` y `reduccion_riesgo`.
- Escala cerrada `1, 2, 3, 5, 8, 13, 21` más la opción `Sin asignar`.
- Ayuda breve por campo, asociada con `aria-describedby`.
- Envío de los tres valores en el alta y en el PATCH de la historia.
- Tercera vista `Priorizado` dentro del selector Tabla/Tablero de
  `BacklogPage`.
- Tabla de solo lectura con Código, Título, Estado, Prioridad, SP, Valor,
  Criticidad, Riesgo, ROI y WSJF.
- `—` cuando `roi` o `wsjf` son `null`.
- Estados de carga, vacío, error y error de red con reintento.
- Tipos en `frontend/src/api/types.ts` para los campos y para la respuesta
  priorizada.
- Errores `422 VALIDACION` mostrados en el campo indicado por `campo`.
- Tests Vitest/RTL de API, modal, vista priorizada e integración en la pantalla.

**No incluye**

- Cambios en el backend, en `tablas.sql` ni en las migraciones.
- Reordenar historias arrastrando ni persistir el orden calculado.
- Priorización automática, Monte Carlo o historial de valores.
- Cálculo de ROI/WSJF en el frontend: se muestran los valores del backend.
- Editar `story_points` desde el modal.
- Cambios en el dashboard de métricas.

## Entradas y precondiciones

### Precondiciones

- US-23 implementada en el backend.
- Usuario autenticado y miembro del proyecto (US-F10).
- El usuario está en `/proyectos/:id/backlog` (US-F07).
- `id` de la ruta es un entero positivo.

### Endpoint consumido

| Endpoint | Método | Origen |
| :--- | :---: | :--- |
| `/api/proyectos/{id}/backlog/priorizado` | `GET` | US-23 |

La respuesta real es un objeto con sobre:

```json
{
  "total": 2,
  "historias": [
    {
      "id": 5,
      "codigo": "US-05",
      "titulo": "Historia priorizada",
      "descripcion": "Con valores cargados",
      "prioridad": "Alta",
      "estado": "Backlog",
      "story_points": 5,
      "esfuerzo_real": null,
      "horas_estimadas": null,
      "valor_negocio": 8,
      "criticidad_tiempo": 3,
      "reduccion_riesgo": 2,
      "roi": 1.6,
      "wsjf": 2.6,
      "creado_en": "2026-11-15T10:00:00Z",
      "actualizado_en": "2026-11-15T10:00:00Z"
    }
  ]
}
```

El cliente devuelve el arreglo `historias` (o `[]` si falta), igual que
`listarHistorias`.

### Datos de entrada del modal

| Campo | Tipo | Obligatorio | Regla |
| :--- | :--- | :---: | :--- |
| `valor_negocio` | número \| `null` | no | escala `1, 2, 3, 5, 8, 13, 21` |
| `criticidad_tiempo` | número \| `null` | no | escala `1, 2, 3, 5, 8, 13, 21` |
| `reduccion_riesgo` | número \| `null` | no | escala `1, 2, 3, 5, 8, 13, 21` |

`Sin asignar` se traduce a `null`. El alta y el PATCH siguen enviando los
campos existentes (`titulo`, `descripcion`, `prioridad`, `estado`).

## Salidas esperadas

### Modal de alta/edición

- Tres selectores etiquetados: **Valor de negocio**, **Criticidad en el
  tiempo** y **Reducción de riesgo**.
- Cada selector ofrece `Sin asignar` y los valores `1, 2, 3, 5, 8, 13, 21`; no
  hay texto libre.
- Ayuda visible por campo (por ejemplo, “Valor que aporta al negocio o al
  cliente”), enlazada con `aria-describedby`.
- Al editar, los selectores muestran los valores ya cargados.
- Si la historia ya tiene un valor, `Sin asignar` queda deshabilitado porque el
  PATCH con `null` no borra el valor (RN-06 de US-23).
- Un `422 VALIDACION` con `campo` igual a uno de los tres nombres se muestra
  junto a ese selector.

### Vista Priorizado

| Columna | Contenido |
| :--- | :--- |
| Código | `codigo` de la historia |
| Título | `titulo` truncado con `title` completo |
| Estado | `estado` |
| Prioridad | `prioridad` manual |
| SP | `story_points` o `—` |
| Valor | `valor_negocio` o `—` |
| Criticidad | `criticidad_tiempo` o `—` |
| Riesgo | `reduccion_riesgo` o `—` |
| ROI | `roi` o `—` |
| WSJF | `wsjf` o `—` |

- El orden es exactamente el que devuelve el backend: WSJF descendente, nulos
  al final, desempate por ID. El frontend no ordena.
- La tabla tiene encabezados accesibles y scroll horizontal en pantallas
  angostas.
- No incluye historias `Hecho` porque el backend ya las excluye.

### Estados de pantalla

| Estado | Comportamiento |
| :--- | :--- |
| Cargando | `role="status"` con `Cargando backlog priorizado…` |
| Vacío | Mensaje `No hay historias pendientes de priorizar` |
| Error | `Alert` de error y botón `Reintentar` |
| Error de red | Mismo tratamiento que error, con reintento |
| Sin acceso (404) | `No tenés acceso a este proyecto o no existe` |

## Reglas de negocio

| ID | Regla |
| :--- | :--- |
| RN-01 | Los tres valores son opcionales. `Sin asignar` equivale a `null`. |
| RN-02 | Solo se ofrecen los valores `1, 2, 3, 5, 8, 13, 21`; la escala de valor no admite `0` ni `0.5`. |
| RN-03 | El modal no calcula ROI ni WSJF. |
| RN-04 | El alta y el PATCH envían los tres valores junto con los campos existentes. |
| RN-05 | Al editar una historia con un valor cargado, `Sin asignar` se deshabilita porque el backend no borra valores con `null`. |
| RN-06 | La vista priorizada es de solo lectura: no crea, edita ni elimina. |
| RN-07 | El frontend respeta el orden y los valores de ROI/WSJF que devuelve el backend. |
| RN-08 | `roi` o `wsjf` en `null` se muestran como `—`. |
| RN-09 | Las historias en `Hecho` no se muestran porque el endpoint las excluye. |
| RN-10 | Un `422` con `campo` igual a un valor de priorización se muestra en ese selector; otros errores se muestran como alerta general. |
| RN-11 | Un `404 PROYECTO_NO_ENCONTRADO` se muestra como “No tenés acceso a este proyecto o no existe”. |
| RN-12 | La prioridad manual (`Alta`, `Media`, `Baja`) no se modifica por WSJF. |

## Restricciones

### Técnicas

- No se toca el backend, `internal/`, `cmd/`, `sql/` ni `tablas.sql`.
- No se agregan dependencias npm.
- Se reutilizan `api.get`, `Alert`, `Button`, `Card`, `PageHeader` y los estilos
  del sistema de diseño (`teal-dashboard-ui`).
- Se usan `<select>` nativos: no hay combobox propio ni librería de componentes.
- La vista priorizada solo se consulta cuando el usuario la selecciona.

### De arquitectura frontend

- `api/` solo define tipos y llamadas HTTP.
- `backlog/` concentra componentes, hooks y etiquetas del backlog.
- Los tests no golpean la red: usan mocks de `api/client` o de
  `api/historias`.

## Casos límite

| ID | Caso | Resultado |
| :--- | :--- | :--- |
| CL-01 | Historia sin story points | ROI y WSJF `—`. |
| CL-02 | Historia con `roi: null` y `wsjf: null` | Ambos `—`; el resto de columnas se muestra igual. |
| CL-03 | Historia sin ninguno de los tres valores | Valor, criticidad, riesgo, ROI y WSJF `—` (o los valores que devuelva el backend). |
| CL-04 | Proyecto sin historias pendientes | Estado vacío de la vista priorizada. |
| CL-05 | Proyecto sin acceso | Mensaje de 404 y sin tabla. |
| CL-06 | Error de red | Alerta y reintento; el reintento vuelve a llamar. |
| CL-07 | Respuesta sin `historias` | Se trata como arreglo vacío. |
| CL-08 | Valor inválido en el selector | No es posible seleccionarlo. |
| CL-09 | 422 en `criticidad_tiempo` | El mensaje aparece bajo ese selector. |
| CL-10 | Dos consultas consecutivas | Mismo resultado; no hay efectos laterales. |

## Criterios de aceptación

- [ ] **CA-01** Given el modal de historia, When se abre, Then aparecen los tres
  selectores de priorización con su ayuda visible.
- [ ] **CA-02** Given un selector de priorización, When se despliega, Then solo
  ofrece `Sin asignar` y `1, 2, 3, 5, 8, 13, 21`.
- [ ] **CA-03** Given los tres valores elegidos, When se guarda, Then se envían
  al backend y al volver a editar la historia se muestran.
- [ ] **CA-04** Given la vista Priorizado, When se carga, Then las historias
  aparecen en el orden de WSJF que devuelve el backend.
- [ ] **CA-05** Given una historia sin story points, When aparece en la vista,
  Then ROI y WSJF muestran `—`.
- [ ] **CA-06** Given historias en `Hecho`, When se consulta la vista, Then no
  aparecen.
- [ ] **CA-07** Given un `422 VALIDACION` en un campo de priorización, When se
  guarda, Then el mensaje aparece en ese campo.
- [ ] **CA-08** Given un error de red, When se carga la vista, Then se muestra el
  mensaje y se puede reintentar.
- [ ] **CA-09** Given la vista Priorizado, When los datos están en camino, Then
  se muestra el estado de carga.
- [ ] **CA-10** Given un proyecto sin historias pendientes, When se abre la
  vista, Then se muestra el estado vacío y no un error.

## Condiciones de error y mapeo HTTP

| Situación | Estado | Código API | Campo |
| :--- | :---: | :--- | :--- |
| `{id}` no numérico o `0` | 400 | `PARAMETRO_INVALIDO` | — |
| Sin token o token inválido | 401 | `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` | — |
| Proyecto inexistente o no miembro | 404 | `PROYECTO_NO_ENCONTRADO` | — |
| `valor_negocio` inválido | 422 | `VALIDACION` | `valor_negocio` |
| `criticidad_tiempo` inválida | 422 | `VALIDACION` | `criticidad_tiempo` |
| `reduccion_riesgo` inválida | 422 | `VALIDACION` | `reduccion_riesgo` |
| Error de red | — | — | — |
| Error interno | 500 | `ERROR_INTERNO` | — |

## Arquitectura (frontend)

- **api**: `types.ts` agrega los tres campos a `Historia`, `AltaHistoria` y
  `ActualizacionHistoria`, y define `HistoriaPriorizada` y
  `ListadoBacklogPriorizado`; `historias.ts` agrega
  `listarBacklogPriorizado`.
- **backlog/etiquetas.ts**: agrega `VALORES_PRIORIZACION` y los textos de ayuda.
- **backlog/tipos.ts**: extiende `DatosHistoria` con los tres campos.
- **backlog/ModalHistoria.tsx**: agrega los selectores, la ayuda, la
  inicialización desde `historia` y el mapeo de `campo` para 422.
- **backlog/useHistorias.ts**: propaga los tres campos en la actualización
  optimista.
- **backlog/useBacklogPriorizado.ts**: hook de carga con estados y reintento.
- **backlog/VistaBacklogPriorizado.tsx**: tabla de solo lectura y estados.
- **BacklogPage.tsx**: agrega la tercera opción `Priorizado` y monta la vista
  cuando se selecciona.

## Plan de pruebas (TDD)

Ciclo **RED → GREEN → REFACTOR**. Cada ID de escenario BDD se usa como nombre de
test.

### API (`frontend/src/api/historias.test.ts`)

- `listarBacklogPriorizado` llama a
  `GET /proyectos/{id}/backlog/priorizado` con `signal`.
- Devuelve el arreglo `historias`.
- Tolera una respuesta sin `historias`.
- Propaga el error.

### Modal (`frontend/src/pages/proyecto/backlog/ModalHistoria.test.tsx`)

- Renderiza los tres selectores con etiqueta y ayuda (`aria-describedby`).
- Opciones exactas de escala.
- Envía los tres valores numéricos.
- Carga los valores existentes al editar.
- Deshabilita `Sin asignar` cuando hay un valor cargado.
- `Sin asignar` se envía como `null`.
- Mapea 422 al campo correcto.

### Vista priorizada

`frontend/src/pages/proyecto/backlog/VistaBacklogPriorizado.test.tsx`

- Tabla con las columnas y el orden recibido.
- `—` para `roi`, `wsjf`, `story_points` y valores `null`.
- Estado de carga.
- Estado vacío.
- Error de red y reintento.
- No muestra historias `Hecho` (fixture del backend ya excluye).

### Integración (`frontend/src/pages/proyecto/BacklogPage.test.tsx`)

- El selector incluye `Tabla`, `Tablero` y `Priorizado`.
- Elegir `Priorizado` consulta el endpoint y muestra la tabla priorizada.
- Tabla y Tablero mantienen su comportamiento.
- Fixtures y mocks actualizados con los tres campos.

### Fixtures

Todos los literales de `Historia` de los tests agregan
`valor_negocio: null`, `criticidad_tiempo: null` y `reduccion_riesgo: null`,
salvo los casos que necesiten valores.

## Resumen de escenarios BDD

Archivo: `docs/features/US-F24-priorizacion-backlog.feature`.

| ID | Tipo | Escenario | CA |
| :--- | :--- | :--- | :--- |
| E-01 | Normal | Campos de priorización con ayuda visible | CA-01 |
| E-02 | Límite | El selector solo ofrece la escala Fibonacci | CA-02 |
| E-03 | Normal | Guardar los tres valores y verlos al editar | CA-03 |
| E-04 | Normal | Valores existentes cargados en el modal | CA-03 |
| E-05 | Normal | Vista priorizada en el orden de WSJF del backend | CA-04 |
| E-06 | Límite | Historia sin story points muestra `—` en ROI y WSJF | CA-05 |
| E-07 | Alternativo | Historia en `Hecho` no aparece | CA-06 |
| E-08 | Error | 422 en un campo de priorización | CA-07 |
| E-09 | Error | Error de red al cargar la vista y reintento | CA-08 |
| E-10 | No funcional | Estados de carga y vacío | CA-09, CA-10 |

## Decisiones

- **D-01 — Ubicación.** La vista priorizada es una tercera opción del selector
  `Tabla / Tablero / Priorizado`, no una ruta nueva.
- **D-02 — Cálculos en el backend.** El frontend no calcula ROI ni WSJF ni
  reordena.
- **D-03 — `Sin asignar` en edición.** Se deshabilita si la historia ya tiene
  valor porque el PATCH con `null` no borra (RN-06 de US-23).
- **D-04 — Contrato de respuesta.** El endpoint responde `{ total, historias }`;
  el cliente devuelve `historias` y tolera su ausencia.
- **D-05 — `—`.** Aplica a `story_points`, los tres valores y `roi`/`wsjf`
  cuando son `null`.
- **D-06 — Sin backend.** No se modifica ninguna capa Go ni el esquema.

## Definición de terminado

- [ ] `npm test`, `npm run lint` y `npm run build` en verde.
- [ ] Spec SDD y feature BDD escritos antes del código.
- [ ] Tests RED confirmados antes de implementar.
- [ ] Todos los fixtures de `Historia` actualizados con los tres campos.
- [ ] Verificación manual con historias de distinto valor y tamaño.
- [ ] Sin cambios en backend ni en `tablas.sql`.
