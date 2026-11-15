import { useState } from 'react'
import type { FormEvent } from 'react'
import type { ActualizacionDefecto, Defecto, SeveridadDefecto } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import type { ResultadoOperacion } from '../sprints/tipos'
import { CampoAreaTexto, CampoSelect } from './campos'
import { SEVERIDADES_DEFECTO, codigoDefecto, etiquetaSeveridad } from './etiquetasDefecto'
import MarcoModal from './MarcoModal'
import { validarEdicionDefecto } from './validacionDefecto'
import type { CamposEdicionDefecto, ErroresEdicionDefecto } from './validacionDefecto'

type ModalEdicionDefectoProps = {
  defecto: Defecto
  onGuardar: (actualizacion: ActualizacionDefecto) => Promise<ResultadoOperacion>
  onCerrar: () => void
}

function esSeveridad(valor: string): valor is SeveridadDefecto {
  return SEVERIDADES_DEFECTO.some((severidad) => severidad === valor)
}

export default function ModalEdicionDefecto({
  defecto,
  onGuardar,
  onCerrar,
}: ModalEdicionDefectoProps) {
  const [campos, setCampos] = useState<CamposEdicionDefecto>({
    descripcion: defecto.descripcion,
    severidad: defecto.severidad,
  })
  const [errores, setErrores] = useState<ErroresEdicionDefecto>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (enviando) {
      return
    }

    const erroresCliente = validarEdicionDefecto(campos)
    if (Object.keys(erroresCliente).length > 0) {
      setErrores(erroresCliente)
      return
    }

    // Solo se envían los campos que cambiaron: el backend rechaza un PATCH vacío
    // y la edición no debe tocar el estado ni los sprints del defecto.
    const descripcion = campos.descripcion.trim()
    const actualizacion: ActualizacionDefecto = {
      ...(descripcion !== defecto.descripcion ? { descripcion } : {}),
      ...(campos.severidad !== defecto.severidad ? { severidad: campos.severidad } : {}),
    }

    if (Object.keys(actualizacion).length === 0) {
      onCerrar()
      return
    }

    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      const resultado = await onGuardar(actualizacion)

      if (!resultado.ok) {
        if (resultado.campo === 'descripcion' || resultado.campo === 'severidad') {
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
    <MarcoModal titulo="Editar defecto" subtitulo={codigoDefecto(defecto.id)} onCerrar={onCerrar}>
      <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
        {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}

        <CampoAreaTexto
          label="Descripción"
          value={campos.descripcion}
          onChange={(valor) => setCampos((anteriores) => ({ ...anteriores, descripcion: valor }))}
          error={errores.descripcion}
          autoFocus
        />

        <CampoSelect
          label="Severidad"
          value={campos.severidad}
          onChange={(valor) => {
            if (esSeveridad(valor)) {
              setCampos((anteriores) => ({ ...anteriores, severidad: valor }))
            }
          }}
          opciones={SEVERIDADES_DEFECTO.map((severidad) => ({
            valor: severidad,
            texto: etiquetaSeveridad(severidad),
          }))}
          error={errores.severidad}
        />

        <div className="mt-2 flex flex-wrap justify-end gap-3">
          <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
            Cancelar
          </Button>
          <Button type="submit" disabled={enviando}>
            {enviando ? 'Guardando…' : 'Guardar cambios'}
          </Button>
        </div>
      </form>
    </MarcoModal>
  )
}
