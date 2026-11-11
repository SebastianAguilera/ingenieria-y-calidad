import { useState } from 'react'
import type { FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { login } from '../api/auth'
import { esApiError } from '../api/errors'
import { useSesion } from '../session/sesion'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import ModeSwitch from '../ui/ModeSwitch'
import TextField from '../ui/TextField'

const MENSAJE_EMAIL_VACIO = 'Ingresá tu email'
const MENSAJE_PASSWORD_VACIO = 'Ingresá tu contraseña'
const MENSAJE_CREDENCIALES = 'Email o contraseña incorrectos'
const MENSAJE_RED = 'No se pudo conectar con el servidor.'
const MENSAJE_INESPERADO = 'Ocurrió un error inesperado. Intentá de nuevo.'

type CampoConError = 'email' | 'password'

function mensajeDeError(error: unknown): string {
  if (esApiError(error)) {
    if (error.code === 'NETWORK_ERROR') {
      return MENSAJE_RED
    }
    if (error.code === 'CREDENCIALES_INVALIDAS') {
      return MENSAJE_CREDENCIALES
    }
    if (error.message.trim() !== '') {
      return error.message
    }
  }
  return MENSAJE_INESPERADO
}

function campoInvalidoDelBackend(error: unknown): CampoConError | null {
  if (esApiError(error) && error.status === 422) {
    if (error.campo === 'email' || error.campo === 'password') {
      return error.campo
    }
  }
  return null
}

export default function LoginPage() {
  const { token, ingresar } = useSesion()
  const navegar = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [campoError, setCampoError] = useState<CampoConError | null>(null)
  const [cargando, setCargando] = useState(false)

  if (token) {
    return <Navigate to="/proyectos" replace />
  }

  async function manejarIngreso(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (cargando) {
      return
    }

    const emailLimpio = email.trim()
    if (emailLimpio === '') {
      setError(MENSAJE_EMAIL_VACIO)
      setCampoError('email')
      return
    }
    if (password === '') {
      setError(MENSAJE_PASSWORD_VACIO)
      setCampoError('password')
      return
    }

    setCargando(true)
    setError(null)
    setCampoError(null)

    try {
      const sesion = await login(emailLimpio, password)
      ingresar(sesion.token, sesion.usuario)
      navegar('/proyectos', { replace: true })
    } catch (err) {
      setError(mensajeDeError(err))
      setCampoError(campoInvalidoDelBackend(err))
    } finally {
      setCargando(false)
    }
  }

  const mensajeDeCampo = (campo: CampoConError): string | undefined =>
    campoError === campo ? (error ?? undefined) : undefined

  return (
    <div className="flex min-h-screen flex-col bg-page text-ink">
      <header className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-5">
        <span className="font-display text-lg font-semibold">Software Metrics</span>
        <ModeSwitch />
      </header>

      <main className="flex flex-1 items-start justify-center px-4 pb-16 pt-6 sm:items-center sm:pt-0">
        <div className="w-full max-w-md">
          <Card
            title="Ingresá al panel"
            description="Ingresá con tu email y contraseña para usar el panel."
          >
            <form className="flex flex-col gap-4" onSubmit={manejarIngreso} noValidate>
              <TextField
                label="Email"
                name="email"
                type="email"
                autoComplete="username"
                autoFocus
                value={email}
                onChange={(evento) => setEmail(evento.target.value)}
                disabled={cargando}
                error={mensajeDeCampo('email')}
              />
              <TextField
                label="Contraseña"
                name="password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(evento) => setPassword(evento.target.value)}
                disabled={cargando}
                error={mensajeDeCampo('password')}
              />
              {!campoError && error && <Alert>{error}</Alert>}
              <Button type="submit" disabled={cargando}>
                {cargando ? 'Ingresando...' : 'Ingresar'}
              </Button>
            </form>
          </Card>
        </div>
      </main>
    </div>
  )
}
