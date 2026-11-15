import { useState } from 'react'
import type { FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { login, registrar } from '../api/auth'
import { esApiError } from '../api/errors'
import { useSesion } from '../session/sesion'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import ModeSwitch from '../ui/ModeSwitch'
import TextField from '../ui/TextField'

type Vista = 'login' | 'registro'

const MENSAJE_NOMBRE_VACIO = 'Ingresá tu nombre'
const MENSAJE_EMAIL_VACIO = 'Ingresá tu email'
const MENSAJE_PASSWORD_VACIO = 'Ingresá tu contraseña'
const MENSAJE_CREDENCIALES = 'Email o contraseña incorrectos'
const MENSAJE_DUPLICADO = 'Ese email ya está registrado. Volvé a ingresar.'
const MENSAJE_RED = 'No se pudo conectar con el servidor.'
const MENSAJE_INESPERADO = 'Ocurrió un error inesperado. Intentá de nuevo.'

type CampoConError = 'nombre' | 'email' | 'password'

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
    if (error.campo === 'nombre' || error.campo === 'email' || error.campo === 'password') {
      return error.campo
    }
  }
  return null
}

export default function LoginPage() {
  const { token, ingresar } = useSesion()
  const navegar = useNavigate()

  const [vista, setVista] = useState<Vista>('login')
  const [nombre, setNombre] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [campoError, setCampoError] = useState<CampoConError | null>(null)
  const [cargando, setCargando] = useState(false)

  if (token) {
    return <Navigate to="/proyectos" replace />
  }

  function abrirRegistro() {
    setError(null)
    setCampoError(null)
    setVista('registro')
  }

  function volverAlLogin() {
    setError(null)
    setCampoError(null)
    setVista('login')
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

  async function manejarRegistro(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (cargando) {
      return
    }

    const nombreLimpio = nombre.trim()
    const emailLimpio = email.trim()
    if (nombreLimpio === '') {
      setError(MENSAJE_NOMBRE_VACIO)
      setCampoError('nombre')
      return
    }
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
      const sesion = await registrar(nombreLimpio, emailLimpio, password)
      ingresar(sesion.token, sesion.usuario)
      navegar('/proyectos', { replace: true })
    } catch (err) {
      if (esApiError(err) && err.code === 'USUARIO_DUPLICADO') {
        setError(MENSAJE_DUPLICADO)
        setCampoError(null)
        return
      }
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
          {vista === 'login' ? (
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
              <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-border pt-4 text-sm">
                <span className="text-ink-soft">¿No tenés cuenta?</span>
                <button
                  type="button"
                  className="font-semibold text-primary underline-offset-4 hover:underline"
                  onClick={abrirRegistro}
                >
                  Crear cuenta
                </button>
              </div>
            </Card>
          ) : (
            <Card
              title="Creá tu cuenta"
              description="Completá tu nombre, email y contraseña para entrar al panel."
            >
              <form className="flex flex-col gap-4" onSubmit={manejarRegistro} noValidate>
                <TextField
                  label="Nombre"
                  name="nombre"
                  autoComplete="name"
                  autoFocus
                  value={nombre}
                  onChange={(evento) => setNombre(evento.target.value)}
                  disabled={cargando}
                  error={mensajeDeCampo('nombre')}
                />
                <TextField
                  label="Email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(evento) => setEmail(evento.target.value)}
                  disabled={cargando}
                  error={mensajeDeCampo('email')}
                />
                <TextField
                  label="Contraseña"
                  name="password"
                  type="password"
                  autoComplete="new-password"
                  value={password}
                  onChange={(evento) => setPassword(evento.target.value)}
                  disabled={cargando}
                  error={mensajeDeCampo('password')}
                />
                {!campoError && error && <Alert>{error}</Alert>}
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                  <Button type="submit" disabled={cargando}>
                    {cargando ? 'Creando cuenta...' : 'Crear cuenta'}
                  </Button>
                  <Button type="button" variante="ghost" onClick={volverAlLogin} disabled={cargando}>
                    Volver
                  </Button>
                </div>
              </form>
            </Card>
          )}
        </div>
      </main>
    </div>
  )
}