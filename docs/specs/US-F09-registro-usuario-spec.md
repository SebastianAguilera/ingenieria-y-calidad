# Especificación: US-F09 - Registro de usuario con contraseña desde el login

## Objetivo

Dar de alta un usuario **desde la pantalla de login** con nombre, email y contraseña, y
entrar al panel de una vez (auto-login). Se resuelve con un endpoint **público**
`POST /api/auth/registro` que valida los campos, persiste el usuario con hash bcrypt
y responde `{token, usuario}` igual que el login; en `/login` queda un link
**"Crear cuenta"** siempre visible que abre el formulario.

La historia satisface el issue `SebastianAguilera/ingenieria-y-calidad#65`
(prioridad `P2`, milestone Sprint 2) y cierra el hueco que dejó US-F08 al eliminar el
alta simulada de US-F02. Dependencias: US-10 (JWT, middleware y lista pública de
rutas), US-F08 (pantalla de login, `SesionContext`, `RutaProtegida`), US-F01 (cliente
HTTP tipado). Complementa a US-08, que sigue siendo el alta administrativa
(protegida, sin contraseña).

**Trazabilidad:** ciclo TDD (RED → GREEN → REFACTOR) en backend Go y frontend. Los
escenarios BDD de `docs/features/US-F09-registro-usuario.feature` son el origen de
los tests de `internal/service`, `internal/handler`, `cmd/api` y de los tests
Vitest del frontend.

## Entradas

### Precondiciones

- Backend US-10 en marcha: middleware default-deny con lista pública
  (`GET /health`, `POST /api/auth/login`).
- Tabla `usuarios` con índice único en `email` y columna `password_hash varchar(72)`
  (US-10); el repositorio traduce la clave duplicada en `domain.ErrUsuarioDuplicado`.
- Pantalla `/login` de US-F08 con `SesionContext` (token + usuario en
  `localStorage`, clave `app.sesion.v1`).
- Cliente HTTP tipado de US-F01 con `Authorization: Bearer` y limpieza de sesión
  ante `401 TOKEN_*`.

### Pantalla "Crear cuenta" (vista de registro en `/login`)

| Campo | Tipo | Requerido | Regla |
| :--- | :--- | :---: | :--- |
| Nombre | string | Sí | Se recorta (trim); 1 a 100 caracteres |
| Email | string | Sí | Se recorta; se normaliza a minúsculas en el backend |
| Contraseña | `password` | Sí | Nunca se guarda en claro ni en el cliente |

### Contrato del backend

**`POST /api/auth/registro`** (ruta pública, sin `Authorization`):
cuerpo `{"nombre": "...", "email": "...", "password": "..."}`.

| HTTP | `error` | `mensaje` del backend | Cuándo ocurre | Uso en el frontend |
| :---: | :--- | :--- | :--- | :--- |
| 200 | — | — | Alta exitosa | Guardar `token` + `usuario` y navegar a `/proyectos` |
| 400 | `JSON_INVALIDO` | el cuerpo no es un objeto JSON válido | Body malformado | No debe ocurrir: el cliente envía JSON |
| 422 | `VALIDACION` | mensaje del backend + `campo` (`nombre` / `email` / `password`) | Campo vacío, nombre > 100 caracteres o email con formato inválido | Mostrar mensaje y resaltar el campo |
| 422 | `USUARIO_DUPLICADO` | `el email ya está registrado` + `campo: email` | El email ya existe (índice único) | `Ese email ya está registrado. Volvé a ingresar.` y no se guarda token |
| red | `NETWORK_ERROR` | — (el detalle lo tiene el cliente) | `fetch` sin respuesta | `No se pudo conectar con el servidor.` + reintentar |

- Respuesta 200: `{"token": "<JWT HS256>", "usuario": {id, nombre, email, creado_en}}`,
  con los mismos claims que el login (`sub`, `email`, `iat`, `exp`).
  **Nunca** incluye `password_hash` ni la contraseña en claro.
- Orden de validación (espejo del login, R-08): JSON (400) → campos (422) →
  duplicado (422) → éxito (200).

## Salidas esperadas

