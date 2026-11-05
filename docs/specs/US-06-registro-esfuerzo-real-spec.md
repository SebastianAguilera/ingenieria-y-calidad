# US-06 — Registro de esfuerzo real en la historia

## Objetivo

Permitir que el integrante del equipo registre el esfuerzo real (horas trabajadas) de una historia de usuario mediante un único campo `esfuerzo_real`, para poder comparar posteriormente el esfuerzo estimado con el real (base de la métrica de desviación de esfuerzo).

**Trazabilidad:** Issue #33 → Esta especificación → Criterios de aceptación → Escenarios BDD (`docs/features/US-06-registro-esfuerzo-real.feature`) → Tests Go (TDD) → Código.

## Alcance

**Incluye**
- Aceptar `esfuerzo_real` en el alta (`POST`) y en la modificación parcial (`PATCH`) de una historia.
- Validar que, si se informa, sea un número mayor a `0`.
- Persistir y devolver `esfuerzo_real` en las respuestas de historia (alta, modificación y listado).

**No incluye**
- Registro por usuario, fecha y descripción de actividad.
- Acumulación de múltiples registros / worklogs por historia.
- Rutas nuevas: se extienden los endpoints existentes de US-02.
- Cálculo de la métrica de desviación (corresponde al motor de métricas).

## Entradas

### Precondiciones
- Esquema de US-02 (historias) existente.
- Proyecto e historia referenciados existen.

### Campos

| Endpoint | Campo | Tipo JSON | Obligatorio |
|---|---|---|---|
| `POST /api/proyectos/{id}/historias` | `esfuerzo_real` | número (horas) | No |
| `PATCH /api/proyectos/{id}/historias/{historiaId}` | `esfuerzo_real` | número (horas) | No |

### Cuerpos de request

```json
{
  "titulo": "Historia con esfuerzo registrado",
  "esfuerzo_real": 4.5
}
```

```json
{
  "esfuerzo_real": 6
}
```

## Salidas esperadas

| Caso | Estado HTTP | Cuerpo |
|---|---|---|
| Alta con o sin `esfuerzo_real` válido | 201 | historia con `esfuerzo_real` (número o `null`) |
| Modificación con `esfuerzo_real` válido | 200 | historia actualizada con `esfuerzo_real` |
| `esfuerzo_real` inválido (≤ 0 o no numérico) | 422 | `{"error":"VALIDACION","mensaje":"...","campo":"esfuerzo_real"}` |

### Representación de historia

```json
{
  "id": 10,
  "codigo": "US-01",
  "titulo": "Historia con esfuerzo registrado",
  "descripcion": "",
  "prioridad": "Media",
  "estado": "Backlog",
  "story_points": null,
  "esfuerzo_real": 4.5,
  "creado_en": "2026-03-02T15:04:05Z",
  "actualizado_en": "2026-03-02T15:04:05Z"
}
```

## Modelo de datos

### Dominio
- `Historia.EsfuerzoReal *float64` (`nil` = sin registrar).
- `NuevaHistoria.EsfuerzoReal *float64` (`nil` = no informado).
- `ActualizacionHistoria.EsfuerzoReal *float64` (`nil` = no informado).
- Función de dominio: `EsfuerzoRealValido(valor float64) bool` → `valor > 0`.
- Error centinela: `ErrEsfuerzoRealInvalido`.

### Persistencia
- Columna `historias.esfuerzo_real` de tipo `numeric`, nullable (GORM mapea `*float64` a `decimal`/`numeric` en PostgreSQL).
- Se agrega vía `AutoMigrate` (sin migración manual).
- Mappers `ToDomainHistoria` / `FromDomainHistoria` deben copiar el campo.
- `Crear` y `Actualizar` deben persistir el campo.

## Reglas de negocio

