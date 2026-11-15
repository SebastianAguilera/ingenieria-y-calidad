import { useState } from 'react'
import type { FormEvent } from 'react'
import type { AltaDefecto, Historia, Sprint } from '../../../api/types'
import Alert from '../../../ui/Alert'
import Button from '../../../ui/Button'
import type { ResultadoOperacion } from '../sprints/tipos'
import { CampoAreaTexto, CampoSelect } from './campos'
import { SEVERIDADES_DEFECTO, etiquetaSeveridad } from './etiquetasDefecto'
import MarcoModal from './MarcoModal'
import { validarAltaDefecto } from './validacionDefecto'
import type { CamposAltaDefecto, ErroresAltaDefecto } from './validacionDefecto'

type ModalDefectoProps = {
  historias: Historia[]
  sprints: Sprint[]
  onGuardar: (alta: AltaDefecto) => Promise<ResultadoOperacion>
  onCerrar: () => void
}

const CAMPOS: ReadonlyArray<keyof CamposAltaDefecto> = [
  'descripcion',
  'severidad',
  'historia_id',
  'sprint_deteccion_id',
]

function esCampo(valor: string | undefined): valor is keyof CamposAltaDefecto {
  return CAMPOS.some((campo) => campo === valor)
}

export default function ModalDefecto({ historias, sprints, onGuardar, onCerrar }: ModalDefectoProps) {
  const [campos, setCampos] = useState<CamposAltaDefecto>({
    descripcion: '',
    severidad: '',
    historia_id: '',
    sprint_deteccion_id: '',
  })
  const [errores, setErrores] = useState<ErroresAltaDefecto>({})
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null)
  const [enviando, setEnviando] = useState(false)

  const faltanHistorias = historias.length === 0
  const faltanSprints = sprints.length === 0

  function cambiar(campo: keyof CamposAltaDefecto, valor: string) {
    setCampos((anteriores) => ({ ...anteriores, [campo]: valor }))
  }

  async function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault()
    if (enviando) {
      return
    }

    const { severidad } = campos
    const erroresCliente = validarAltaDefecto(campos)
    if (Object.keys(erroresCliente).length > 0 || severidad === '') {
      setErrores(erroresCliente)
      return
    }

    setEnviando(true)
    setErrorGeneral(null)
    setErrores({})

    try {
      const resultado = await onGuardar({
        descripcion: campos.descripcion.trim(),
        severidad,
        historia_id: Number(campos.historia_id),
        sprint_deteccion_id: Number(campos.sprint_deteccion_id),
      })

      if (!resultado.ok) {
        if (esCampo(resultado.campo)) {
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
    <MarcoModal titulo="Nuevo defecto" onCerrar={onCerrar}>
      <form noValidate onSubmit={enviar} className="mt-5 flex flex-col gap-4">
        {faltanHistorias && (
          <Alert tono="info">
            No hay historias en el proyecto. Creá una historia antes de registrar un defecto.
          </Alert>
        )}
        {faltanSprints && (
          <Alert tono="info">
            No hay sprints en el proyecto. Creá un sprint antes de registrar un defecto.
          </Alert>
        )}
        {errorGeneral && <Alert tono="error">{errorGeneral}</Alert>}

        <CampoAreaTexto
          label="Descripción"
          value={campos.descripcion}
          onChange={(valor) => cambiar('descripcion', valor)}
          error={errores.descripcion}
          autoFocus
        />

        <div className="grid gap-4 sm:grid-cols-2">
          <CampoSelect
            label="Severidad"
            value={campos.severidad}
            onChange={(valor) => cambiar('severidad', valor)}
            placeholder="Seleccioná una severidad"
            opciones={SEVERIDADES_DEFECTO.map((severidad) => ({
              valor: severidad,
              texto: etiquetaSeveridad(severidad),
            }))}
            error={errores.severidad}
          />
          <CampoSelect
            label="Historia"
            value={campos.historia_id}
            onChange={(valor) => cambiar('historia_id', valor)}
            placeholder="Seleccioná una historia"
            opciones={historias.map((historia) => ({
              valor: String(historia.id),
              texto: `${historia.codigo} — ${historia.titulo}`,
            }))}
            error={errores.historia_id}
          />
        </div>

        <CampoSelect
          label="Sprint de detección"
          value={campos.sprint_deteccion_id}
          onChange={(valor) => cambiar('sprint_deteccion_id', valor)}
          placeholder="Seleccioná un sprint"
          opciones={sprints.map((sprint) => ({
            valor: String(sprint.id),
            texto: sprint.identificador,
          }))}
          error={errores.sprint_deteccion_id}
        />

        <div className="mt-2 flex flex-wrap justify-end gap-3">
          <Button variante="ghost" onClick={onCerrar} disabled={enviando}>
            Cancelar
          </Button>
          <Button type="submit" disabled={enviando || faltanHistorias || faltanSprints}>
            {enviando ? 'Guardando…' : 'Registrar defecto'}
          </Button>
        </div>
      </form>
    </MarcoModal>
  )
}