| # | Acción | Resultado | Efecto |
| :---: | :--- | :--- | :--- |
| 1 | Clic en "Crear cuenta" | Se abre el formulario de registro | Sin llamada a la API |
| 2 | Registro completo y válido | `200` con token y usuario | Sesión iniciada en `/proyectos` |
| 3 | Email ya registrado | `422 USUARIO_DUPLICADO` | Mensaje en pantalla, sin token |
| 4 | Campo vacío o inválido | Validación en el cliente / `422 VALIDACION` | Sin alta, campo resaltado |
| 5 | Clic en "Volver" | Se muestra el login | Sin alta ni petición |
| 6 | Servidor caído | `NETWORK_ERROR` | Mensaje + reintentar |
| 7 | Registro exitoso y luego login | `200` en `POST /api/auth/login` | El hash quedó persistido |

## Reglas de negocio

- **RB-01** En `/login` el link "Crear cuenta" está siempre visible; mostrarlo no
  dispara ninguna petición. → **CA-1**
- **RB-02** El alta llama exclusivamente a `POST /api/auth/registro` con `nombre`,
  `email` y `password`; nunca a `POST /api/usuarios` (protegida y sin contraseña). → **CA-2, CA-4**
- **RB-03** El backend valida (nombre, email, password), guarda el usuario con hash
  bcrypt y devuelve token + usuario; el frontend guarda la sesión y navega a
  `/proyectos`. → **CA-2**
- **RB-04** Los campos vacíos se validan en el cliente antes de enviar: no se emite
  ninguna petición. → **CA-3**
- **RB-05** Un `422 USUARIO_DUPLICADO` muestra `Ese email ya está registrado.
  Volvé a ingresar.` y no guarda token. → **CA-4**
- **RB-06** Un `422 VALIDACION` muestra el mensaje del backend con el campo
  (`nombre`, `email` o `password`) resaltado. → **CA-5**
- **RB-07** "Volver" regresa al login sin crear el usuario y sin petición. → **CA-7**
- **RB-08** Un error de red muestra `No se pudo conectar con el servidor.` con
  opción de reintentar, sin perder lo escrito. → **CA-8**
- **RB-09** La ruta es pública: un request **sin** `Authorization` nunca recibe
  `401 TOKEN_REQUERIDO`. → **CA-2**
- **RB-10** El email se normaliza (trim + minúsculas) antes de persistir; el índice
  único de la tabla es el que decide el duplicado. → **CA-4, CA-6**
- **RB-11** La contraseña solo viaja en el cuerpo del registro: se persiste
  únicamente como hash bcrypt y no se guarda en el navegador. → **CA-2, CA-6, CA-12**
- **RB-12** A diferencia del login (US-F08), el registro **sí** revela si el email
  ya existe: es el tradeoff aceptado de un alta pública. → **CA-4**

## Restricciones

- `POST /api/auth/registro` se agrega a la lista blanca del middleware junto a
  `/health` y `/api/auth/login`; todo lo demás sigue default-deny.
- El alta reutiliza la validación de usuario de US-08 (`validarUsuario`) y el
  mapeo de errores existente (`ErrUsuarioDuplicado` → 422 `USUARIO_DUPLICADO`).
- Costo bcrypt `DefaultCost`, el mismo que US-10.
- La vista de registro vive en `LoginPage` (no se crea una ruta nueva): se cambia
  el estado local `vista`, como en US-F02.
- Textos de interfaz en español; tests con Vitest + React Testing Library y mocks
  de la API.

## Casos límite

- Email con espacios o mayúsculas: se recorta en el cliente y se normaliza en el
  backend; `ANA@Example.com ` y `ana@example.com` son el mismo usuario.
- Nombre vacío, con un solo espacio o de más de 100 caracteres: `422 VALIDACION`
  con campo `nombre`.
- Email con formato inválido o de más de 150 caracteres: `422 VALIDACION` con
  campo `email`.
- Contraseña vacía: validación en el cliente (sin request) y `422 VALIDACION`
  con campo `password` si llega al backend.
- Dos registros simultáneos con el mismo email: gana el índice único y el otro
  recibe `422 USUARIO_DUPLICADO`.
- Doble envío del formulario: el botón queda deshabilitado mientras `cargando`
  está activo (una sola petición).
- Sesión ya activa: `/login` sigue redirigiendo a `/proyectos` (no se ofrece
  registro con sesión iniciada).
- Sin política de complejidad de contraseñas: se exige que no esté vacía, igual
  que en el login.

## Condiciones de error

