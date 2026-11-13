# Especificación: US-08 - Gestión independiente de usuarios y asociación a proyectos con roles

## Objetivo

Implementar la **gestión independiente de usuarios** y la **asociación con rol** al proyecto. Las identidades de las personas pasan a ser entidades globales (`usuarios`) registradas de forma autónoma. El alta de proyectos deja de recibir usuarios anidados: `POST /api/proyectos` crea únicamente el proyecto sin miembros asociados. Los usuarios se registran con `POST /api/usuarios` y se asocian individualmente al proyecto con `POST /api/proyectos/{id}/usuarios`, validando que el usuario ya exista.

> **Actualización US-12 (#68):** el alta de integrantes pasa a ser por **email**
> (`{ "email": "...", "rol": "..." }`) para que el líder no necesite conocer el
> `usuario_id`. Se agregan `GET` y `PATCH` de integrantes. Ver
> `docs/specs/US-12-integrantes-por-email-spec.md`.

La funcionalidad satisface el issue `SebastianAguilera/ingenieria-y-calidad#39` y se alinea con los roles definidos en Scrum (Product Architect, Agile Enabler, Product Builders).

**Trazabilidad:** ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de `docs/features/US-08-gestion-usuarios.feature` son el origen de las pruebas. Cada criterio de aceptación es verificable por un test concreto.

## Entradas

### Precondiciones
- Esquema `usuarios`, `proyectos`, `proyecto_usuarios` existente (generado por `AutoMigrate`).
- La autenticación la aporta US-10 (JWT) y el acceso por membresía US-F10: para `POST/DELETE /api/proyectos/{id}/usuarios` el actor debe ser miembro del proyecto (si no, el guard responde `404 PROYECTO_NO_ENCONTRADO`), de modo que nadie puede auto-insertarse en un proyecto ajeno.

### Convenciones
| Convención | Regla | Consecuencia |
| :--- | :--- | :--- |
| Email | Único, máximo 150, normalizado a minúsculas/sin espacios | `422 VALIDACION` campo `email` si duplicado o inválido |
| Rol | `"LIDER"` o `"MIEMBRO"` (US-11 reemplaza los 3 roles originales; si se omite, `"MIEMBRO"`) | `422 VALIDACION` campo `rol` si es inválido |
| Usuario existente | Al asociar debe existir en `usuarios` | `404 USUARIO_NO_ENCONTRADO` si no existe |
| Asociación | No duplicada por proyecto | `409 USUARIO_DUPLICADO` (US-12) |
| Proyecto | Debe existir y estar activo para asociar/desasociar | `404 PROYECTO_NO_ENCONTRADO` / `409 PROYECTO_CERRADO` |

### Cuerpos de request

#### `POST /api/usuarios` — registrar usuario independiente
| Campo | Tipo | Requerido | Formato |
| :--- | :--- | :---: | :--- |
| `nombre` | string | Sí | 1 a 100 caracteres tras `TrimSpace` |
| `email` | string | Sí | Formato válido, único |

```json
{ "nombre": "Ana Pérez", "email": "dev@example.com" }
```

#### `GET /api/usuarios` — listar usuarios
Sin body.

#### `POST /api/proyectos/{id}/usuarios` — asociar usuario existente con rol
| Campo | Tipo | Requerido | Formato |
| :--- | :--- | :---: | :--- |
| `email` | string | Sí | Email válido de un usuario registrado (US-12) |
| `rol` | string | No | `LIDER` o `MIEMBRO`; omitido/vacío → `MIEMBRO` |

```json
{ "email": "ada@utn.edu.ar", "rol": "MIEMBRO" }
```

#### `POST /api/proyectos` — crear proyecto (desacoplado)
Sin campo `usuarios`. Ver `docs/specs/US-01-gestion-de-proyectos-spec.md` (actualizado por esta historia).

## Salidas esperadas

### Diseño de API REST
| # | Método | Path | Descripción | Éxito | Errores |
| :---: | :--- | :--- | :--- | :---: | :--- |
| 1 | `POST` | `/api/usuarios` | Registro independiente | `201 Created` | `400`, `422`, `500` |
| 2 | `GET` | `/api/usuarios` | Listado de usuarios | `200 OK` | `500` |
| 3 | `POST` | `/api/proyectos/{id}/usuarios` | Asociar usuario existente con rol | `200 OK` | `400`, `404`, `409`, `422`, `500` |

## Reglas de negocio
- **RB-01** El email de usuario debe ser único en todo el sistema.
- **RB-02** Al asociar, el usuario debe existir (`POST /api/usuarios` previo).
- **RB-03** El rol debe pertenecer al conjunto cerrado de roles de proyecto.
- **RB-04** Un usuario no puede asociarse dos veces al mismo proyecto.
- **RB-05** `POST /api/proyectos` no acepta ni crea usuarios anidados.
- **RB-06** El rol se persiste en la tabla puente `proyecto_usuarios`.

## Errores
| Error de dominio | Código HTTP | `error` | `campo` |
| :--- | :---: | :--- | :--- |
| `ErrUsuarioDuplicado` | 422 | `USUARIO_DUPLICADO` | `email` |
| `ErrUsuarioNoEncontrado` | 404 | `USUARIO_NO_ENCONTRADO` | `""` |
| `ErrRolInvalido` | 422 | `VALIDACION` | `rol` |
| `ErrUsuarioYaEsIntegrante` | 409 | `USUARIO_DUPLICADO` | `email` |
| `ErrUsuarioEmailNoRegistrado` | 404 | `USUARIO_NO_ENCONTRADO` | `""` |
| `ErrUsuarioEmailInvalido`, `ErrUsuarioNombreVacio/Largo` | 422 | `VALIDACION` | `email`/`nombre` |

## Criterios de aceptación
- [ ] Se puede registrar un usuario independiente por `POST /api/usuarios` (201).
- [ ] No se permite registrar dos usuarios con el mismo email (422).
- [ ] `GET /api/usuarios` devuelve la lista de usuarios registrados.
- [ ] `POST /api/proyectos` crea el proyecto sin miembros asociados (sin `usuarios` en body).
- [ ] `POST /api/proyectos/{id}/usuarios` asocia un usuario existente con rol (200) y persiste el rol.
- [ ] Asociar un usuario no registrado devuelve 404.
- [ ] Asociar con rol inválido devuelve 422 campo `rol`.

## Escenarios BDD
Ver `docs/features/US-08-gestion-usuarios.feature`.