- **RN-01** El campo es opcional: una historia se puede crear sin esfuerzo real (`esfuerzo_real: null`).
- **RN-02** Si se informa, debe ser un número mayor a `0` (se admiten decimales, por ejemplo `4.5`).
- **RN-03** Un valor `0` o negativo se rechaza con `ErrEsfuerzoRealInvalido`.
- **RN-04** Un valor no numérico (texto, booleano, objeto o arreglo) se rechaza con el mismo error y el mismo código HTTP que RN-03.
- **RN-05** `esfuerzo_real` se devuelve en el alta, en la modificación y en el listado.
- **RN-06** En el `PATCH`, `esfuerzo_real: null` equivale a "no informado": no borra el valor persistido.
- **RN-07** En el `PATCH`, enviar únicamente `esfuerzo_real` es una modificación válida.
- **RN-08** El valor es único por historia: no se acumula ni multiplica (los worklogs múltiples están fuera de alcance).

## Casos límite

- **CL-01** `esfuerzo_real: 4.5` se acepta y se persiste como `4.5`.
- **CL-02** `esfuerzo_real: 0` se rechaza (no es "sin registrar": para eso se omite o se envía `null`).
- **CL-03** `esfuerzo_real` negativo se rechaza.
- **CL-04** `esfuerzo_real: null` en el `PATCH` equivale a "no informado": no borra el valor. Si es el único campo, responde `ErrActualizacionHistoriaVacia`.
- **CL-05** Un decimal muy pequeño pero mayor a `0` (por ejemplo `0.25`) se acepta.
- **CL-06** Un número con representación textual (por ejemplo `"6"`) se rechaza por no ser numérico.

## Criterios de aceptación

- **CA-F1** Se registra un `esfuerzo_real` válido en el alta y la respuesta lo devuelve.
- **CA-F2** Se registra o actualiza un `esfuerzo_real` válido en la modificación y el detalle lo refleja.
- **CA-F3** Una historia sin `esfuerzo_real` devuelve `esfuerzo_real: null`.
- **CA-F4** El `esfuerzo_real` aparece en el listado del backlog.
- **CA-R1** Valores `0` o negativos devuelven 422 `VALIDACION` con campo `esfuerzo_real`.
- **CA-R2** Valores no numéricos devuelven 422 `VALIDACION` con campo `esfuerzo_real`.

## Mapeo de errores HTTP

| Error de dominio | Estado | Código | Campo |
|---|---|---|---|
| `ErrEsfuerzoRealInvalido` | 422 | `VALIDACION` | `esfuerzo_real` |
| `ErrActualizacionHistoriaVacia` | 422 | `VALIDACION` | — |

## Arquitectura

- **domain:** `EsfuerzoReal *float64` en los tres structs, `EsfuerzoRealValido` y `ErrEsfuerzoRealInvalido`.
- **repository:** `HistoriaModel.EsfuerzoReal *float64` con `gorm:"column:esfuerzo_real"`; mappers y `Actualizar` escriben la columna.
- **service:** `historia_service.go` valida el campo en `Crear` y `Actualizar`, y lo incluye en el chequeo de actualización vacía.
- **handler:** `dto_historia.go` recibe `esfuerzo_real` como `json.RawMessage` para distinguir ausente, `null` y tipo incorrecto, y responder 422 (no 400) si no es numérico; `respuestaHistoria` incluye el campo.

## Plan de pruebas (TDD)

- **Dominio:** `TestEsfuerzoRealValido` — valores positivos, decimales, `0` y negativos.
- **Servicio:** alta y modificación con `esfuerzo_real`, propagación de `ErrEsfuerzoRealInvalido`, campo único en el `PATCH` y `null` como "no informado".
- **Handler:** 201/200 con `esfuerzo_real`, 422 con campo `esfuerzo_real` para `0`, negativos y no numéricos (en alta y modificación), y `null` en `PATCH`.
- **Repositorio:** integración de persistencia y lectura del campo en alta y modificación.
- **BDD:** `docs/features/US-06-registro-esfuerzo-real.feature` con escenarios `BDD-01` a `BDD-07`, referenciados en los mensajes/etiquetas de los tests Go.
