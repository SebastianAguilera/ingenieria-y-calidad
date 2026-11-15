import { useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { crearProyecto } from '../api/proyectos'
import { esApiError } from '../api/errors'
import PageHeader from '../components/PageHeader'
import Alert from '../ui/Alert'
import Button from '../ui/Button'
import Card from '../ui/Card'
import TextField from '../ui/TextField'
import { validarProyecto } from './validacionProyecto'
import type { CamposFormularioProyecto, ErroresFormularioProyecto } from './validacionProyecto'

const CAMPOS_FORMULARIO: ReadonlyArray<keyof CamposFormularioProyecto> = [
  'nombre',
  'fecha_inicio',
  'fecha_fin',
]

function esCampoFormulario(campo: string | undefined): campo is keyof CamposFormularioProyecto {
  return CAMPOS_FORMULARIO.some((valido) => valido === campo)
}

export default function NuevoProyectoPage() {
  const navegar = useNavigate()
  const [campos, setCampos] = useState<CamposFormularioProyecto>({
    nombre: '',
    fecha_inicio: '',
    fecha_fin: '',
  })
  const [errores, setErrores] = useState<ErroresFormularioProyecto>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)
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

  async function enviar() {
    if (envioEnCurso.current) {
      return
    }
    envioEnCurso.current = true
    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      const erroresCliente = validarProyecto(campos)
      if (Object.keys(erroresCliente).length > 0) {
        setErrores(erroresCliente)
        return
      }

      try {
        // El backend vincula al creador como LIDER en la misma transacción del
        // alta (US-F10 / US-11): no hace falta asociarlo después.
        const proyecto = await crearProyecto({
          nombre: campos.nombre.trim(),
          fecha_inicio: campos.fecha_inicio,
          fecha_fin: campos.fecha_fin === '' ? null : campos.fecha_fin,
        })
        navegar(`/proyectos/${proyecto.id}`)
      } catch (error) {
        mostrarError(error)
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
            autoComplete="off"
          />
          <TextField
            label="Fecha de inicio"
            type="date"
            value={campos.fecha_inicio}
            onChange={(evento) => cambiar('fecha_inicio', evento.target.value)}
            error={errores.fecha_inicio}
          />
          <TextField
            label="Fecha de fin"
            type="date"
            value={campos.fecha_fin}
            onChange={(evento) => cambiar('fecha_fin', evento.target.value)}
            error={errores.fecha_fin}
          />
          <div className="flex flex-wrap gap-3">
            <Button type="submit" disabled={enviando}>
              {enviando ? 'Guardando…' : 'Guardar'}
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
