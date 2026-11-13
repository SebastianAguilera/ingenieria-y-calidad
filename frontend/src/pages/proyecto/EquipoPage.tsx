import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { RolProyecto, UsuarioProyecto } from '../../api/types'
import PageHeader from '../../components/PageHeader'
import { useSesion } from '../../session/sesion'
import Alert from '../../ui/Alert'
import Button from '../../ui/Button'
import Card from '../../ui/Card'
import ModalConfirmacion from './sprints/ModalConfirmacion'
import { claseRol, ROLES } from './equipo/etiquetasRol'
import ModalIntegrante from './equipo/ModalIntegrante'
import { useEquipo } from './equipo/useEquipo'

type Confirmacion = { usuario: UsuarioProyecto; esSalida: boolean }

function parsearProyectoId(id: string | undefined): number | null {
  if (id === undefined) {
    return null
  }
  const numero = Number(id)
  return Number.isInteger(numero) && numero > 0 ? numero : null
}

export default function EquipoPage() {
  const { id } = useParams()
  const proyectoId = parsearProyectoId(id)
  const { usuarioActivo } = useSesion()
  const navegar = useNavigate()
  const { estado, reintentar, agregar, cambiarRol, quitar, procesandoId } = useEquipo(proyectoId)
  const [altaAbierta, setAltaAbierta] = useState(false)
  const [confirmacion, setConfirmacion] = useState<Confirmacion | null>(null)
  const [errorOperacion, setErrorOperacion] = useState<string | null>(null)
  const [mensajeExito, setMensajeExito] = useState<string | null>(null)

  const esLider = estado.tipo === 'listo' && estado.rolActual === 'LIDER'
  const cerrado = estado.tipo === 'listo' && estado.cerrado
  const puedeGestionar = esLider && !cerrado

  async function cambiarRolDe(usuario: UsuarioProyecto, rol: RolProyecto) {
    const rolActual = usuario.rol ?? 'MIEMBRO'
    if (rol === rolActual) {
      return
    }

    setErrorOperacion(null)
    setMensajeExito(null)
    const resultado = await cambiarRol(usuario.id, rol)
    if (!resultado.ok) {
      setErrorOperacion(resultado.mensaje)
    }
  }

  async function confirmarAccion() {
    if (confirmacion === null) {
      return
    }

    setErrorOperacion(null)
    setMensajeExito(null)
    const resultado = await quitar(confirmacion.usuario.id)
    setConfirmacion(null)

    if (!resultado.ok) {
      setErrorOperacion(resultado.mensaje)
      return
    }

    if (confirmacion.esSalida) {
      navegar('/proyectos')
      return
    }

    setMensajeExito('Integrante quitado del proyecto.')
  }

  return (
    <section>
      <PageHeader
        title="Equipo"
        description="Consultá quién integra el proyecto y gestioná sus roles."
      />

      {errorOperacion && (
        <div className="mb-5">
          <Alert tono="error">{errorOperacion}</Alert>
        </div>
      )}

      {mensajeExito && (
        <div className="mb-5">
          <Alert tono="success">{mensajeExito}</Alert>
        </div>
      )}

      {estado.tipo === 'cargando' && (
        <p role="status" className="text-ink-soft">
          Cargando equipo…
        </p>
      )}

      {estado.tipo === 'error' && (
        <div className="flex flex-col items-start gap-4">
          <Alert tono="error">{estado.mensaje}</Alert>
          {estado.reintentable && (
            <Button variante="secondary" onClick={reintentar}>
              Reintentar
            </Button>
          )}
        </div>
      )}

      {estado.tipo === 'listo' && (
        <>
          {cerrado && (
            <div className="mb-5">
              <Alert tono="info">
                Este proyecto está cerrado. Reabrí el proyecto para gestionar el equipo.
              </Alert>
            </div>
          )}

          <Card
            title="Integrantes"
            description={
              estado.integrantes.length === 0
                ? 'Todavía no hay integrantes para mostrar.'
                : `${estado.integrantes.length} personas en el proyecto.`
            }
          >
            {puedeGestionar && (
              <div className="mb-5 flex flex-wrap justify-end">
                <Button onClick={() => setAltaAbierta(true)}>Agregar integrante</Button>
              </div>
            )}

            {estado.integrantes.length > 0 && (
              <div className="overflow-x-auto rounded-tile border border-border bg-surface">
                <table className="w-full min-w-[40rem] text-left text-sm">
                  <thead className="border-b border-border text-ink-soft">
                    <tr>
                      <th scope="col" className="px-6 py-4 font-semibold">
                        Nombre
                      </th>
                      <th scope="col" className="px-6 py-4 font-semibold">
                        Email
                      </th>
                      <th scope="col" className="px-6 py-4 font-semibold">
                        Rol
                      </th>
                      <th scope="col" className="px-6 py-4 font-semibold">
                        Acciones
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {estado.integrantes.map((integrante) => {
                      const rol = integrante.rol ?? 'MIEMBRO'
                      const esPropia = integrante.id === usuarioActivo?.id
                      return (
                        <tr
                          key={integrante.id}
                          className="border-b border-border last:border-b-0"
                        >
                          <td className="px-6 py-5 font-medium text-ink">
                            {integrante.nombre}
                            {esPropia && <span className="text-ink-soft"> (vos)</span>}
                          </td>
                          <td className="px-6 py-5">{integrante.email}</td>
                          <td className="px-6 py-5">
                            {puedeGestionar ? (
                              <select
                                aria-label={`Rol de ${integrante.nombre}`}
                                value={rol}
                                disabled={procesandoId === integrante.id}
                                onChange={(evento) =>
                                  void cambiarRolDe(integrante, evento.target.value as RolProyecto)
                                }
                                className="rounded-pill border border-border bg-[var(--surface-2)] px-3 py-1.5 text-xs font-semibold text-ink focus:border-primary focus:outline-none disabled:opacity-60"
                              >
                                {ROLES.map((opcion) => (
                                  <option key={opcion} value={opcion}>
                                    {opcion}
                                  </option>
                                ))}
                              </select>
                            ) : (
                              <span
                                className={`inline-flex items-center rounded-pill border px-3 py-1.5 text-xs font-semibold ${claseRol(rol)}`}
                              >
                                {rol}
                              </span>
                            )}
                          </td>
                          <td className="px-6 py-5">
                            <div className="flex flex-wrap gap-2">
                              {puedeGestionar && !esPropia && (
                                <Button
                                  variante="danger"
                                  tamano="sm"
                                  aria-label={`Quitar a ${integrante.nombre}`}
                                  disabled={procesandoId === integrante.id}
                                  onClick={() =>
                                    setConfirmacion({ usuario: integrante, esSalida: false })
                                  }
                                >
                                  Quitar
                                </Button>
                              )}
                              {esPropia && !cerrado && (
                                <Button
                                  variante="ghost"
                                  tamano="sm"
                                  disabled={procesandoId === integrante.id}
                                  onClick={() =>
                                    setConfirmacion({ usuario: integrante, esSalida: true })
                                  }
                                >
                                  Salir del proyecto
                                </Button>
                              )}
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </>
      )}

      {altaAbierta && (
        <ModalIntegrante
          onCerrar={() => setAltaAbierta(false)}
          onGuardar={async (email, rol) => {
            setErrorOperacion(null)
            setMensajeExito(null)
            const resultado = await agregar(email, rol)
            if (resultado.ok) {
              setAltaAbierta(false)
              setMensajeExito('Integrante agregado al proyecto.')
            }
            return resultado
          }}
        />
      )}

      {confirmacion && (
        <ModalConfirmacion
          titulo={confirmacion.esSalida ? 'Salir del proyecto' : 'Quitar integrante'}
          mensaje={
            confirmacion.esSalida
              ? '¿Querés salir del proyecto? Dejarás de ver su información.'
              : `¿Querés quitar a "${confirmacion.usuario.nombre}" del proyecto?`
          }
          textoConfirmar={confirmacion.esSalida ? 'Salir' : 'Quitar'}
          deshabilitado={procesandoId !== null}
          onConfirmar={() => void confirmarAccion()}
          onCerrar={() => setConfirmacion(null)}
        />
      )}
    </section>
  )
}
