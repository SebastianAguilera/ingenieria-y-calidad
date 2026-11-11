# Especificación: US-F02 - Ingreso al panel con un usuario existente

> **HISTÓRICA — reemplazada por US-F08** (`docs/specs/US-F08-login-real-jwt-spec.md`,
> issue `#62`). El ingreso simulado de esta historia fue sustituido por el login real
> con email y contraseña (JWT). Se conserva como registro de lo implementado en
> `feature/US-F02-ingreso-con-usuario-existente`; sus criterios de aceptación ya no
> describen el comportamiento vigente de la pantalla de login.

## Objetivo

Permitir que un integrante del equipo ingrese al panel escribiendo su email o su nombre,
sin elegir de una lista. Se trata de un **login simulado**: no hay contraseña ni
verificación real de identidad. Si el usuario existe en la API queda como usuario
activo; si no existe, se ofrece crearlo y luego ingresar.

La funcionalidad satisface el issue `SebastianAguilera/ingenieria-y-calidad#51` y es
prerrequisito de US-F03 (rutas y menú del proyecto), US-F04 (listado de proyectos) y
US-F06 (detalle de proyecto).

**Trazabilidad:** ciclo TDD (RED → GREEN → REFACTOR). Los escenarios BDD de
`docs/features/US-F02-ingreso-panel.feature` son el origen de las pruebas Vitest.

## Entradas

### Precondiciones
- Backend disponible con `GET /api/usuarios` y `POST /api/usuarios` (US-08).
- Cliente HTTP tipado de US-F01 (`frontend/src/api/client.ts`).
- No hay autenticación real: no se usan contraseñas, tokens ni sesiones de servidor.

### Pantalla `/login`
| Campo | Tipo | Requerido | Regla |
| :--- | :--- | :---: | :--- |
| Identificador | string | Sí | Email o nombre, sin espacios al inicio/fin |
| Nombre (registro) | string | Sí | 1 a 100 caracteres |
| Email (registro) | string | Sí | Formato válido y único |

### Búsqueda del usuario
- Se consulta `GET /api/usuarios`, que devuelve un arreglo de usuarios
  (`id`, `nombre`, `email`, `creado_en`).
- Primero se compara el identificador contra `email` normalizado (minúsculas, sin espacios).
- Si no hay coincidencia por email, se compara contra `nombre` normalizado
  (minúsculas, sin espacios) de forma exacta.
- Si hay exactamente un usuario por nombre, se ingresa con él.
- Si hay más de un usuario con el mismo nombre, se considera **ambiguo** y se pide email.

## Salidas esperadas

| # | Acción | Resultado | Efecto |
| :---: | :--- | :--- | :--- |
| 1 | Identificador vacío | Mensaje `Ingresá tu email` | Sin llamada a la API |
| 2 | Usuario encontrado | Sesión iniciada | Navega a `/proyectos` |
| 3 | Usuario no encontrado | Tarjeta "Usuario no encontrado" + formulario de registro | Ninguno |
| 4 | Registro exitoso | Sesión iniciada con el usuario creado | Navega a `/proyectos` |
| 5 | Email duplicado | Mensaje de email ya registrado | Sin sesión |
| 6 | Nombre ambiguo | Mensaje para ingresar email | Sin sesión |
| 7 | Recarga del panel | Sesión conservada | Sigue en el panel |
| 8 | Acción "Salir" | Sesión eliminada | Vuelve a `/login` |

## Reglas de negocio

- **RB-01** Sin usuario activo, toda ruta del panel redirige a `/login`.
- **RB-02** El usuario activo se persiste en `localStorage` (solo sus datos públicos).
- **RB-03** El email es único y se compara sin distinguir mayúsculas ni espacios.
- **RB-04** La coincidencia por nombre es exacta; no se admiten coincidencias parciales.
- **RB-05** Si el nombre coincide con más de un usuario, se debe pedir email.
- **RB-06** El registro usa `POST /api/usuarios` con `nombre` y `email`.
- **RB-07** Tras un ingreso o registro exitoso se navega a `/proyectos`.
- **RB-08** "Salir" limpia la sesión y redirige a `/login`.
- **RB-09** No se guardan credenciales en `localStorage`, solo el usuario activo.

## Restricciones

- Sin autenticación real: no hay contraseña, token ni verificación de identidad.
- La búsqueda se resuelve en el frontend; no se crea un endpoint dedicado.
- El cliente no debe llamar a la API con el campo vacío.
- Los errores de red o del backend deben mostrarse sin romper la pantalla.
- Los textos de interfaz van en español.

## Casos límite

- Identificador con espacios o mayúsculas: se normaliza antes de comparar.
- Dos usuarios con el mismo nombre y distinto email: se pide email.
- Usuario recién creado: queda como usuario activo sin volver a buscar.
- `localStorage` con JSON inválido o incompleto: se ignora y se inicia sin sesión.
- Error de red durante el login o el registro: mensaje visible con opción de reintentar.

## Condiciones de error

| Situación | Origen | Mensaje de interfaz |
| :--- | :--- | :--- |
| Campo vacío | Frontend | `Ingresá tu email` |
| Nombre ambiguo | Frontend | `Hay más de un usuario con ese nombre. Ingresá tu email.` |
| Usuario inexistente | Búsqueda | Tarjeta "Usuario no encontrado" con formulario |
| Email duplicado | `422 USUARIO_DUPLICADO` | `Ese email ya está registrado. Volvé a ingresar.` |
| Validación de nombre/email | `422 VALIDACION` | Mensaje del backend y campo resaltado |
| Error de red | `NETWORK_ERROR` | `No se pudo conectar con el servidor.` |
| Error interno | `500 ERROR_INTERNO` | Mensaje genérico y opción de reintentar |

## Criterios de aceptación

- [ ] Given no hay usuario activo, When entro a cualquier ruta del panel, Then soy redirigido a `/login`.
- [ ] Given estoy en `/login`, When ingreso un email que existe, Then se guarda el usuario activo y soy redirigido a `/proyectos`.
- [ ] Given estoy en `/login`, When ingreso un nombre que identifica a un único usuario, Then se guarda el usuario activo y soy redirigido a `/proyectos`.
- [ ] Given estoy en `/login`, When ingreso un email que no existe, Then veo la tarjeta "Usuario no encontrado" para crear el usuario.
- [ ] Given dejo el campo vacío, When intento ingresar, Then veo `Ingresá tu email` y no se envía la búsqueda.
- [ ] Given hay más de un usuario con el mismo nombre, When intento ingresar por nombre, Then se me pide el email.
- [ ] Given completo el registro de un usuario nuevo, When la API responde 201, Then se inicia sesión y voy a `/proyectos`.
- [ ] Given intento registrar un email ya existente, When la API responde 422, Then veo el mensaje de email duplicado.
- [ ] Given ingresé correctamente, When recargo la página, Then sigo con la sesión iniciada.
- [ ] Given estoy en el panel, When presiono "Salir", Then se limpia la sesión y vuelvo a `/login`.

## Fuera de alcance

- Autenticación real (contraseña, token, OAuth).
- Listado y detalle de proyectos (US-F04 / US-F06).
- Menú y layout completo del proyecto (US-F03).
- Editor de apariencia y gráficos del design system.

## Escenarios BDD

Ver `docs/features/US-F02-ingreso-panel.feature`.
