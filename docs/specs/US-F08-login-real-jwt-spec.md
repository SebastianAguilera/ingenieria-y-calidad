# Especificación: US-F08 - Login real con email y contraseña (JWT)

## Objetivo

Reemplazar el ingreso simulado de US-F02 por una **autenticación real**: la pantalla
pide email y contraseña, el frontend llama a `POST /api/auth/login` (US-10), guarda el
JWT devuelto y lo envía en `Authorization: Bearer` en cada request a `/api/*`. Si el
backend responde 401 en cualquier llamada, la sesión se limpia y se vuelve a `/login`.

La funcionalidad satisface el issue `SebastianAguilera/ingenieria-y-calidad#62` y
**reemplaza** a US-F02 (`#51`, login simulado por búsqueda de email/nombre), que queda
como registro histórico. Dependencias resueltas: US-10 (`#61`, backend JWT), US-F01
(`#50`, cliente HTTP tipado) y US-F02 (`#51`, pantalla, `SesionContext` y
`RutaProtegida`). Afecta a US-F03 en adelante (menú, "Salir", layouts del panel).

**Trazabilidad:** ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de
`docs/features/US-F08-login-real-jwt.feature` son el origen de las pruebas Vitest.

## Entradas

### Precondiciones

- Backend US-10 disponible y verificado (commits `a0d6ba6`, `78cef9b`, `c0666ca`):
  login real y middleware sobre `/api/*`.
- Cliente HTTP tipado de US-F01 (`frontend/src/api/client.ts`).
- Sesión heredada de US-F02: `SesionContext` con persistencia en `localStorage`
  (clave `app.sesion.v1`), `RutaProtegida` y botón "Salir" en `AppShell`.
- Frontend Vite en `http://localhost:5173` con proxy `/api` → `http://localhost:8080`.

### Pantalla `/login`

| Campo | Tipo | Requerido | Regla |
| :--- | :--- | :---: | :--- |
| Email | string | Sí | Se recorta (trim) antes de enviar; la normalización final (minúsculas) la hace el backend |
| Contraseña | `password` | Sí | Nunca se guarda ni se loguea; solo viaja en el cuerpo del request |

> Se elimina el campo "Email o nombre" y la vista de registro inline de US-F02.

### Contrato del backend (US-10)

**`POST /api/auth/login`** (ruta pública): cuerpo `{"email": "...", "password": "..."}`.

| HTTP | `error` | `mensaje` del backend | Cuándo ocurre | Uso en el frontend |
| :---: | :--- | :--- | :--- | :--- |
| 200 | — | — | Credenciales válidas | Guardar `token` + `usuario` y navegar a `/proyectos` |
| 400 | `JSON_INVALIDO` | el cuerpo no es un objeto JSON válido | Body malformado | No debe ocurrir: el cliente envía JSON |
| 422 | `VALIDACION` | mensaje del backend + `campo` (`email` / `password`) | Campo vacío o inválido sin llegar a bcrypt | Mostrar mensaje y resaltar el campo |
| 401 | `CREDENCIALES_INVALIDAS` | `email o contraseña incorrectos` | Contraseña mala, email inexistente o usuario sin hash (mismo mensaje exacto, anti enumeración) | Mensaje genérico; **no** se guarda token ni se limpia nada |
| 401 | `TOKEN_REQUERIDO` | `la solicitud no incluye un token de autenticación` | Middleware: falta `Authorization` o no empieza con `Bearer ` | Limpiar sesión → `/login` |
| 401 | `TOKEN_INVALIDO` | `el token de autenticación no es válido o ha expirado` | Firma incorrecta, token expirado, algoritmo distinto de HS256 o token malformado | Limpiar sesión → `/login` |
| red | `NETWORK_ERROR` | — (el detalle lo tiene el cliente) | `fetch` sin respuesta | `No se pudo conectar con el servidor.` + reintentar |

- Respuesta 200: `{"token": "<JWT HS256>", "usuario": {id, nombre, email, creado_en}}`.
  **Nunca** incluye `password_hash`.
- Claims del token: `sub` (id como texto), `email`, `iat`, `exp = iat + JWT_EXPIRY`
  (default `24h`). Sin logout server-side ni blacklist: un token válido sirve hasta
  expirar y dos sesiones simultáneas son válidas.
- Rutas públicas (las únicas): `GET /health` y `POST /api/auth/login`. Todo lo demás
  de `/api/*` exige `Authorization: Bearer <token>`.

## Salidas esperadas

