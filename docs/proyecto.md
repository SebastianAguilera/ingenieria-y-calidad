# Trabajo Práctico Integrador

## Objetivo
Los alumnos deberán desarrollar una aplicación que permita realizar la estimación, seguimiento y medición de proyectos de software, aplicando los conceptos y prácticas adquiridos durante la asignatura[cite: 1].

El proyecto deberá desarrollarse obligatoriamente utilizando[cite: 1]:
* **Go (Golang)** como lenguaje de programación[cite: 1].
* **Scrum** para la organización y gestión del proyecto[cite: 1].
* **SDD (Specification-Driven Development)** para la especificación de funcionalidades[cite: 1].
* **BDD (Behavior-Driven Development)** para la definición y validación de comportamientos[cite: 1].
* **TDD (Test-Driven Development)** para el desarrollo y prueba de componentes[cite: 1].
* **Git** para control de versiones[cite: 1].
* **Herramientas de Inteligencia Artificial** como soporte al proceso de desarrollo[cite: 1].

Los equipos estarán integrados por un máximo de 5 alumnos[cite: 1].

---

## Proyecto: Software Metrics & Estimation
Se deberá desarrollar una aplicación que permita administrar un proyecto de software y obtener información relacionada con su estimación, planificación, seguimiento y calidad[cite: 1].

La aplicación podrá ser web, de escritorio o consola, pero el núcleo de la solución y sus reglas de negocio deberán estar desarrollados en Go[cite: 1].

---

## Requerimientos Mínimos

### 1. Gestión de proyectos
La aplicación deberá permitir[cite: 1]:
* Crear y modificar proyectos[cite: 1].
* Registrar integrantes[cite: 1].
* Registrar fecha de inicio y finalización[cite: 1].
* Consultar el estado de un proyecto[cite: 1].

### 2. Product Backlog
Para cada proyecto se deberá administrar un Product Backlog[cite: 1]. Cada elemento deberá contener como mínimo[cite: 1]:
* Identificador[cite: 1].
* Título[cite: 1].
* Descripción[cite: 1].
* Prioridad[cite: 1].
* Estado[cite: 1].
* Story Points[cite: 1].
* Criterios de aceptación[cite: 1].

### 3. Gestión de Sprints
La aplicación deberá permitir[cite: 1]:
* Crear Sprints[cite: 1].
* Definir un Sprint Goal[cite: 1].
* Asignar historias a un Sprint[cite: 1].
* Registrar historias completadas[cite: 1].
* Cerrar un Sprint[cite: 1].
* Consultar Sprints anteriores[cite: 1].

### 4. Estimación
La aplicación deberá permitir estimar historias mediante Story Points[cite: 1]. Además, deberá implementar un mecanismo de Planning Poker que permita[cite: 1]:
* Registrar las estimaciones individuales[cite: 1].
* Mantenerlas ocultas hasta finalizar la votación[cite: 1].
* Mostrar las estimaciones realizadas[cite: 1].
* Detectar diferencias entre estimaciones[cite: 1].
* Realizar nuevas rondas[cite: 1].
* Registrar la estimación acordada[cite: 1].

### 5. Registro de esfuerzo
Para cada historia o tarea se deberá poder registrar[cite: 1]:
* Integrante[cite: 1].
* Fecha[cite: 1].
* Actividad realizada[cite: 1].
* Horas trabajadas[cite: 1].

Esto deberá permitir comparar posteriormente el esfuerzo estimado con el esfuerzo real[cite: 1].

### 6. Gestión de defectos
El sistema deberá permitir registrar defectos indicando como mínimo[cite: 1]:
* Descripción[cite: 1].
* Severidad[cite: 1].
* Estado[cite: 1].
* Historia relacionada[cite: 1].
* Sprint de detección[cite: 1].
* Sprint de resolución[cite: 1].

### 7. Métricas
La aplicación deberá calcular como mínimo[cite: 1]:
* Story Points planificados[cite: 1].
* Story Points completados[cite: 1].
* Velocidad del equipo[cite: 1].
* Horas estimadas[cite: 1].
* Horas reales[cite: 1].
* Desviación entre esfuerzo estimado y real[cite: 1].
* Porcentaje de historias completadas[cite: 1].
* Cantidad de defectos detectados[cite: 1].
* Cantidad de defectos resueltos[cite: 1].

### 8. Dashboard
El sistema deberá presentar un Dashboard con información sobre el estado del proyecto y representaciones gráficas de algunas de las métricas obtenidas[cite: 1].

### 9. Reportes
La aplicación deberá generar un reporte de un proyecto o Sprint incluyendo[cite: 1]:
* Historias planificadas y completadas[cite: 1].
* Estimaciones[cite: 1].
* Esfuerzo registrado[cite: 1].
* Métricas[cite: 1].
* Defectos[cite: 1].

---

## Metodología de Desarrollo

El proyecto deberá gestionarse mediante Scrum, utilizando los siguientes roles[cite: 1]:
* **Product Architect:** profesores[cite: 1].
* **Agile Enabler:** un integrante del equipo[cite: 1].
* **Product Builders:** integrantes del equipo responsables de construir el producto[cite: 1].

El proyecto deberá organizarse mediante un Product Backlog y Sprints, realizando las correspondientes actividades de planificación, seguimiento, revisión y retrospectiva[cite: 1].

### Aplicación de SDD
Las funcionalidades principales deberán ser especificadas antes de su implementación[cite: 1]. Las especificaciones deberán establecer como mínimo[cite: 1]:
* Objetivo[cite: 1].
* Entradas[cite: 1].
* Salidas esperadas[cite: 1].
* Reglas de negocio[cite: 1].
* Restricciones[cite: 1].
* Casos límite[cite: 1].
* Condiciones de error[cite: 1].
* Criterios de aceptación[cite: 1].

