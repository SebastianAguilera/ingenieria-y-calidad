package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ingenieria-y-calidad/internal/domain"
)

// ProyectoService implementa las reglas de negocio de la gestión de
// proyectos. Depende únicamente de las interfaces definidas en
// internal/domain, nunca de implementaciones concretas.
type ProyectoService struct {
	proyectos domain.ProyectoRepository
	usuarios  domain.UsuarioRepository
}

// limiteDescripcionProyecto es la cantidad máxima de caracteres que admite la
// descripción de un proyecto (RB-02 de US-F11).
const limiteDescripcionProyecto = 2000

// NewProyectoService construye el servicio de proyectos con inyeccion de
// dependencias.
func NewProyectoService(proyectos domain.ProyectoRepository, usuarios domain.UsuarioRepository) *ProyectoService {
	return &ProyectoService{
		proyectos: proyectos,
		usuarios:  usuarios,
	}
}

// Crear valida el alta, persiste el proyecto y vincula al usuario que lo da de
// alta en la misma transacción (US-F10, RB-03).
func (s *ProyectoService) Crear(ctx context.Context, usuarioID uint, nuevo domain.NuevoProyecto) (*domain.Proyecto, error) {
	if err := validarNombreProyecto(nuevo.Nombre); err != nil {
		return nil, err
	}
	if err := validarDescripcion(nuevo.Descripcion); err != nil {
		return nil, err
	}
	if err := validarFechas(nuevo.FechaInicio, nuevo.FechaFin); err != nil {
		return nil, err
	}

	creador, err := s.usuarios.ObtenerPorID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo creador del proyecto: %w", err)
	}

	proyecto := &domain.Proyecto{
		Nombre:      strings.TrimSpace(nuevo.Nombre),
		Descripcion: strings.TrimSpace(nuevo.Descripcion),
		FechaInicio: nuevo.FechaInicio,
		FechaFin:    nuevo.FechaFin,
		Estado:      domain.EstadoActivo,
	}

	equipo := []domain.Usuario{*creador}
	if err := s.proyectos.CrearConUsuarios(ctx, proyecto, equipo); err != nil {
		return nil, fmt.Errorf("creando proyecto: %w", err)
	}
	proyecto.Usuarios = equipo

	return proyecto, nil
}

// PerteneceA informa si la persona forma parte del equipo del proyecto
// (US-F10, RB-04). Un id invalido no tiene equipo posible: se responde false
// sin consultar la base; un fallo de la base se propaga para que el guard
// responda 500 en lugar de abrir el proyecto.
func (s *ProyectoService) PerteneceA(ctx context.Context, proyectoID, usuarioID uint) (bool, error) {
	if proyectoID == 0 || usuarioID == 0 {
		return false, nil
	}
	asociado, err := s.usuarios.EstaAsociado(ctx, proyectoID, usuarioID)
	if err != nil {
		return false, fmt.Errorf("verificando membresia del proyecto: %w", err)
	}
	return asociado, nil
}

// Listar devuelve los proyectos no dados de baja en los que participa el
// usuario.
func (s *ProyectoService) Listar(ctx context.Context, usuarioID uint) ([]domain.Proyecto, error) {
	proyectos, err := s.proyectos.ListarPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listando proyectos: %w", err)
	}
	if proyectos == nil {
		return []domain.Proyecto{}, nil
	}
	return proyectos, nil
}

// ObtenerPorID devuelve el detalle de un proyecto con sus usuarios.
func (s *ProyectoService) ObtenerPorID(ctx context.Context, id uint) (*domain.Proyecto, error) {
	proyecto, err := s.proyectos.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo proyecto: %w", err)
	}
	return proyecto, nil
}

// Actualizar modifica nombre y fechas de un proyecto activo.
func (s *ProyectoService) Actualizar(ctx context.Context, id uint, actualizacion domain.ActualizacionProyecto) (*domain.Proyecto, error) {
	if err := validarNombreProyecto(actualizacion.Nombre); err != nil {
		return nil, err
	}
	if err := validarDescripcion(actualizacion.Descripcion); err != nil {
		return nil, err
	}
	if err := validarFechas(actualizacion.FechaInicio, actualizacion.FechaFin); err != nil {
		return nil, err
	}

	proyecto, err := s.proyectos.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo proyecto: %w", err)
	}
	if err := exigirProyectoActivo(proyecto); err != nil {
		return nil, err
	}

	proyecto.Nombre = strings.TrimSpace(actualizacion.Nombre)
	proyecto.Descripcion = strings.TrimSpace(actualizacion.Descripcion)
	proyecto.FechaInicio = actualizacion.FechaInicio
	proyecto.FechaFin = actualizacion.FechaFin

	if err := s.proyectos.Actualizar(ctx, proyecto); err != nil {
		return nil, fmt.Errorf("actualizando proyecto: %w", err)
	}

	return proyecto, nil
}

