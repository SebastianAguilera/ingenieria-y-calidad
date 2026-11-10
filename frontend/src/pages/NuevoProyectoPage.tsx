import { useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { asociarUsuario, crearProyecto } from '../api/proyectos'
import { esApiError } from '../api/errors'
import type { RolProyecto } from '../api/types'
import PageHeader from '../components/PageHeader'
import { useSesion } from '../session/sesion'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import TextField from '../ui/TextField'
import { validarProyecto } from './validacionProyecto'
import type { CamposFormularioProyecto, ErroresFormularioProyecto } from './validacionProyecto'

const ROL_CREADOR: RolProyecto = 'Facilitador Ágil'
const CAMPOS_FORMULARIO: ReadonlyArray<keyof CamposFormularioProyecto> = [
  'nombre',
  'fecha_inicio',
  'fecha_fin',
]
const CODIGO_USUARIO_YA_ASOCIADO = 'USUARIO_DUPLICADO'

function esCampoFormulario(campo: string | undefined): campo is keyof CamposFormularioProyecto {
  return CAMPOS_FORMULARIO.some((valido) => valido === campo)
}

export default function NuevoProyectoPage() {
  const navegar = useNavigate()
  const { usuarioActivo } = useSesion()
  const [campos, setCampos] = useState<CamposFormularioProyecto>({
    nombre: '',
    fecha_inicio: '',
    fecha_fin: '',
  })
  const [errores, setErrores] = useState<ErroresFormularioProyecto>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)
  const [proyectoCreadoId, setProyectoCreadoId] = useState<number | null>(null)
  const envioEnCurso = useRef(false)

  function cambiar(campo: keyof CamposFormularioProyecto, valor: string) {
    setCampos((anteriores) => ({ ...anteriores, [campo]: valor }))
  }

  function mostrarError(error: unknown) {
    if (esApiError(error)) {
      if (error.status === 0) {
        setErrorGeneral(
          'No se pudo conectar con el servidor. Tus datos se conservaron: intentá nuevamente.',
        )
        return
      }
      if (esCampoFormulario(error.campo)) {
        setErrores({ [error.campo]: error.message })
        return
      }
      setErrorGeneral(error.message)
      return
    }
    setErrorGeneral('Ocurrió un error inesperado. Intentá nuevamente.')
  }

  async function asociarCreador(proyectoId: number): Promise<boolean> {
    if (!usuarioActivo) {
      setErrorGeneral('Necesitás una sesión activa para crear un proyecto.')
      return false
    }
    try {
      await asociarUsuario(proyectoId, usuarioActivo.id, ROL_CREADOR)
      return true
    } catch (error) {
      if (esApiError(error) && error.status === 409 && error.code === CODIGO_USUARIO_YA_ASOCIADO) {
        return true
      }
      setErrorGeneral(
        'El proyecto se creó, pero no se pudo asociar tu usuario. Reintentá la asociación.',
      )
      return false
    }
  }

  async function enviar() {
    if (envioEnCurso.current) {
      return
    }
    envioEnCurso.current = true
    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      let proyectoId = proyectoCreadoId
      if (proyectoId === null) {
        const erroresCliente = validarProyecto(campos)
        if (Object.keys(erroresCliente).length > 0) {
          setErrores(erroresCliente)
          return
        }

        try {
          const proyecto = await crearProyecto({
            nombre: campos.nombre.trim(),
            fecha_inicio: campos.fecha_inicio,
            fecha_fin: campos.fecha_fin === '' ? null : campos.fecha_fin,
          })
          proyectoId = proyecto.id
          setProyectoCreadoId(proyecto.id)
        } catch (error) {
          mostrarError(error)
          return
        }
      }

      if (await asociarCreador(proyectoId)) {
        navegar(`/proyectos/${proyectoId}`)
      }
    } finally {
      envioEnCurso.current = false
      setEnviando(false)
    }
  }

  function alEnviarFormulario(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    void enviar()
  }

  const proyectoYaCreado = proyectoCreadoId !== null

  return (
    <section>
      <PageHeader
        title="Crear proyecto"
        description="Cargá el nombre y las fechas para empezar a planificar."
      />
      <Card>
        <form noValidate onSubmit={alEnviarFormulario} className="flex flex-col gap-5">
          {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}
          <TextField
            label="Nombre"
            value={campos.nombre}
            onChange={(evento) => cambiar('nombre', evento.target.value)}
            error={errores.nombre}
            disabled={proyectoYaCreado}
            autoComplete="off"
          />
          <TextField
            label="Fecha de inicio"
            type="date"
            value={campos.fecha_inicio}
            onChange={(evento) => cambiar('fecha_inicio', evento.target.value)}
            error={errores.fecha_inicio}
            disabled={proyectoYaCreado}
          />
          <TextField
            label="Fecha de fin"
            type="date"
            value={campos.fecha_fin}
            onChange={(evento) => cambiar('fecha_fin', evento.target.value)}
            error={errores.fecha_fin}
            disabled={proyectoYaCreado}
          />
          <div className="flex flex-wrap gap-3">
            <Button type="submit" disabled={enviando}>
              {enviando ? 'Guardando…' : proyectoYaCreado ? 'Reintentar asociación' : 'Guardar'}
            </Button>
            <Button variante="ghost" disabled={enviando} onClick={() => navegar('/proyectos')}>
              Cancelar
            </Button>
          </div>
        </form>
      </Card>
    </section>
  )
}
