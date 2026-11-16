import { useState } from 'react'
import type { FormEvent } from 'react'
import type { ActualizacionDefecto, Defecto, EstadoDefecto, Sprint } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import type { ResultadoOperacion } from '../sprints/tipos'
import { CampoSelect } from './campos'
import { ESTADOS_DEFECTO, codigoDefecto, etiquetaEstado } from './etiquetasDefecto'
import MarcoModal from './MarcoModal'

type ModalEstadoDefectoProps = {
  defecto: Defecto
  sprints: Sprint[]
  onGuardar: (actualizacion: ActualizacionDefecto) => Promise<ResultadoOperacion>
  onCerrar: () => void
}

type ErroresEstado = { estado?: string; sprint_resolucion_id?: string }

function exigeSprintDeResolucion(estado: EstadoDefecto): boolean {
  return estado === 'resolved' || estado === 'closed'
}

function esEstado(valor: string): valor is EstadoDefecto {
  return ESTADOS_DEFECTO.some((estado) => estado === valor)
}

export default function ModalEstadoDefecto({
  defecto,
  sprints,
  onGuardar,
  onCerrar,
}: ModalEstadoDefectoProps) {
  const [estado, setEstado] = useState<EstadoDefecto>(defecto.estado)
  const [sprintResolucion, setSprintResolucion] = useState(
    defecto.sprint_resolucion_id === null ? '' : String(defecto.sprint_resolucion_id),
  )
  const [errores, setErrores] = useState<ErroresEstado>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)

  const pideSprint = exigeSprintDeResolucion(estado)

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (enviando) {
      return
    }

    // La regla "resolver o cerrar exige sprint de resolución" la valida el backend (US-09):
    // el cliente solo omite el campo cuando no corresponde enviarlo.
    const actualizacion: ActualizacionDefecto =
      pideSprint && sprintResolucion !== ''
        ? { estado, sprint_resolucion_id: Number(sprintResolucion) }
        : { estado }

    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      const resultado = await onGuardar(actualizacion)

      if (!resultado.ok) {
        if (resultado.campo === 'estado' || resultado.campo === 'sprint_resolucion_id') {
          setErrores({ [resultado.campo]: resultado.mensaje })
        } else {
          setErrorGeneral(resultado.mensaje)
        }
      }
    } finally {
      setEnviando(false)
    }
  }

  return (
    <MarcoModal titulo="Cambiar estado" subtitulo={codigoDefecto(defecto.id)} onCerrar={onCerrar}>
      <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
        {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}

        <CampoSelect
          label="Estado"
          value={estado}
          onChange={(valor) => {
            if (esEstado(valor)) {
              setEstado(valor)
            }
          }}
          opciones={ESTADOS_DEFECTO.map((valor) => ({ valor, texto: etiquetaEstado(valor) }))}
          error={errores.estado}
          autoFocus
        />

        {pideSprint && (
          <CampoSelect
            label="Sprint de resolución"
            value={sprintResolucion}
            onChange={setSprintResolucion}
            placeholder="Seleccioná un sprint"
            opciones={sprints.map((sprint) => ({
              valor: String(sprint.id),
              texto: sprint.identificador,
            }))}
            ayuda="Obligatorio para pasar a Resuelto o Cerrado"
            error={errores.sprint_resolucion_id}
          />
        )}

        <div className="mt-2 flex flex-wrap justify-end gap-3">
          <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
            Cancelar
          </Button>
          <Button type="submit" disabled={enviando}>
            {enviando ? 'Guardando…' : 'Confirmar'}
          </Button>
        </div>
      </form>
    </MarcoModal>
  )
}
