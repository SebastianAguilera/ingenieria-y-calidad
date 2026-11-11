import type { ActualizacionHistoria, Historia } from '../../../api/types'
import Button from '../../../ui/Button'
import { ESTADOS_HISTORIA, PRIORIDADES_HISTORIA, claseEstado, clasePrioridad } from './etiquetas'
import SelectPill from './SelectPill'

type TablaBacklogProps = {
  historias: Historia[]
  onCambiar: (historia: Historia, cambios: ActualizacionHistoria) => void
  onEditar: (historia: Historia) => void
  guardandoId: number | null
}

export default function TablaBacklog({
  historias,
  onCambiar,
  onEditar,
  guardandoId,
}: TablaBacklogProps) {
  return (
    <div className="overflow-x-auto rounded-tile border border-border bg-surface">
      <table className="w-full min-w-[48rem] text-left text-sm">
        <thead className="border-b border-border text-ink-soft">
          <tr>
            <th scope="col" className="px-6 py-4 font-semibold">
              Código
            </th>
            <th scope="col" className="px-6 py-4 font-semibold">
              Título
            </th>
            <th scope="col" className="px-6 py-4 font-semibold">
              Estado
            </th>
            <th scope="col" className="px-6 py-4 font-semibold">
              Prioridad
            </th>
            <th scope="col" className="px-6 py-4 font-semibold">
              Acciones
            </th>
          </tr>
        </thead>
        <tbody>
          {historias.map((historia) => {
            const guardando = guardandoId === historia.id
            return (
              <tr
                key={historia.id}
                className="border-b border-border last:border-b-0 hover:bg-[var(--primary-100)]"
              >
                <td className="px-6 py-4">
                  <span className="font-mono text-xs font-semibold text-ink-soft">
                    {historia.codigo}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <span className="font-medium text-ink">{historia.titulo}</span>
                </td>
                <td className="px-6 py-4">
                  <SelectPill
                    label={`Estado de ${historia.codigo}`}
                    value={historia.estado}
                    opciones={ESTADOS_HISTORIA}
                    onChange={(estado) => onCambiar(historia, { estado })}
                    disabled={guardando}
                    className={claseEstado(historia.estado)}
                  />
                </td>
                <td className="px-6 py-4">
                  <SelectPill
                    label={`Prioridad de ${historia.codigo}`}
                    value={historia.prioridad}
                    opciones={PRIORIDADES_HISTORIA}
                    onChange={(prioridad) => onCambiar(historia, { prioridad })}
                    disabled={guardando}
                    className={clasePrioridad(historia.prioridad)}
                  />
                </td>
                <td className="px-6 py-4">
                  <Button
                    variante="ghost"
                    onClick={() => onEditar(historia)}
                    aria-label={`Editar ${historia.codigo}`}
                    disabled={guardando}
                  >
                    Editar
                  </Button>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