| # | Acción | Resultado | Efecto |
| :---: | :--- | :--- | :--- |
| 1 | Credenciales correctas | Sesión iniciada con token | Navega a `/proyectos` |
| 2 | Contraseña incorrecta / email inexistente | `Email o contraseña incorrectos` | Sin token guardado |
| 3 | Campo vacío | Mensaje de validación en el campo | Sin llamada a la API |
| 4 | Recarga con sesión activa | Sesión restaurada desde `localStorage` | Sigue en el panel |
| 5 | Acción "Salir" | Token y usuario eliminados | Vuelve a `/login` |
| 6 | Token vencido o inválido en cualquier request | Sesión limpia | Vuelve a `/login` |
| 7 | Servidor sin respuesta | `No se pudo conectar con el servidor.` | Pantalla disponible para reintentar |

## Reglas de negocio

- **RB-01** Sin token en la sesión, toda ruta del panel redirige a `/login`. → **CA-1**
- **RB-02** El login llama exclusivamente a `POST /api/auth/login` con `email` y
  `password`; no se usa `GET /api/usuarios` ni ninguna búsqueda por nombre. → **CA-2, CA-3**
- **RB-03** El `token` y el `usuario` se persisten juntos en `localStorage`; el token
  se adjunta como `Authorization: Bearer <token>` en toda request a `/api/*`. → **CA-2, CA-5**
- **RB-04** Un `401 CREDENCIALES_INVALIDAS` en el login muestra
  `Email o contraseña incorrectos`, no guarda token y no navega. → **CA-3**
- **RB-05** Los campos vacíos se validan en el cliente antes de enviar: no se emite
  ninguna petición. → **CA-4**
- **RB-06** Al recargar la página la sesión se reconstruye desde `localStorage`; si el
  almacenamiento está vacío, corrupto o con un formato inválido, se ignora y se entra
  sin sesión. → **CA-5**
- **RB-07** "Salir" borra token y usuario del estado y del `localStorage` y redirige a
  `/login`. → **CA-6**
- **RB-08** Un `401 TOKEN_REQUERIDO` o `401 TOKEN_INVALIDO` en **cualquier** request
  limpia la sesión y redirige a `/login` (no se reintentan requests con el mismo token). → **CA-7**
- **RB-09** Un error de red muestra `No se pudo conectar con el servidor.` con opción
  de reintentar, sin perder el formulario. → **CA-8**
- **RB-10** El frontend distingue los dos orígenes del 401: en el login,
  `CREDENCIALES_INVALIDAS` es un error de credenciales (mensaje en la pantalla); en el
  resto de las rutas, `TOKEN_REQUERIDO` / `TOKEN_INVALIDO` son sesión vencida (limpiar
  y salir). Nunca se limpia la sesión por un 401 recibido en el propio login. → **CA-3, CA-7**
- **RB-11** La contraseña viaja solo en el cuerpo de `POST /api/auth/login`: no se
  guarda en `localStorage`, ni en estado de React persistido, ni en la URL, ni en logs
  del cliente. → **CA-2**

## Restricciones

- El token se guarda en `localStorage` (persistencia elegida en US-02 de la historia;
  no se usan cookies ni storage de sesión del navegador).
- Toda la lógica de sesión vive en `SesionContext`/`sesion.ts`: las pantallas no
  acceden a `localStorage` de forma directa.
- `client.ts` no debe importar React: si necesita el token, lo obtiene del módulo de
  sesión (o de un proveedor inyectado) para poder testearse con Vitest.
- El backend no expone logout ni revocación: "Salir" es puramente del lado cliente.
- Los textos de interfaz van en español.
- Los tests usan Vitest + React Testing Library con la API mockeada
  (`frontend/src/**/*.test.ts(x)`).

## Casos límite

- Email con espacios o mayúsculas: el frontend recorta; la normalización completa es
  del backend (trim + minúsculas).
- Contraseña vacía y email vacío en el mismo envío: manda la validación del primer
  campo detectado y no se envía nada.
- `localStorage` con JSON inválido, o con el formato anterior de US-F02 (solo usuario,
  sin token): se ignora y se inicia sin sesión.
- Token presente pero vencido: el primer request del panel devuelve 401 → sesión
  limpia → `/login`.
- Dos pestañas con sesiones distintas: cada una lee su propio `localStorage` (gana la
  última escritura; sin sincronización entre pestañas, fuera de alcance).
- El login se dispara dos veces (doble submit): el botón queda deshabilitado mientras
  `cargando` está activo.

## Condiciones de error

