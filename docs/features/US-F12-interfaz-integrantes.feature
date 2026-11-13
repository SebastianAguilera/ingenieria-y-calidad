# language: es
Característica: Interfaz de integrantes y roles del proyecto
  Como integrante de un proyecto
  Quiero consultar el equipo y, si soy líder, gestionarlo
  Para mantener actualizados los roles y la pertenencia al proyecto

  Antecedentes:
    Dado que existe el proyecto "Métricas" con id 1
    Y existe una sesión activa

  Escenario: E-F12-01 un integrante ve el equipo con sus roles
    Cuando entra a la pantalla Equipo del proyecto
    Entonces ve a cada integrante con su nombre, email y rol
    Y los roles mostrados son "LIDER" o "MIEMBRO"

  Escenario: E-F12-02 un líder ve las acciones de gestión
    Dado que el usuario activo es LIDER del proyecto
    Cuando entra a la pantalla Equipo
    Entonces ve el botón "Agregar integrante"
    Y ve un selector de rol por cada fila
    Y ve la acción "Quitar" en las demás filas

  Escenario: E-F12-03 un miembro no ve acciones de gestión sobre otros
    Dado que el usuario activo es MIEMBRO del proyecto
    Cuando entra a la pantalla Equipo
    Entonces no ve el botón "Agregar integrante"
    Y no ve selectores de rol ajenos
    Y no ve la acción "Quitar" en filas ajenas

  Escenario: E-F12-04 cualquier integrante puede salir del proyecto
    Dado que el usuario activo es integrante del proyecto
    Cuando confirma "Salir del proyecto" en su fila
    Entonces se envía la baja de su propio usuario
    Y la aplicación lo lleva al listado de proyectos

  Escenario: E-F12-05 el líder agrega un integrante por email
    Dado que el email "cris@utn.edu.ar" corresponde a un usuario registrado
    Y no es integrante del proyecto
    Cuando el líder lo agrega con rol "MIEMBRO"
    Entonces el modal se cierra
    Y la persona aparece en la tabla como "MIEMBRO"

  Escenario: E-F12-06 el email no está registrado
    Cuando el líder intenta agregar "dario@utn.edu.ar"
    Entonces ve junto al campo email "No existe un usuario registrado con ese email"
    Y el modal permanece abierto

  Escenario: E-F12-07 la persona ya es integrante
    Cuando el líder intenta agregar un email que ya pertenece al proyecto
    Entonces ve junto al campo email "Ese usuario ya es integrante"

  Escenario: E-F12-08 email con formato inválido
    Cuando el líder envía un email vacío o inválido
    Entonces ve el error en el campo email
    Y no se envía la petición

  Escenario: E-F12-09 el líder cambia el rol de un integrante
    Dado que "Ben" es MIEMBRO
    Cuando el líder lo cambia a "LIDER"
    Entonces se envía el cambio de rol
    Y la fila de Ben muestra "LIDER"

  Escenario: E-F12-10 cambiar al mismo rol no genera una petición
    Dado que "Ana" ya es LIDER
    Cuando el líder selecciona "LIDER" para Ana
    Entonces no se envía ninguna petición

  Escenario: E-F12-11 degradar al último líder se rechaza
    Dado que "Ana" es la única LIDER
    Cuando intenta cambiar su rol a "MIEMBRO"
    Entonces ve "el proyecto debe conservar al menos un líder"
    Y Ana sigue siendo LIDER

  Escenario: E-F12-12 el líder quita a otro integrante
    Dado que "Ben" es integrante
    Cuando el líder confirma "Quitar" sobre Ben
    Entonces se envía la baja de Ben
    Y Ben desaparece de la tabla

  Escenario: E-F12-13 proyecto cerrado es de solo lectura
    Dado que el proyecto está "Cerrado"
    Cuando un líder entra a la pantalla Equipo
    Entonces ve un aviso para reabrir el proyecto
    Y no ve acciones de alta, cambio de rol ni baja

  Escenario: E-F12-14 el proyecto no existe o no hay acceso
    Dado que el usuario no es miembro del proyecto
    Cuando entra a la pantalla Equipo
    Entonces ve "No tenés acceso a este proyecto o no existe"

  Escenario: E-F12-15 error de red durante el alta
    Cuando falla la conexión al agregar un integrante
    Entonces ve "No se pudo conectar con el servidor. Intentá nuevamente."
    Y el email ingresado se conserva en el formulario
