import { api } from './client'
import type { Usuario } from './types'

export type SesionLogin = {
  token: string
  usuario: Usuario
}

export function login(email: string, password: string): Promise<SesionLogin> {
  return api.post<SesionLogin>('/auth/login', { email, password })
}
