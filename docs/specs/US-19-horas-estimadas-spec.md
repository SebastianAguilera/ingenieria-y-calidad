# US-19 — Horas estimadas por historia

## Objetivo

Permitir que cualquier integrante del equipo registre las horas estimadas de una historia mediante un único campo opcional `horas_estimadas`, para poder compararlas luego con las horas reales (`esfuerzo_real`, US-06) cuando se mida la desviación de esfuerzo.

**Trazabilidad:** Issue #98 → Esta especificación → Criterios de aceptación → Escenarios BDD (`docs/features/US-19-horas-estimadas.feature`) → Tests Go (TDD) → Código.

## Dependencias y relación con otras historias

| Historia | Relación |
|---|---|
| US-02 | Define la historia y sus endpoints de alta, `PATCH` y listado, que se extienden. |
| US-06 | Mismo patrón: `horas_estimadas` es simétrico a `esfuerzo_real`. |
| US-05 | Los story points siguen siendo la estimación relativa; `horas_estimadas` no los reemplaza ni los modifica. |
| US-13 | El Sprint Backlog reutiliza la representación de historia, por lo que incluye el campo sin cambios extra. |
| US-F10 | La membresía del proyecto es la única autorización requerida (sin rol `LIDER`). |
| US-20 (futura) | Calculará la desviación de esfuerzo usando este campo. **La fórmula no se define aquí.** |

## Alcance

**Incluye**
- Campo `horas_estimadas` (opcional) en alta (`POST`) y modificación parcial (`PATCH`) de historias.
- Validación: si se informa, debe ser un número mayor a `0`.
- Persistencia con columna nullable vía `AutoMigrate`.
- Devolver el campo en alta, `PATCH`, listado de historias y Sprint Backlog.

**No incluye**
- Fórmula de desviación de esfuerzo (US-20).
- Derivar horas desde story points ni tasas por proyecto.
- Historial de estimaciones.
- Pantallas de frontend.
- Cambios en `tablas.sql`.
- Rutas nuevas.

## Entradas

### Precondiciones
- Esquema de historias (US-02) existente.
- El actor es integrante del proyecto (US-F10); cualquier rol alcanza.

### Campos

| Endpoint | Campo | Tipo JSON | Obligatorio |
|---|---|---|---|
| `POST /api/proyectos/{id}/historias` | `horas_estimadas` | número (horas) | No |
| `PATCH /api/proyectos/{id}/historias/{historiaId}` | `horas_estimadas` | número (horas) | No |

### Cuerpos de request

```json
{ "titulo": "Historia estimada en horas", "horas_estimadas": 5 }
```

```json
{ "horas_estimadas": 8 }
```

## Salidas esperadas

| Caso | Estado HTTP | Cuerpo |
|---|---|---|
| Alta con o sin `horas_estimadas` válido | 201 | historia con `horas_estimadas` (número o `null`) |
| Modificación con `horas_estimadas` válido | 200 | historia actualizada con `horas_estimadas` |
| Listado de historias | 200 | cada historia incluye `horas_estimadas` |
| Sprint Backlog (US-13) | 200 | cada `historia` anidada incluye `horas_estimadas` |
| `horas_estimadas` inválido (≤ 0 o no numérico) | 422 | `{"error":"VALIDACION","mensaje":"...","campo":"horas_estimadas"}` |

### Representación de historia

```json
{
  "id": 10,
  "codigo": "US-01",
  "titulo": "Historia estimada en horas",
  "descripcion": "",
  "prioridad": "Media",
  "estado": "Backlog",
  "story_points": null,
  "esfuerzo_real": null,
  "horas_estimadas": 5,
  "creado_en": "2026-03-02T15:04:05Z",
  "actualizado_en": "2026-03-02T15:04:05Z"
}
```

## Modelo de datos

### Dominio
- `Historia.HorasEstimadas *float64` (`nil` = sin registrar).
- `NuevaHistoria.HorasEstimadas *float64` (`nil` = no informado).
- `ActualizacionHistoria.HorasEstimadas *float64` (`nil` = no informado).
- Función de dominio: `HorasEstimadasValidas(valor float64) bool` → `valor > 0`.
- Error centinela: `ErrHorasEstimadasInvalidas`.

