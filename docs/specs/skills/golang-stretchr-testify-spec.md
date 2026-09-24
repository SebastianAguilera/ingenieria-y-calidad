# Especificación: SK-027 - golang-stretchr-testify: Uso de stretchr/testify en tests Go

## Objetivo
Guía de uso de `stretchr/testify` en profundidad para testing de Go: paquetes `assert`, `require`, `mock` y `suite`, elección entre assert y require, expectativas de mocks, argument matchers, verificación de llamadas, lifecycle de suites y patrones avanzados (Eventually, JSONEq, custom matchers). Aplica cuando el codebase importa `github.com/stretchr/testify`. Se usan tests como especificaciones ejecutables: restringen comportamiento y hacen los fallos auto-explicativos, no apuntan a cobertura.

## Entradas
- Código Go bajo prueba y tests existentes (`**/*.go`) que usan testify.
- Modo: **Write** (agregar tests/mocks) o **Review** (auditar tests existentes por mal uso de testify).
- Binarios: `go`, `gotests` (`go install github.com/cweill/gotests/...@latest`).

## Salidas esperadas
1. Tests con `assert` (verificaciones, continúan) y `require` (precondiciones, detienen).
2. Mocks de interfaces con expectativas tipadas y verificación de llamadas.
3. Suites con lifecycle (`SetupSuite`, `SetupTest`, `TearDownTest`, `TearDownSuite`) cuando agrupan tests con estado compartido.
4. Assertions avanzadas: `Eventually`, `JSONEq`, custom matchers.
5. En Review mode: reporte de malos usos (assert/require mezclados al azar, mocks sobre tipos concretos).

## Reglas de negocio
1. `require` para precondiciones (setup, chequeos de error — si continúa, panic o resultados engañosos); `assert` para verificaciones. NUNCA mezclarlos al azar.
2. Usar `assert.New(t)`/`require.New(t)` para legibilidad; nombres convencionales `is` y `must`.
3. testify complementa, no reemplaza `testing` — siempre `*testing.T` como entry point.
4. Mockear interfaces, no tipos concretos.
5. `assert` registra el fallo y continúa (se ven todos los fallos de una vez); `require` llama `t.FailNow()`.
6. Preferir `assert.ElementsMatch`/`assert.JSONEq` sobre comparaciones frágiles de slices/JSON serializados.
7. Para asserts: `Equal` con DeepEqual + tipo exacto; `EqualValues` convierte a tipo común antes.
8. El skill no es exhaustivo: ante dudas de API, remitirse a la documentación oficial de testify; verificar símbolos/vulnerabilidades con `godig` (skill `golang-pkg-go-dev`) y navegar el uso local con `gopls`.

## Restricciones
- Metodología general de tests (tabla-driven, goleak, integración) → skill `golang-testing`.
- Medición de rendimiento → skill `golang-benchmark`.

## Casos límite
- Test que parsea un config y necesita detenerse si falla: `must.NoError(err)` antes de usar `cfg`.
- Verificación de una llamada a mock con cualquier argumento: `mock.Anything` en la expectativa.
- Comparar mapas/slices sin importar el orden: `ElementsMatch`, no `Equal`.
- Test que espera un evento asincrónico: `Eventually` con timeout en lugar de sleep.

## Condiciones de error
- `require` usado en una verificación no crítica (frena el test y oculta el resto de fallos): mover a `assert`.
- `assert` usado en una precondición (libera nil dereferences aguas abajo): cambiar a `require`.
- Mock con expectativa que no se cumple: falla en `AssertExpectations`/verificación de llamadas (con mensaje claro).

## Criterios de aceptación
- [ ] La elección assert/require sigue la regla precondiciones vs verificaciones.
- [ ] Los mocks se definen sobre interfaces y verifican llamadas.
- [ ] Las suites usan los hooks de lifecycle correctos.
- [ ] Las comparaciones usan matchers idiomáticos (ElementsMatch, JSONEq, Eventually).
- [ ] En Review mode, los malos usos se reportan con la corrección sugerida.

## Escenarios BDD

### Caso normal
Given un test que verifica el parseo de un config válido
When el agente lo escribe con testify
Then usa `must := require.New(t)` para NoError/NotNil (precondiciones)
And `is := assert.New(t)` para Equal de los campos (verificaciones)

### Caso alternativo
Given un service que envía emails y necesita testearse
When el agente genera el test
Then mockea la interface `Mailer` con expectativas tipadas
And verifica la llamada con `AssertCalled` al final del test

### Caso límite
Given un handler que responde un JSON con campos en distinto orden
When el agente compara la respuesta
Then usa `is.JSONEq(expectedJSON, actualBody)`
And evita la comparación frágil de strings

### Caso de error
Given un test con una precondición que falla usando `assert` (sin detener)
When el test continúa y dereferencia un valor nil
Then el test panic en lugar de reportar la precondición fallida
And el agente corrige cambiando esa assertion a `require`