Las especificaciones deberán mantenerse versionadas junto con el proyecto[cite: 1].

### Aplicación de BDD
Las funcionalidades seleccionadas deberán contar con escenarios que describan su comportamiento mediante[cite: 1]:
> **Given - When - Then**[cite: 1]

Los escenarios deberán contemplar[cite: 1]:
* Casos normales[cite: 1].
* Casos alternativos[cite: 1].
* Casos límite[cite: 1].
* Errores[cite: 1].

Siempre que sea posible deberán automatizarse[cite: 1].

### Aplicación de TDD
Las principales reglas de negocio y cálculos deberán desarrollarse utilizando el ciclo[cite: 1]:
> **RED → GREEN → REFACTOR**[cite: 1]

Se deberán implementar pruebas unitarias en Go para, como mínimo[cite: 1]:
* Cálculo de métricas[cite: 1].
* Cálculos de estimación[cite: 1].
* Reglas de negocio[cite: 1].
* Validaciones[cite: 1].

El equipo deberá mantener evidencia del proceso de TDD mediante el historial del repositorio[cite: 1].

### Uso de Inteligencia Artificial
Se aceptará el uso de herramientas de Inteligencia Artificial durante el proyecto para tareas como[cite: 1]:
* Análisis de requisitos[cite: 1].
* Elaboración y revisión de especificaciones[cite: 1].
* Generación de código[cite: 1].
* Generación de pruebas[cite: 1].
* Refactorización[cite: 1].
* Revisión de código[cite: 1].
* Documentación[cite: 1].

Todo resultado generado mediante IA deberá ser comprendido, revisado y validado por el equipo[cite: 1]. Los alumnos serán responsables por todo el código incorporado al proyecto, independientemente de que haya sido desarrollado manualmente o generado con asistencia de IA[cite: 1].

---

## Plan de Trabajo

Se propone desarrollar el proyecto mediante[cite: 1]:

* **Sprint 0: La Preparación**[cite: 1]
  * **Objetivo:** Formar equipos, asignar roles, configurar el entorno de trabajo[cite: 1].
  * **Tareas:**
    * Crear el repositorio en GitHub/GitLab[cite: 1].
    * Configurar el tablero del proyecto en GitHub Projects[cite: 1].
    * Reunión inicial: El profesor presenta la visión del producto[cite: 1].
    * El equipo crea el Product Backlog inicial, escribiendo las primeras historias de usuario junto al Cliente[cite: 1].

* **Sprint 1: El MVP**[cite: 1]
  * **Objetivo:** Entregar una versión funcional básica[cite: 1].
  * **Ceremonias:** Se realizan todas: Planning, Daily, Review y Retrospective[cite: 1].

* **Sprint 2: La Interfaz**[cite: 1]
  * **Objetivo:** Dotar al software de una interfaz usable[cite: 1].

* **Sprint 3: Funcionalidad y Calidad**[cite: 1]
  * **Objetivo:** Añadir funcionalidades clave y robustecer el sistema (implementar la visualización)[cite: 1].

* **Sprint 4: El Cierre y la Entrega Final**[cite: 1]
  * **Objetivo:** Pulir el producto, documentar y preparar la entrega final (el sistema permitirá exportar un informe del proyecto a un archivo en formato PDF)[cite: 1].
  * **Sprint Review Final:** Presentación formal del producto completo al Cliente[cite: 1].

---

## Trazabilidad

El equipo deberá demostrar trazabilidad entre[cite: 1]:
> **Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go**[cite: 1]

Durante la presentación final se seleccionará al menos una funcionalidad para demostrar este recorrido completo[cite: 1].

---

## Entregables

Cada equipo deberá entregar[cite: 1]:
1. Repositorio Git con historial de contribuciones[cite: 1].
2. Tablero Scrum[cite: 1].
3. Product Backlog y Sprint Backlogs[cite: 1].
4. Especificaciones SDD[cite: 1].
5. Escenarios BDD[cite: 1].
6. Pruebas automatizadas[cite: 1].
7. Código fuente en Go[cite: 1].
8. Evidencias de aplicación de TDD[cite: 1].
9. Software funcional[cite: 1].
10. Informe de métricas y cobertura de pruebas[cite: 1].
11. Actas de retrospectivas[cite: 1].
12. Documentación técnica y manual breve de usuario[cite: 1].
13. Presentación y demostración final[cite: 1].

---

## Criterios de Evaluación

| Criterio | Ponderación | Detalles |
| :--- | :---: | :--- |
| **Producto funcional** | **25%** | Cumplimiento de los requerimientos, correctitud, usabilidad y robustez[cite: 1]. |
| **SDD, BDD y TDD** | **25%** | Calidad de las especificaciones, calidad de escenarios BDD, aplicación de TDD, pruebas automatizadas y trazabilidad[cite: 1]. |
| **Calidad del software** | **20%** | Arquitectura, calidad del código Go, modularidad, mantenibilidad, pruebas, cobertura y manejo adecuado de errores[cite: 1]. |
| **Gestión del proyecto** | **20%** | Calidad del Product Backlog, planificación y cumplimiento de Sprints, gestión del tablero, reviews, retrospectivas y uso adecuado del repositorio Git[cite: 1]. |
| **Trabajo en equipo y presentación** | **10%** | Participación de los integrantes, colaboración, capacidad para justificar decisiones tomadas, calidad de la presentación y demostración final[cite: 1]. |