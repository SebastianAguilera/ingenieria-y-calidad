import { api } from './client'
import type { AltaUsuario, Usuario } from './types'

export type ResultadoBusquedaUsuario =
  | { estado: 'encontrado'; usuario: Usuario }
  | { estado: 'no_encontrado' }
  | { estado: 'ambiguo'; usuarios: Usuario[] }

export function normalizarIdentificador(valor: string): string {
  return valor.trim().toLowerCase()
}

export function buscarUsuarioEnLista(
  usuarios: Usuario[],
  identificador: string,
): ResultadoBusquedaUsuario {
  const normalizado = normalizarIdentificador(identificador)
  if (normalizado === '') {
    return { estado: 'no_encontrado' }
  }

  const porEmail = usuarios.filter(
    (usuario) => normalizarIdentificador(usuario.email) === normalizado,
  )
  if (porEmail.length === 1) {
    return { estado: 'encontrado', usuario: porEmail[0] }
  }

  const porNombre = usuarios.filter(
    (usuario) => normalizarIdentificador(usuario.nombre) === normalizado,
  )
  if (porNombre.length === 1) {
    return { estado: 'encontrado', usuario: porNombre[0] }
  }
  if (porNombre.length > 1) {
    return { estado: 'ambiguo', usuarios: porNombre }
  }

  return { estado: 'no_encontrado' }
}

export function listarUsuarios(): Promise<Usuario[]> {
  return api.get<Usuario[]>('/usuarios')
}

export async function buscarUsuario(identificador: string): Promise<ResultadoBusquedaUsuario> {
  if (normalizarIdentificador(identificador) === '') {
    return { estado: 'no_encontrado' }
  }

  const usuarios = await listarUsuarios()
  return buscarUsuarioEnLista(usuarios, identificador)
}

export function crearUsuario(alta: AltaUsuario): Promise<Usuario> {
  return api.post<Usuario>('/usuarios', alta)
}
