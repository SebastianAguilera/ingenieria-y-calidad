import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import PageHeader from './PageHeader'

describe('PageHeader', () => {
  it('muestra el título y la descripción recibidos', () => {
    render(<PageHeader title="Proyectos" description="Lista de proyectos" />)

    expect(screen.getByRole('heading', { level: 1, name: 'Proyectos' })).toBeInTheDocument()
    expect(screen.getByText('Lista de proyectos')).toBeInTheDocument()
  })
})
