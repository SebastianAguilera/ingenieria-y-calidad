import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../api/client'
import { ApiError } from '../api/errors'
import HomePage from './HomePage'

vi.mock('../api/client', () => ({
  api: {
    get: vi.fn(),
  },
}))

const getMock = vi.mocked(api.get)

describe('HomePage', () => {
  beforeEach(() => {
    getMock.mockReset()
  })

  it('dado un backend con proyectos, muestra el total', async () => {
    getMock.mockResolvedValue({ total: 3, proyectos: [] })

    render(<HomePage />)

    expect(await screen.findByText('Backend conectado: 3 proyecto(s).')).toBeInTheDocument()
    expect(getMock).toHaveBeenCalledWith('/proyectos')
  })

  it('dado un ApiError, muestra el mensaje del error', async () => {
    getMock.mockRejectedValue(
      new ApiError({ status: 500, code: 'ERROR_INTERNO', message: 'fallo interno' }),
    )

    render(<HomePage />)

    expect(await screen.findByText('No se pudo conectar: fallo interno')).toBeInTheDocument()
  })
})
