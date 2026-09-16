# Especificación: SK-008 - golang-database: Acceso a bases de datos en Go

## Objetivo
Guía de acceso a bases de datos en Go: queries parametrizadas, escaneo de structs, columnas NULLable, transacciones, niveles de aislamiento, `SELECT ... FOR UPDATE`, connection pool, procesamiento por lotes, propagación de context y herramientas de migración. Aplica a PostgreSQL, MariaDB, MySQL y SQLite con `database/sql`, `sqlx` o `pgx`. El skill NO genera esquemas de BD ni SQL de migración.

## Entradas
- Código Go que interactúa con bases de datos (funciones de repository, helpers de query, wrappers de transacción).
- Modo: Write (generar código siguiendo patrones existentes del repo) o Review/debug (auditar: `rows.Close()` faltantes, queries sin parametrizar, context ausente, errores no chequeados).
- Binarios: `go`.

## Salidas esperadas
1. Queries parametrizadas (nunca concatenación de input de usuario en SQL).
2. Uso de `sqlx` o `pgx` sobre `database/sql` — nunca ORMs.
3. `*Context` variants (`QueryContext`, `ExecContext`, `GetContext`) en toda operación.
4. Manejo explícito de `sql.ErrNoRows` con `errors.Is`.
5. `defer rows.Close()` inmediatamente después de `QueryContext`.
6. Uso de `db.Exec` para statements que no retornan filas.
7. Transacciones con `BeginTxx`/`Commit` para operaciones multi-statement.
8. `SELECT ... FOR UPDATE` cuando se lee para modificar; isolation levels custom cuando READ COMMITTED no alcanza.
9. Columnas NULLable con punteros (`*string`, `*int`) o `sql.NullXxx`.
10. Connection pool configurado; operaciones batch en tamaños razonables.

## Reglas de negocio
1. Usar `sqlx` o `pgx`, NUNCA ORMs (ocultan SQL, generan queries impredecibles, complican el debugging).
2. Queries DEBEN usar placeholders parametrizados — NUNCA concatenar input de usuario en strings SQL (SQL injection).
3. Context DEBE pasarse a todas las operaciones (`*Context` variants).
4. `sql.ErrNoRows` DEBE manejarse explícitamente; distinguir "no encontrado" de errores reales con `errors.Is`.
5. Rows DEBEN cerrarse tras iterar: `defer rows.Close()` justo después del `QueryContext`.
6. NUNCA usar `db.Query` para statements sin filas de retorno (conexión filtrada al pool si se olvida el close); usar `db.Exec`.
7. Operaciones multi-statement en transacción (`BeginTxx`/`Commit`).
8. `SELECT ... FOR UPDATE` al leer datos que se modificarán (previene race conditions).
9. Aislamiento custom solo cuando READ COMMITTED es insuficiente (ej. serializable para operaciones financieras).
10. Connection pool DEBE configurarse (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, `SetConnMaxIdleTime`).
11. Migraciones con herramientas externas (golang-migrate o Flyway), nunca SQL de migración manual o generado por IA.
12. Batch en tamaños razonables — ni fila por fila (demasiados round trips) ni millones de una vez (locks y memoria).
13. NUNCA crear o modificar esquemas de BD; NUNCA depender de triggers, views, stored procedures o row-level security en código de aplicación.

## Restricciones
- No genera esquemas ni SQL de migración.
- Para APIs específicas de un driver, remitirse a la documentación oficial del mismo.

## Casos límite
- Columnas NULLable leídas en structs sin pointer/NullXxx: escaneo falla — usar `*string`/`sql.NullString`.
- Statement que no retorna filas ejecutado con `db.Query`: conexión filtrada si no se cierra — usar `db.Exec`.
- Lectura-para-modificar sin `FOR UPDATE` bajo concurrencia: actualización perdida — aplicar `SELECT ... FOR UPDATE`.

## Condiciones de error
- `rows.Err()` no chequeado tras el loop de iteración: errores de I/O silenciados — verificar siempre.
- Query con concatenación de input: SQL injection — convertir a placeholders (`$1`, `?`).
- Transacción sin Rollback en paths de error: locks retenidos — usar `defer tx.Rollback()` y commit explícito.

## Criterios de aceptación
- [ ] Todas las queries usan placeholders parametrizados.
- [ ] El código usa `sqlx`/`pgx` sobre `database/sql`; no hay ORMs.
- [ ] Todas las operaciones usan variantes con context.
- [ ] `sql.ErrNoRows` se maneja con `errors.Is`.
- [ ] Las rows se cierran con `defer` inmediato y se chequea `rows.Err()`.
- [ ] Los statements sin filas usan `db.Exec`.
- [ ] Las operaciones multi-statement usan transacciones.
- [ ] El connection pool está configurado.

## Escenarios BDD

### Caso normal
Given un repository que busca un usuario por email
When el agente escribe la query
Then usa `QueryContext` con placeholder `$1` (o `?`)
And escanea el resultado a un struct

### Caso alternativo
Given una operación que inserta un usuario y su dirección en dos tablas
When el agente implementa la persistencia
Then envuelve ambas escrituras en una transacción `BeginTxx`/`Commit`
And hace rollback si la segunda escritura falla

### Caso límite
Given una consulta que puede no devolver filas
When el agente maneja el resultado
Then compara el error con `errors.Is(err, sql.ErrNoRows)`
And distingue "no encontrado" de un error real de BD

### Caso de error
Given una función que construye SQL concatenando el email del usuario
When el agente revisa el código
Then detecta la vulnerabilidad de SQL injection
And reescribe la query con placeholder parametrizado