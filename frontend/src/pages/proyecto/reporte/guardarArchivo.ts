/** Ofrece el archivo al usuario para guardarlo, con el nombre indicado (US-F23). */
export function guardarArchivo(contenido: Blob, nombre: string): void {
  const url = URL.createObjectURL(contenido)
  const enlace = document.createElement('a')
  enlace.href = url
  enlace.download = nombre
  enlace.rel = 'noopener'
  document.body.append(enlace)
  enlace.click()
  enlace.remove()
  // Se libera después del click: algunos navegadores leen la URL en forma asíncrona.
  setTimeout(() => URL.revokeObjectURL(url), 0)
}