| Situación | Origen | Mensaje de interfaz |
| :--- | :--- | :--- |
| Campo nombre, email o contraseña vacío/inválido | Frontend o `422 VALIDACION` | Mensaje de validación con el campo resaltado, sin alta |
| Email ya registrado | `422 USUARIO_DUPLICADO` | `Ese email ya está registrado. Volvé a ingresar.` |
| Body malformado | `400 JSON_INVALIDO` | `Ocurrió un error inesperado. Intentá de nuevo.` |
| Error de red | `NETWORK_ERROR` | `No se pudo conectar con el servidor.` |
| Error inesperado del backend | `500 ERROR_INTERNO` / `RESPUESTA_INVALIDA` | `Ocurrió un error inesperado. Intentá de nuevo.` |

## Criterios de aceptación

- [ ] **CA-1** Given estoy en `/login`, When miro el formulario, Then veo siempre el link "Crear cuenta" que abre el registro sin enviar ninguna petición.
- [ ] **CA-2** Given completo nombre, email y contraseña válidos, When envío el alta, Then el backend responde 200 con token y usuario, se guarda la sesión y soy redirigido a `/proyectos`.
- [ ] **CA-3** Given dejo un campo vacío, When intento registrar, Then veo el mensaje de validación en el campo y no se envía la petición.
- [ ] **CA-4** Given el email ya está registrado, When el backend responde `422 USUARIO_DUPLICADO`, Then veo "Ese email ya está registrado. Volvé a ingresar." y no se guarda ningún token.
- [ ] **CA-5** Given un email con formato inválido, When el backend responde `422 VALIDACION` con campo `email`, Then veo el mensaje con el campo resaltado.
- [ ] **CA-6** Given el registro se completó, When después intento entrar con ese email y contraseña, Then el login devuelve 200 (el hash quedó persistido).
- [ ] **CA-7** Given estoy en el registro, When presiono "Volver", Then regreso al login sin crear nada.
- [ ] **CA-8** Given el servidor no responde, When envío el alta, Then veo "No se pudo conectar con el servidor." y puedo reintentar.

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `internal/service/auth_service.go` | Nuevo `Registrar(ctx, nombre, email, password)`: validación, hash bcrypt, alta y token |
| `internal/service/auth_service_test.go` | Tests RED/GREEN del registro (éxito, duplicado, validaciones, hash) |
| `internal/handler/auth_handler.go` | `ServicioAuth` gana `Registrar`; ruta `POST /api/auth/registro` y su handler |
| `internal/handler/auth_handler_test.go` | Tests del endpoint (200, 400, 422 por campo, 422 duplicado) |
| `cmd/api/router.go` | `/api/auth/registro` se agrega a la lista pública del middleware |
| `cmd/api/router_auth_test.go` | Test de ruta pública sin token + doble `servicioAuthPrueba.Registrar` |
| `frontend/src/api/auth.ts` | Nueva `registrar(nombre, email, password)` → `POST /api/auth/registro` |
| `frontend/src/api/auth.test.ts` | Tests de la capa API del registro |
| `frontend/src/pages/LoginPage.tsx` | Vista "Crear cuenta" (siempre visible) con validación y auto-login |
| `frontend/src/pages/LoginPage.test.tsx` | Escenarios E-01 a E-08 del registro |

No se modifica `internal/repository` ni el esquema: el índice único y la columna
`password_hash` ya existen (US-10).

## Dependencias

- **US-10** (`#61`): JWT, middleware y lista de rutas públicas.
- **US-F08** (`#65` es la continuación; la historia de login es `#62`): pantalla
  de login, sesión con token y `RutaProtegida`.
- **US-F01** (`#50`): cliente HTTP tipado y manejo central de errores.
- **US-08** (`#39`): validación de usuario y alta administrativa (se complementa,
  no se reemplaza).

## Fuera de alcance

- Recuperación de contraseña, verificación de email por envío, edición y baja de
  usuarios.
- Política de complejidad o expiración de contraseñas, MFA.
- Cambiar el alta administrativa de US-08 (`POST /api/usuarios`).
- Cambios en el login de US-F08 (incluida la anti-enumeración).

## Nota de trazabilidad

La issue `#65` se cierra con el PR que entregue esta historia (`Closes #65`),
con su checklist marcado.

## Escenarios BDD

Ver `docs/features/US-F09-registro-usuario.feature`.
