# Contexto y Reglas del Proyecto: Software Metrics & Estimation Engine

## 1. Información General
- **Objetivo:** Aplicación de gestión de proyectos (ALM) y cálculo de métricas de calidad de software.
- **Lenguaje:** Go (Golang).

## 2. Metodologías Obligatorias (¡CRÍTICO PARA EL AGENTE!)
1. **TDD (Test-Driven Development):** 
   - SIEMPRE escribir o proponer primero las pruebas unitarias en Go (`*_test.go`) siguiendo el ciclo **RED -> GREEN -> REFACTOR** para las reglas de negocio y métricas.
2. **BDD (Behavior-Driven Development):**
   - Las funcionalidades deben contemplar escenarios en formato `Given - When - Then` (Casos normales, alternativos, límite y errores).
3. **SDD (Specification-Driven Development):**
   - Definir claramente Objetivo, Entradas, Salidas esperadas, Reglas de negocio y Criterios de aceptación antes de implementar.

## 3. Arquitectura del Proyecto (Clean Architecture)
- `cmd/api/`: Punto de entrada (`main.go`).
- `internal/domain/`: Modelos y estructuras de datos puras (`structs`) sin dependencias externas.
- `internal/repository/`: Capa de persistencia (PostgreSQL con GORM).
- `internal/service/`: Lógica de negocio, reglas de Planning Poker y motor de métricas (Velocidad, Desviaciones, Defectos).
- `internal/handler/`: Controladores HTTP / Endpoints API.

Stack de infraestructura: **Gin** (HTTP/REST), **PostgreSQL 16** (driver GORM/Postgres), configuración por `.env` + godotenv.

## 4. Requerimientos del Dominio
- **Proyectos e Integrantes:** Gestión y fechas.
- **Product Backlog & Sprints:** Historias de usuario, Story Points, Sprint Goals, cierre de Sprints.
- **Planning Poker:** Votos individuales ocultos, revelado, detección de dispersión y rondas.
- **Worklogs & Defectos:** Registro de horas reales vs estimadas y seguimiento de Bugs por Sprint.
- **Métricas (Core en Go):** Velocidad del equipo, Desviación de esfuerzo, porcentaje de completitud, ratio de defectos.

## 5. Reglas de Generación de Código para la IA
- Escribir código idiomático en Go con manejo explícito de errores (`if err != nil`).
- Mantener evidencia de pruebas unitarias automatizadas para respaldar TDD.
- Priorizar código limpio, modular y fácil de testear mediante interfaces.