| Situación | Origen | Mensaje de interfaz |
| :--- | :--- | :--- |
| Campo email o contraseña vacío | Frontend | Mensaje de validación en el campo, sin request |
| Credenciales incorrectas | `401 CREDENCIALES_INVALIDAS` | `Email o contraseña incorrectos` |
| Token ausente/vencido en el panel | `401 TOKEN_REQUERIDO` / `401 TOKEN_INVALIDO` | Sin mensaje: se limpia la sesión y se navega a `/login` |
| Validación del backend | `422 VALIDACION` | Mensaje del backend y campo resaltado |
| Error de red | `NETWORK_ERROR` | `No se pudo conectar con el servidor.` |
| Respuesta inesperada | `500 ERROR_INTERNO` / `RESPUESTA_INVALIDA` | `Ocurrió un error inesperado. Intentá de nuevo.` |

## Criterios de aceptación

- [ ] **CA-1** Given no hay token en la sesión, When intento abrir cualquier ruta del panel, Then soy redirigido a `/login`.
- [ ] **CA-2** Given estoy en `/login`, When ingreso email y contraseña válidos, Then la API responde 200, se guardan token y usuario y soy redirigido a `/proyectos` (sin `password_hash` en la respuesta).
- [ ] **CA-3** Given estoy en `/login`, When el backend responde `401 CREDENCIALES_INVALIDAS`, Then veo `Email o contraseña incorrectos` y no se guarda ningún token.
- [ ] **CA-4** Given dejo un campo vacío, When intento ingresar, Then veo el mensaje de validación y no se envía la petición a la API.
- [ ] **CA-5** Given inicié sesión correctamente, When recargo la página, Then el token se recupera de `localStorage`, sigo logueado y puedo volver al panel.
- [ ] **CA-6** Given estoy en el panel con sesión iniciada, When presiono "Salir", Then se borran token y usuario y vuelvo a `/login`.
- [ ] **CA-7** Given tengo una sesión con token vencido o inválido, When hago cualquier request al panel, Then el 401 limpia la sesión y termino en `/login`.
- [ ] **CA-8** Given el servidor no responde, When intento ingresar, Then veo `No se pudo conectar con el servidor.` y puedo reintentar sin perder lo escrito.

## Archivos afectados

| Archivo | Cambio |
| :--- | :--- |
| `frontend/src/api/auth.ts` | **Nuevo**: `login(email, password)` → `POST /api/auth/login` con el cliente tipado |
| `frontend/src/api/auth.test.ts` | **Nuevo**: casos de la capa API (200, 401, 422, red) |
| `frontend/src/api/client.ts` | Adjuntar `Authorization: Bearer <token>` y distinguir el 401 de sesión vencida |
| `frontend/src/api/client.test.ts` | Cubrir los nuevos encabezados y el manejo del 401 |
| `frontend/src/session/sesion.ts` | Guardar `{token, usuario}`; exponer `token`; validación del formato nuevo |
| `frontend/src/session/SesionContext.tsx` | `ingresar(token, usuario)`, `salir()`, limpieza ante 401 |
| `frontend/src/session/SesionContext.test.tsx` | Nuevos casos de persistencia y limpieza |
| `frontend/src/components/RutaProtegida.tsx` | Exigir **token** (no solo usuario activo) |
| `frontend/src/components/RutaProtegida.test.tsx` | Actualizar expectativas a token |
| `frontend/src/pages/LoginPage.tsx` | Reescribir: email + contraseña, sin registro inline ni búsqueda por nombre |
| `frontend/src/pages/LoginPage.test.tsx` | Reescribir con los escenarios de US-F08 |
| `frontend/src/layouts/AppShell.tsx` | "Salir" borra también el token (vía `salir()`) |
| `frontend/src/api/usuarios.ts` | `buscarUsuario` deja de usarse en el login (queda para US-F17 si se necesita) |

`frontend/src/App.tsx` no cambia: las rutas ya están envueltas en `RutaProtegida`.

## Fuera de alcance

- Registro de usuarios desde el login (US-F17 / alta en US-08).
- Refresh tokens, recuperación de contraseña, revocación server-side de tokens.
- MFA/SSO, rate limiting, sincronización de sesión entre pestañas.
- Cambios en el backend (US-10 está cerrada y verificada con 167 tests).
- Menú, layout y navegación del panel (US-F03) y el resto de las pantallas.

## Nota de trazabilidad

La issue `#62` fue cerrada como `COMPLETED` **sin implementación ni PR**. El cierre
definitivo debe quedar ligado al PR que entregue esta historia (`Closes #62`), con el
checklist de la issue marcado.

## Escenarios BDD

Ver `docs/features/US-F08-login-real-jwt.feature`.
