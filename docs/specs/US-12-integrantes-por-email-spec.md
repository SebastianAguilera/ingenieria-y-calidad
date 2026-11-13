# Especificación: US-12 - Integrantes por email con rol

## Objetivo

Que el `LIDER` de un proyecto pueda **armar el equipo escribiendo el email de
una persona ya registrada**, asignarle un rol (`LIDER`/`MIEMBRO`), cambiarlo y
quitarla del proyecto. Cualquier integrante puede consultar el equipo con su rol.
Nadie entra a un proyecto sin ser invitado.

La historia consume las reglas de US-11 (roles, `EXIGIR_LIDER`, `403
ACCESO_DENEGADO`, `409 ULTIMO_LIDER`) y el guard de membresía de US-F10 (`404
PROYECTO_NO_ENCONTRADO`). No las redefine.

Esta historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#68` y
**reemplaza** el alta por `usuario_id` de US-08 en el contrato HTTP: el alta
pasa a ser por `email`. No incluye interfaz (US-F12).

**Trazabilidad:** Issue #68 → SDD → criterios de aceptación → escenarios BDD
(`docs/features/US-12-integrantes-por-email.feature`) → tests Go
(`internal/domain`, `internal/repository`, `internal/service`, `internal/handler`,
`cmd/api`) → código.

## Cross-referencias

- **US-08**: usuarios con email único y vínculo/desvínculo. El alta por
  `usuario_id` se reemplaza por alta por email.
- **US-10 / US-F08**: identidad del request (`sub` del JWT).
- **US-F10**: guard de membresía; un no miembro recibe `404` antes que cualquier
  `403`.
- **US-11**: roles `LIDER`/`MIEMBRO`, `EsLider`, auto-baja, `ULTIMO_LIDER` y
  `CambiarRol`.
- **US-F12**: interfaz de gestión de integrantes (fuera de alcance).

## Entradas

### Precondiciones

- El actor está autenticado (JWT) y es integrante del proyecto.
- El usuario a agregar ya existe en `usuarios` (US-08).

### `POST /api/proyectos/{id}/usuarios`

| Campo | Tipo | Requerido | Formato |
| :--- | :--- | :---: | :--- |
| `email` | string | Sí | Email válido, se normaliza con `TrimSpace` + minúsculas |
| `rol` | string | No | `LIDER` o `MIEMBRO`; omitido/vacío → `MIEMBRO` |

```json
{ "email": "ada@utn.edu.ar", "rol": "MIEMBRO" }
```

### `PATCH /api/proyectos/{id}/usuarios/{usuarioId}`

| Campo | Tipo | Requerido | Formato |
| :--- | :--- | :---: | :--- |
| `rol` | string | Sí | `LIDER` o `MIEMBRO` |

```json
{ "rol": "LIDER" }
```

### `GET /api/proyectos/{id}/usuarios`

Sin body. Devuelve el equipo con el rol de cada integrante.

### `DELETE /api/proyectos/{id}/usuarios/{usuarioId}`

Sin body. Un `LIDER` puede quitar a cualquier integrante; un integrante puede
quitarse a sí mismo (US-11, RB-05).

## Salidas esperadas

| # | Método | Path | Descripción | Éxito | Errores |
| :---: | :--- | :--- | :--- | :---: | :--- |
| 1 | `GET` | `/api/proyectos/{id}/usuarios` | Listar integrantes con rol | `200 OK` | `401`, `404`, `500` |
| 2 | `POST` | `/api/proyectos/{id}/usuarios` | Alta por email con rol | `200 OK` | `400`, `401`, `403`, `404`, `409`, `422`, `500` |
| 3 | `PATCH` | `/api/proyectos/{id}/usuarios/{usuarioId}` | Cambiar rol | `200 OK` | `400`, `401`, `403`, `404`, `409`, `422`, `500` |
| 4 | `DELETE` | `/api/proyectos/{id}/usuarios/{usuarioId}` | Quitar integrante / auto-baja | `204 No Content` | `401`, `403`, `404`, `409`, `500` |

### Forma de `GET .../usuarios`

```json
[
  { "id": 1, "nombre": "Ana", "email": "ana@utn.edu.ar", "rol": "LIDER" },
  { "id": 2, "nombre": "Ada", "email": "ada@utn.edu.ar", "rol": "MIEMBRO" }
]
```

El array siempre está presente; un proyecto sin integrantes responde `[]`.

### Forma de `POST .../usuarios`

```json
{
  "id": 2,
  "nombre": "Ada",
  "email": "ada@utn.edu.ar",
  "rol": "MIEMBRO",
  "proyecto_id": 1,
  "creado_en": "2026-11-12T12:00:00Z"
}
```

### Forma de `PATCH .../usuarios/{usuarioId}`

```json
{ "usuario_id": 2, "rol": "LIDER" }
```

## Reglas de negocio

- **RB-01**: el email se normaliza (`TrimSpace` + minúsculas) antes de buscar.
- **RB-02**: solo un `LIDER` agrega, cambia roles o quita a otros integrantes
  (US-11, RB-03). Un `MIEMBRO` recibe `403 ACCESO_DENEGADO` y nada cambia.
