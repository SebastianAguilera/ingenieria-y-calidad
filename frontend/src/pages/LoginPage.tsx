import { useState } from 'react'
import type { FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { esApiError } from '../api/errors'
import { buscarUsuario, crearUsuario } from '../api/usuarios'
import { useSesion } from '../session/sesion'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import ModeSwitch from '../ui/ModeSwitch'
import TextField from '../ui/TextField'

type Vista = 'login' | 'registro'

type DatosRegistro = {
  nombre: string
  email: string
}

const MENSAJE_VACIO = 'Ingresá tu email'
const MENSAJE_AMBIGUO = 'Hay más de un usuario con ese nombre. Ingresá tu email.'
const MENSAJE_DUPLICADO = 'Ese email ya está registrado. Volvé a ingresar.'
const MENSAJE_RED = 'No se pudo conectar con el servidor.'
const MENSAJE_INESPERADO = 'Ocurrió un error inesperado. Intentá de nuevo.'
const MENSAJE_REGISTRO_INCOMPLETO = 'Completá tu nombre y tu email'

function pareceEmail(valor: string): boolean {
  return valor.includes('@')
}

function mensajeDeError(error: unknown): string {
  if (esApiError(error)) {
    if (error.code === 'NETWORK_ERROR') {
      return MENSAJE_RED
    }
    if (error.message.trim() !== '') {
      return error.message
    }
  }
  return MENSAJE_INESPERADO
}

export default function LoginPage() {
  const { usuarioActivo, ingresar } = useSesion()
  const navegar = useNavigate()

  const [vista, setVista] = useState<Vista>('login')
  const [identificador, setIdentificador] = useState('')
  const [registro, setRegistro] = useState<DatosRegistro>({ nombre: '', email: '' })
  const [error, setError] = useState<string | null>(null)
  const [cargando, setCargando] = useState(false)

  if (usuarioActivo) {
    return <Navigate to="/proyectos" replace />
  }

  async function manejarIngreso(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()

    const limpio = identificador.trim()
    if (limpio === '') {
      setError(MENSAJE_VACIO)
      return
    }

    setCargando(true)
    setError(null)

    try {
      const resultado = await buscarUsuario(limpio)
      if (resultado.estado === 'encontrado') {
        ingresar(resultado.usuario)
        navegar('/proyectos', { replace: true })
        return
      }
      if (resultado.estado === 'ambiguo') {
        setError(MENSAJE_AMBIGUO)
        return
      }

      const esEmail = pareceEmail(limpio)
      setRegistro({
        nombre: esEmail ? '' : limpio,
        email: esEmail ? limpio : '',
      })
      setVista('registro')
    } catch (err) {
      setError(mensajeDeError(err))
    } finally {
      setCargando(false)
    }
  }

  async function manejarRegistro(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()

    const nombre = registro.nombre.trim()
    const email = registro.email.trim()
    if (nombre === '' || email === '') {
      setError(MENSAJE_REGISTRO_INCOMPLETO)
      return
    }

    setCargando(true)
    setError(null)

    try {
      const usuario = await crearUsuario({ nombre, email })
      ingresar(usuario)
      navegar('/proyectos', { replace: true })
    } catch (err) {
      const esDuplicado =
        esApiError(err) &&
        (err.code === 'USUARIO_DUPLICADO' || (err.status === 422 && err.campo === 'email'))
      setError(esDuplicado ? MENSAJE_DUPLICADO : mensajeDeError(err))
    } finally {
      setCargando(false)
    }
  }

  function volverAlLogin() {
    setVista('login')
    setError(null)
  }

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
              description="No es autenticación real: escribí el email o el nombre con el que te registraste."
            >
              <form className="flex flex-col gap-4" onSubmit={manejarIngreso} noValidate>
                <TextField
                  label="Email o nombre"
                  name="identificador"
                  autoComplete="username"
                  autoFocus
                  value={identificador}
                  onChange={(evento) => setIdentificador(evento.target.value)}
                  disabled={cargando}
                />
                {error && <Alert>{error}</Alert>}
                <Button type="submit" disabled={cargando}>
                  {cargando ? 'Ingresando...' : 'Ingresar'}
                </Button>
              </form>
            </Card>
          ) : (
            <Card
              title="Usuario no encontrado"
              description="No encontramos ese usuario. Creá tu cuenta para ingresar."
            >
              <form className="flex flex-col gap-4" onSubmit={manejarRegistro} noValidate>
                <TextField
                  label="Nombre"
                  name="nombre"
                  autoComplete="name"
                  value={registro.nombre}
                  onChange={(evento) =>
                    setRegistro((actual) => ({ ...actual, nombre: evento.target.value }))
                  }
                  disabled={cargando}
                />
                <TextField
                  label="Email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  value={registro.email}
                  onChange={(evento) =>
                    setRegistro((actual) => ({ ...actual, email: evento.target.value }))
                  }
                  disabled={cargando}
                />
                {error && <Alert>{error}</Alert>}
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                  <Button type="submit" disabled={cargando}>
                    {cargando ? 'Creando usuario...' : 'Crear usuario'}
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
