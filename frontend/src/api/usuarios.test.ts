import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './client'
import {
  buscarUsuario,
  buscarUsuarioEnLista,
  crearUsuario,
  listarUsuarios,
} from './usuarios'
import type { Usuario } from './types'

vi.mock('./client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)
const postMock = vi.mocked(api.post)

const ana: Usuario = {
  id: 1,
  nombre: 'Ana Pérez',
  email: 'ana@example.com',
  creado_en: '2026-01-01T00:00:00Z',
}

const bruno: Usuario = {
  id: 2,
  nombre: 'Bruno Díaz',
  email: 'bruno@example.com',
  creado_en: '2026-01-02T00:00:00Z',
}

function usuario(nombre: string, email: string, id: number): Usuario {
  return { id, nombre, email, creado_en: '2026-01-01T00:00:00Z' }
}

describe('buscarUsuarioEnLista', () => {
  it('encuentra por email sin distinguir mayúsculas ni espacios', () => {
    const resultado = buscarUsuarioEnLista([ana, bruno], '  ANA@EXAMPLE.COM ')

    expect(resultado).toEqual({ estado: 'encontrado', usuario: ana })
  })

  it('cae a búsqueda por nombre cuando no hay email', () => {
    const resultado = buscarUsuarioEnLista([ana, bruno], 'bruno díaz')

    expect(resultado).toEqual({ estado: 'encontrado', usuario: bruno })
  })

  it('devuelve no_encontrado cuando no hay coincidencia', () => {
    const resultado = buscarUsuarioEnLista([ana, bruno], 'carla@example.com')

    expect(resultado).toEqual({ estado: 'no_encontrado' })
  })

  it('devuelve no_encontrado con identificador vacío', () => {
    const resultado = buscarUsuarioEnLista([ana, bruno], '   ')

    expect(resultado).toEqual({ estado: 'no_encontrado' })
  })

  it('no hace coincidencias parciales por nombre', () => {
    const resultado = buscarUsuarioEnLista([ana, bruno], 'Ana')

    expect(resultado).toEqual({ estado: 'no_encontrado' })
  })

  it('devuelve ambiguo cuando varios usuarios comparten el mismo nombre', () => {
    const otraAna = usuario('Ana Pérez', 'otra@example.com', 3)

    const resultado = buscarUsuarioEnLista([ana, otraAna], 'Ana Pérez')

    expect(resultado).toEqual({ estado: 'ambiguo', usuarios: [ana, otraAna] })
  })
})

describe('listarUsuarios', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('consulta GET /usuarios y devuelve el arreglo', async () => {
    getMock.mockResolvedValue([ana, bruno])

    const usuarios = await listarUsuarios()

    expect(getMock).toHaveBeenCalledWith('/usuarios')
    expect(usuarios).toEqual([ana, bruno])
  })
})

describe('crearUsuario', () => {
  beforeEach(() => {
    postMock.mockReset()
  })

  it('envía POST /usuarios con nombre y email', async () => {
    postMock.mockResolvedValue(ana)

    const creado = await crearUsuario({ nombre: ana.nombre, email: ana.email })

    expect(postMock).toHaveBeenCalledWith('/usuarios', {
      nombre: ana.nombre,
      email: ana.email,
    })
    expect(creado).toEqual(ana)
  })
})

describe('buscarUsuario', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('lista los usuarios y resuelve la coincidencia por email', async () => {
    getMock.mockResolvedValue([ana, bruno])

    const resultado = await buscarUsuario('ana@example.com')

    expect(getMock).toHaveBeenCalledWith('/usuarios')
    expect(resultado).toEqual({ estado: 'encontrado', usuario: ana })
  })

  it('no consulta la API cuando el identificador está vacío', async () => {
    const resultado = await buscarUsuario('   ')

    expect(getMock).not.toHaveBeenCalled()
    expect(resultado).toEqual({ estado: 'no_encontrado' })
  })
})