- **RB-03**: un integrante puede quitarse a sí mismo sea cual sea su rol
  (US-11, RB-05). Es la única operación de gestión permitida a un `MIEMBRO`.
- **RB-04**: el usuario a agregar debe existir; si no, `404` y **no se crea**.
- **RB-05**: un usuario no puede ser integrante dos veces del mismo proyecto.
- **RB-06**: no se puede quitar ni degradar al último `LIDER` (`409
  ULTIMO_LIDER`, US-11 RB-06).
- **RB-07**: cualquier integrante puede listar el equipo; un no miembro recibe
  `404 PROYECTO_NO_ENCONTRADO` (US-F10).
- **RB-08**: un proyecto cerrado no admite alta, cambio de rol ni baja
  (`409 PROYECTO_CERRADO`); el listado es de solo lectura y sigue disponible.
- **RB-09**: el rol omitido o vacío en el alta se interpreta como `MIEMBRO`
  (US-11, RB-02).

## Errores

| Error de dominio | HTTP | `error` | `campo` | Mensaje |
| :--- | :---: | :--- | :--- | :--- |
| `ErrUsuarioEmailInvalido` | `422` | `VALIDACION` | `email` | email obligatorio y con formato válido |
| `ErrRolInvalido` | `422` | `VALIDACION` | `rol` | rol inválido |
| `ErrUsuarioEmailNoRegistrado` | `404` | `USUARIO_NO_ENCONTRADO` | — | `No existe un usuario registrado con ese email` |
| `ErrUsuarioYaEsIntegrante` | `409` | `USUARIO_DUPLICADO` | `email` | `Ese usuario ya es integrante` |
| `ErrUsuarioNoAsociado` | `404` | `USUARIO_NO_ENCONTRADO` | — | el usuario no forma parte del proyecto |
| `ErrUltimoLider` | `409` | `ULTIMO_LIDER` | — | el proyecto debe conservar al menos un líder |
| `ErrProyectoCerrado` | `409` | `PROYECTO_CERRADO` | — | proyecto cerrado |
| `ErrAccesoDenegado` | `403` | `ACCESO_DENEGADO` | — | no tenés permisos de líder |
| `ErrProyectoNoEncontrado` | `404` | `PROYECTO_NO_ENCONTRADO` | — | proyecto inexistente o sin membresía |

## Casos límite

- Email con mayúsculas o espacios (`  ADA@UTN.EDU.AR `) se normaliza y encuentra
  al usuario.
- Email con formato `Nombre <correo>`, vacío o con espacios internos → `422`.
- Rol con valor distinto de `LIDER`/`MIEMBRO` → `422`.
- El mismo email enviado dos veces → `409 USUARIO_DUPLICADO`.
- El `LIDER` se degrada a sí mismo: se permite solo si queda otro `LIDER`; si es
  el último, `409 ULTIMO_LIDER`.
- Un `MIEMBRO` que se quita a sí mismo deja de pertenecer: su siguiente request
  recibe `404`.
- Un `usuarioId` inexistente en `PATCH`/`DELETE` → `404 USUARIO_NO_ENCONTRADO`.
- Un `id`/`usuarioId` no numérico o cero → `400 PARAMETRO_INVALIDO`.

## Criterios de aceptación

- **CA-1**: un miembro lista el equipo y recibe `id`, `nombre`, `email` y `rol`
  de cada integrante (`200`).
- **CA-2**: un no miembro recibe `404 PROYECTO_NO_ENCONTRADO` al listar.
- **CA-3**: un `LIDER` agrega un email registrado con rol válido y el usuario
  queda vinculado con ese rol (`200`).
- **CA-4**: un email inexistente responde `404` con el mensaje exacto y no crea
  usuario.
- **CA-5**: un email que ya es integrante responde `409 USUARIO_DUPLICADO`
  (campo `email`) con `Ese usuario ya es integrante`.
- **CA-6**: un email vacío o inválido responde `422 VALIDACION` (campo `email`).
- **CA-7**: un email con mayúsculas/espacios se normaliza y encuentra al usuario.
- **CA-8**: un `MIEMBRO` no líder recibe `403 ACCESO_DENEGADO` en `POST`, `PATCH`
  o `DELETE` y no se modifica nada.
- **CA-9**: un `LIDER` elimina a un integrante (`204`) y el vínculo desaparece.
- **CA-10**: un miembro se quita a sí mismo (`204`) y deja de pertenecer.
- **CA-11**: el último `LIDER` no puede quitarse ni degradarse (`409
  ULTIMO_LIDER`).
- **CA-12**: un `LIDER` cambia el rol de un integrante (`200`) y el rol se
  actualiza.
- **CA-13**: `go test ./...` en verde y escenarios BDD automatizados.

## Tareas de implementación

1. Spec SDD y feature BDD (este documento y su `.feature`).
2. RED: tests de repositorio, service, handler y router de integrantes.
3. GREEN: implementar listar, alta por email, cambio de rol y baja; consumir los
   guards y reglas de US-11.
4. REFACTOR y E2E por curl con líder, miembro y no miembro.
