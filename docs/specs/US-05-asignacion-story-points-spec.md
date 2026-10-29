# US-05 — Asignación inicial de Story Points

## Objetivo

Permitir al Product Builder asignar o editar manualmente los Story Points de una historia de usuario, para tener una estimación base antes del Planning Poker.

**Trazabilidad:** Issue #32 → Esta especificación → Criterios de aceptación → Escenarios BDD (`features/US-05-asignacion-story-points.feature`) → Tests Go (TDD) → Código.

## Alcance

**Incluye**
- Aceptar `story_points` en el alta (`POST`) y en la modificación parcial (`PATCH`) de una historia.
- Validar la escala Fibonacci permitida y rechazar valores fuera de escala, negativos o no numéricos.
- Persistir y devolver `story_points` en las respuestas de historia (alta, modificación y listado).

**No incluye**
- Planning Poker ni estimación colaborativa.
- Borrar una estimación ya asignada (ver CL-04).
- Rutas nuevas: se extienden los endpoints existentes de US-02.

## Entradas

| Endpoint | Campo | Tipo JSON | Obligatorio |
|---|---|---|---|
| `POST /api/proyectos/{id}/historias` | `story_points` | número | No |
| `PATCH /api/proyectos/{id}/historias/{historiaId}` | `story_points` | número | No |

Escala permitida: `0, 0.5, 1, 2, 3, 5, 8, 13, 21`.

## Salidas esperadas

| Caso | Estado HTTP | Cuerpo |
|---|---|---|
| Alta con o sin `story_points` válido | 201 | historia con `story_points` (número o `null`) |
| Modificación con `story_points` válido | 200 | historia actualizada |
| `story_points` inválido | 422 | `{"error":"VALIDACION","mensaje":"...","campo":"story_points"}` |

## Modelo de datos

- **Dominio:** `Historia.StoryPoints *float64` (nil = sin estimación). `NuevaHistoria.StoryPoints *float64`. `ActualizacionHistoria.StoryPoints *float64` (nil = no informado).
- **Función de dominio:** `StoryPointsValidos(valor float64) bool`.
- **Error centinela:** `ErrStoryPointsInvalido`.
- **Persistencia:** columna `historias.story_points` pasa de entero a `double precision`, nullable.

## Reglas de negocio

- **RN-01** El valor debe pertenecer a la escala `0, 0.5, 1, 2, 3, 5, 8, 13, 21`.
- **RN-02** El campo es opcional: una historia se puede crear sin estimación (`story_points: null`).
- **RN-03** Un valor negativo se rechaza con `ErrStoryPointsInvalido`.
- **RN-04** Un valor no numérico (texto, booleano, objeto, arreglo) se rechaza con el mismo error y el mismo código HTTP que RN-01 y RN-03.
- **RN-05** `0` es un valor válido y distinto de "sin estimación".
- **RN-06** La asignación a un Sprint (US-04) conserva los Story Points.

## Casos límite

- **CL-01** `story_points: 0` se acepta y se persiste como `0`, no como `null`.
- **CL-02** `story_points: 0.5` se acepta (la escala admite decimales).
- **CL-03** `story_points: 4` o `22` se rechaza (fuera de escala).
- **CL-04** `story_points: null` en el `PATCH` equivale a "no informado": no borra la estimación. Si es el único campo, responde `ErrActualizacionHistoriaVacia`.
- **CL-05** En el `PATCH`, enviar solo `story_points` es una modificación válida.

## Criterios de aceptación

- **CA-F1** Se asigna un valor válido en el alta y en la modificación, y el detalle lo refleja.
- **CA-F2** Se crea una historia sin `story_points` y la respuesta trae `story_points: null`.
- **CA-R1** Valores fuera de escala devuelven 422 `VALIDACION` con campo `story_points`.
- **CA-R2** Valores negativos o no numéricos devuelven 422 `VALIDACION` con campo `story_points`.

## Mapeo de errores HTTP

| Error de dominio | Estado | Código | Campo |
|---|---|---|---|
| `ErrStoryPointsInvalido` | 422 | `VALIDACION` | `story_points` |
| `ErrActualizacionHistoriaVacia` | 422 | `VALIDACION` | — |

## Arquitectura

- **handler:** `dto_historia.go` recibe `story_points` como `json.RawMessage` para poder distinguir ausente, `null` y tipo incorrecto, y devolver 422 (no 400) si no es numérico.
- **service:** `historia_service.go` valida la escala en `Crear` y `Actualizar` con `domain.StoryPointsValidos`.
- **repository:** `HistoriaModel.StoryPoints *float64`; `Actualizar` escribe `story_points`.