// CambiarEstado transiciona el estado entre Activo y Cerrado.
func (s *ProyectoService) CambiarEstado(ctx context.Context, id uint, estado domain.EstadoProyecto) (*domain.Proyecto, error) {
	if !estado.Valido() {
		return nil, domain.ErrEstadoInvalido
	}

	proyecto, err := s.proyectos.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obteniendo proyecto: %w", err)
	}

	if proyecto.Estado == estado {
		return proyecto, nil
	}

	if estado == domain.EstadoCerrado && proyecto.FechaFin == nil {
		return nil, domain.ErrFechaFinRequerida
	}

	proyecto.Estado = estado

	if err := s.proyectos.Actualizar(ctx, proyecto); err != nil {
		return nil, fmt.Errorf("actualizando estado del proyecto: %w", err)
	}

	return proyecto, nil
}

// Eliminar da de baja logicamente un proyecto que no tiene historial asociado.
func (s *ProyectoService) Eliminar(ctx context.Context, id uint) error {
	if _, err := s.proyectos.ObtenerPorID(ctx, id); err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}

	tieneHistorial, err := s.proyectos.TieneHistorial(ctx, id)
	if err != nil {
		return fmt.Errorf("consultando historial del proyecto: %w", err)
	}
	if tieneHistorial {
		return domain.ErrProyectoConHistorial
	}

	if err := s.proyectos.Eliminar(ctx, id); err != nil {
		return fmt.Errorf("eliminando proyecto: %w", err)
	}
	return nil
}

// AgregarUsuario asocia una persona ya registrada al equipo del proyecto con su rol.
func (s *ProyectoService) AgregarUsuario(ctx context.Context, proyectoID uint, asociacion domain.AsociacionUsuario) (*domain.Usuario, error) {
	if asociacion.Rol == "" || !asociacion.Rol.Valido() {
		return nil, domain.ErrRolInvalido
	}

	proyecto, err := s.proyectos.ObtenerPorID(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo proyecto: %w", err)
	}
	if err := exigirProyectoActivo(proyecto); err != nil {
		return nil, err
	}

	usuario, err := s.usuarios.ObtenerPorID(ctx, asociacion.UsuarioID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo usuario: %w", err)
	}

	asociado, err := s.usuarios.EstaAsociado(ctx, proyectoID, usuario.ID)
	if err != nil {
		return nil, fmt.Errorf("verificando asociacion del usuario: %w", err)
	}
	if asociado {
		return nil, domain.ErrUsuarioYaAsociado
	}

	if err := s.usuarios.Vincular(ctx, proyectoID, usuario.ID, asociacion.Rol); err != nil {
		return nil, fmt.Errorf("vinculando usuario: %w", err)
	}

	return usuario, nil
}

// QuitarUsuario desasocia una persona del equipo del proyecto.
func (s *ProyectoService) QuitarUsuario(ctx context.Context, proyectoID, usuarioID uint) error {
	proyecto, err := s.proyectos.ObtenerPorID(ctx, proyectoID)
	if err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	if err := exigirProyectoActivo(proyecto); err != nil {
		return err
	}

	asociado, err := s.usuarios.EstaAsociado(ctx, proyectoID, usuarioID)
	if err != nil {
		return fmt.Errorf("verificando asociacion del usuario: %w", err)
	}
	if !asociado {
		return domain.ErrUsuarioNoAsociado
	}

	if err := s.usuarios.Desvincular(ctx, proyectoID, usuarioID); err != nil {
		return fmt.Errorf("desvinculando usuario: %w", err)
	}
	return nil
}

func validarNombreProyecto(nombre string) error {
	recortado := strings.TrimSpace(nombre)
	if len(recortado) == 0 {
		return domain.ErrNombreObligatorio
	}
	if len([]rune(recortado)) > 150 {
		return domain.ErrNombreDemasiadoLargo
	}
	return nil
}

func validarDescripcion(descripcion string) error {
	if len([]rune(strings.TrimSpace(descripcion))) > limiteDescripcionProyecto {
		return domain.ErrDescripcionDemasiadoLarga
	}
	return nil
}

func validarFechas(inicio time.Time, fin *time.Time) error {
	if inicio.IsZero() {
		return domain.ErrFechaInicioObligatoria
	}
	if fin != nil && fin.Before(inicio) {
		return domain.ErrFechaFinAnteriorAlInicio
	}
	return nil
}

func validarUsuario(entrada domain.UsuarioInput) error {
	recortado := strings.TrimSpace(entrada.Nombre)
	if len(recortado) == 0 {
		return domain.ErrUsuarioNombreVacio
	}
	if len([]rune(recortado)) > 100 {
		return domain.ErrUsuarioNombreLargo
	}
	if !domain.EmailValido(entrada.Email) {
		return domain.ErrUsuarioEmailInvalido
	}
	return nil
}

func exigirProyectoActivo(proyecto *domain.Proyecto) error {
	if proyecto.Estado != domain.EstadoActivo {
		return domain.ErrProyectoCerrado
	}
	return nil
}

func isNoEncontrado(err error) bool {
	return errors.Is(err, domain.ErrUsuarioNoEncontrado)
}