### Persistencia
- Columna `historias.horas_estimadas`, `numeric`, nullable, sin valor por defecto.
- Se agrega con `AutoMigrate` (sin migración manual). Las filas existentes quedan en `NULL`, sin pérdida de datos.
- Mappers y `Actualizar` del repositorio copian/escriben el campo.

## Reglas de negocio

- **RN-01** El campo es opcional: una historia se crea sin horas estimadas (`horas_estimadas: null`).
- **RN-02** Si se informa, debe ser un número mayor a `0` (se admiten decimales).
- **RN-03** `0` o negativo se rechazan con `ErrHorasEstimadasInvalidas`.
- **RN-04** Un valor no numérico (texto, booleano, objeto, arreglo) se rechaza con el mismo error y código HTTP que RN-03.
- **RN-05** En el `PATCH`, `horas_estimadas: null` equivale a "no informado": no borra el valor persistido. Si es el único campo, responde `ErrActualizacionHistoriaVacia`.
- **RN-06** Cambiar `horas_estimadas` no modifica `story_points` ni `esfuerzo_real`.
- **RN-07** Un valor inválido no modifica nada: la validación ocurre antes de tocar el repositorio.
- **RN-08** Cualquier integrante del proyecto puede registrar o modificar el campo; no se exige rol `LIDER`.

## Casos límite

- **CL-01** `horas_estimadas: 0.25` se acepta.
- **CL-02** `horas_estimadas: 0` se rechaza (para "sin registrar" se omite o se envía `null`).
- **CL-03** `horas_estimadas: "5"` (texto) se rechaza por no ser numérico.
- **CL-04** `PATCH` con solo `horas_estimadas: null` → `ErrActualizacionHistoriaVacia`.

## Criterios de aceptación

- **CA-01** Alta con `horas_estimadas: 5` devuelve `horas_estimadas: 5`.
- **CA-02** Alta sin el campo o con `null` devuelve `horas_estimadas: null`.
- **CA-03** `PATCH` con `horas_estimadas: 8` actualiza y devuelve el valor.
- **CA-04** `0`, negativo o no numérico en alta o `PATCH` → 422 `VALIDACION`, campo `horas_estimadas`, sin cambios.
- **CA-05** El listado de historias incluye `horas_estimadas`.
- **CA-06** El Sprint Backlog incluye `horas_estimadas` en cada historia.
- **CA-07** Al migrar una base con historias existentes, estas quedan con `horas_estimadas` en `null`.
- **CA-08** Cambiar `horas_estimadas` no altera `story_points` ni `esfuerzo_real`.
- **CA-09** Un integrante no líder puede crear y editar historias con `horas_estimadas`.

## Mapeo de errores HTTP

| Error de dominio | Estado | Código | Campo |
|---|---|---|---|
| `ErrHorasEstimadasInvalidas` | 422 | `VALIDACION` | `horas_estimadas` |
| `ErrActualizacionHistoriaVacia` | 422 | `VALIDACION` | — |

## Arquitectura

- **domain:** `HorasEstimadas *float64` en los tres structs, `HorasEstimadasValidas`, `ErrHorasEstimadasInvalidas`.
- **repository:** `HistoriaModel.HorasEstimadas` (`gorm:"column:horas_estimadas"`); mappers y `Actualizar` lo escriben. Sin cambios en `MigrarEsquema` (la tabla ya está registrada).
- **service:** `historia_service.go` valida en `Crear` y `Actualizar` y lo cuenta en el chequeo de actualización vacía.
- **handler:** `dto_historia.go` recibe el campo como `json.RawMessage` (`parsearHorasEstimadas`) para responder 422 y no 400 ante tipos incorrectos; `respuestaHistoria` lo emite, y con ello el Sprint Backlog. `handler/errores.go` mapea el error nuevo.

## Plan de pruebas (TDD)

- **Dominio:** `TestHorasEstimadasValidas` (positivo, decimal, cero, negativo).
- **Servicio:** alta y `PATCH` con válido, ausente, cero y negativo; `null` no borra; no altera `story_points` ni `esfuerzo_real`.
- **Handler:** 201/200 con el campo, `null` si falta, 422 con campo `horas_estimadas` (cero, negativo, no numérico en alta y `PATCH`), listado y Sprint Backlog.
- **Repositorio (integración):** persistencia en alta, `Actualizar` y listado; migración sobre una fila previa que queda en `NULL`.
- **E2E por curl:** alta con valor, alta sin valor, `PATCH`, inválidos y listado.
