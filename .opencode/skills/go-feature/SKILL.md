---
name: go-feature
description: Workflow completo para crear features en Go siguiendo Clean Architecture (domain → service → repository → handler → tests)
metadata:
  architecture: clean-arch
  language: go
---

## Qué hago

Te guío paso a paso para crear una feature completa en el proyecto, respetando la arquitectura por capas.

## Flujo

### 1. Domain (modelos e interfaces)
- Definí structs en `internal/domain/`
- Definí interfaces que el repository va a implementar
- Sin dependencias externas (solo stdlib)

### 2. Repository (persistencia)
- Implementá la interface de domain en `internal/repository/`
- Solo lógica de acceso a datos (PostgreSQL/GORM)
- Sin reglas de negocio

### 3. Service (lógica de negocio)
- Creá el servicio en `internal/service/`
- Depende solo de interfaces de domain
- Manejá errores con `fmt.Errorf("contexto: %w", err)`
- Testeá con mocks de repository

### 4. Handler (HTTP)
- Creá el handler en `internal/handler/`
- Validá entradas antes de pasar a service
- Respondé con códigos HTTP correctos

### 5. Tests
- Escribí tests ANTES de implementar (TDD)
- Table-driven tests para casos múltiples
- Mocks para dependencias externas

## Formato de output

Cuando me pidas crear una feature, te doy:
1. El código para cada capa
2. Los tests correspondientes
3. Los archivos a crear/modificar

## Ejemplo de uso

> "Creá la feature de Planning Poker: modelar Votacion con votos individuales, revelado y detección de dispersión"
