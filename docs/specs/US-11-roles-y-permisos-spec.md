# Especificación: US-11 - Roles y permisos dentro del proyecto

## Objetivo

Que los miembros de un proyecto tengan **roles con permisos distintos**: el
`LIDER` gestiona el equipo y puede eliminar el proyecto; el `MIEMBRO` trabaja el
contenido (historias, sprints, backlog, defectos, criterios) pero no puede
eliminar el proyecto ni gestionar el equipo.

Reemplaza la decisión de US-F10 ("cualquier miembro gestiona", opción A) y
reutiliza su guard de membresía (`ExigirAccesoAProyecto`) y la identidad del
request (`UsuarioDesdeContexto`). Solo backend: no incluye interfaz (US-F12).

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#69`.
**Bloquea a US-12** (integrantes por email) y a US-F12 (interfaz).

**Trazabilidad:** Issue #69 → SDD → criterios → escenarios BDD
(`docs/features/US-11-roles-y-permisos.feature`) → tests Go (`internal/domain`,
`internal/service`, `internal/handler`, `internal/repository`) → código.

## Cross-referencias

- **US-01**: `DELETE /api/proyectos/{id}` pasa a exigir `LIDER`.
- **US-08**: los 3 roles originales (`Constructor de Producto`,
  `Arquitecto de Producto`, `Facilitador Ágil`) se **reemplazan** por
  `LIDER`/`MIEMBRO`. `POST`/`DELETE /api/proyectos/{id}/usuarios` pasan a exigir
  `LIDER` (la baja a sí mismo es la excepción).
- **US-F10**: el creador queda vinculado como `LIDER` (antes sin rol específico);
  el `404` al no miembro no cambia y se evalúa **antes** que el `403`.
- **US-F11**: editar/cerrar proyectos sigue permitido a cualquier miembro.
- **US-12**: consume `EsLider`, `ErrUltimoLider` y `CambiarRol`; agrega listar,
  alta por email y `PATCH` de rol.
- **US-F12**: interfaz de roles; usa `rol_actual`.

## Entradas

- Roles válidos: `LIDER`, `MIEMBRO` (sensibles a mayúsculas). Cualquier otro valor
  es `422 VALIDACION` (campo `rol`).
- Identidad del request: claim `sub` del JWT (US-F10).
- `POST /api/proyectos/{id}/usuarios` con `rol` omitido o vacío: el usuario entra
  como `MIEMBRO` (US-12 permite elegirlo).

## Reglas de negocio

- **RB-01**: quien crea un proyecto queda vinculado como `LIDER`.
- **RB-02**: un usuario agregado sin rol explícito entra como `MIEMBRO`.
- **RB-03**: solo el `LIDER` puede eliminar el proyecto, agregar integrantes,
  quitar integrantes y cambiar roles. El `MIEMBRO` recibe `403 ACCESO_DENEGADO` y
  no se modifica nada.
- **RB-04**: el `MIEMBRO` conserva todas las operaciones sobre el contenido
  (historias, sprints, sprint backlog, defectos, criterios) y editar/cerrar el
  proyecto.
- **RB-05**: un integrante puede quitarse a sí mismo (`DELETE .../usuarios/{su id}`)
  sea cual sea su rol.
- **RB-06**: no se puede quitar ni degradar al **último** `LIDER`: `409 ULTIMO_LIDER`.
- **RB-07**: un no miembro recibe `404 PROYECTO_NO_ENCONTRADO` (US-F10); el `403`
  solo lo ve un miembro.
- **RB-08**: `GET /api/proyectos/{id}` incluye `rol_actual` (rol del usuario que
  pide).
- **RB-09 (migración)**: al migrar, los roles distintos de `LIDER`/`MIEMBRO` pasan
  a `MIEMBRO`, y en cada proyecto sin líder el integrante de **menor id de usuario**
  (proxy del más antiguo: `proyecto_usuarios` no guarda fecha) pasa a `LIDER`.
  Ningún proyecto con equipo queda sin líder. Es idempotente.

## Salidas esperadas

| HTTP | `error` | Cuándo ocurre |
| :---: | :--- | :--- |
| `200` | — | `GET /api/proyectos/{id}` con `rol_actual` |
| `204` | — | Líder elimina proyecto o integrante; integrante se quita a sí mismo |
| `403` | `ACCESO_DENEGADO` | Miembro no líder intenta eliminar proyecto o gestionar equipo |
| `404` | `PROYECTO_NO_ENCONTRADO` | No miembro (US-F10) |
| `409` | `ULTIMO_LIDER` | Se quita o degrada al único líder |
| `422` | `VALIDACION` (`rol`) | Rol inválido |

## Restricciones y casos límite

- Fuera de alcance: rol `OBSERVADOR`, permisos por tipo de recurso, invitaciones,
  auditoría y frontend.
- Un líder puede quitarse a sí mismo si hay otro líder; si es el único, `409`.
- Un miembro que se quita a sí mismo queda fuera: la siguiente petición recibe `404`.

## Criterios de aceptación

- **CA-1**: crear un proyecto deja al creador como `LIDER`.
- **CA-2**: un líder que agrega a un usuario sin rol lo vincula como `MIEMBRO`.
- **CA-3**: un `MIEMBRO` recibe `403 ACCESO_DENEGADO` al eliminar el proyecto o
  gestionar el equipo, y no cambia nada.
- **CA-4**: un `LIDER` elimina el proyecto y gestiona el equipo.
- **CA-5**: un `MIEMBRO` puede trabajar historias, sprints, backlog, defectos y criterios.
- **CA-6**: un no miembro recibe `404`; los tests de US-F10 siguen en verde.
- **CA-7**: el último líder no puede quitarse ni degradarse (`409 ULTIMO_LIDER`).
- **CA-8**: un `MIEMBRO` puede quitarse a sí mismo.
- **CA-9**: `GET /api/proyectos/{id}` devuelve `rol_actual`.
- **CA-10**: tras la migración ningún proyecto con equipo queda sin líder.
