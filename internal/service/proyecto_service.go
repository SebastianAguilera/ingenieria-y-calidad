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
	proyectos   domain.ProyectoRepository
	integrantes domain.IntegranteRepository
}

// NewProyectoService construye el servicio de proyectos con inyeccion de
// dependencias.
func NewProyectoService(proyectos domain.ProyectoRepository, integrantes domain.IntegranteRepository) *ProyectoService {
	return &ProyectoService{
		proyectos:   proyectos,
		integrantes: integrantes,
	}
}

// Crear valida el alta y persiste el proyecto junto con sus integrantes
// iniciales en una unica operacion transaccional.
func (s *ProyectoService) Crear(ctx context.Context, nuevo domain.NuevoProyecto) (*domain.Proyecto, error) {
	if err := validarNombreProyecto(nuevo.Nombre); err != nil {
		return nil, err
	}
	if err := validarFechas(nuevo.FechaInicio, nuevo.FechaFin); err != nil {
		return nil, err
	}

	integrantes, err := s.resolverIntegrantes(ctx, nuevo.Integrantes)
	if err != nil {
		return nil, err
	}

	proyecto := &domain.Proyecto{
		Nombre:      strings.TrimSpace(nuevo.Nombre),
		FechaInicio: nuevo.FechaInicio,
		FechaFin:    nuevo.FechaFin,
		Estado:      domain.EstadoActivo,
		Integrantes: integrantes,
	}

	if err := s.proyectos.CrearConIntegrantes(ctx, proyecto, integrantes); err != nil {
		return nil, fmt.Errorf("creando proyecto: %w", err)
	}

	return proyecto, nil
}

// Listar devuelve los proyectos no dados de baja.
func (s *ProyectoService) Listar(ctx context.Context) ([]domain.Proyecto, error) {
	proyectos, err := s.proyectos.Listar(ctx)
	if err != nil {
		return nil, fmt.Errorf("listando proyectos: %w", err)
	}
	if proyectos == nil {
		return []domain.Proyecto{}, nil
	}
	return proyectos, nil
}

// ObtenerPorID devuelve el detalle de un proyecto con sus integrantes.
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

// AgregarIntegrante asocia una persona al equipo del proyecto. Si el email ya
// existe en el sistema se reutiliza el registro existente.
func (s *ProyectoService) AgregarIntegrante(ctx context.Context, proyectoID uint, entrada domain.IntegranteInput) (*domain.Integrante, error) {
	if err := validarIntegrante(entrada); err != nil {
		return nil, err
	}

	proyecto, err := s.proyectos.ObtenerPorID(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("obteniendo proyecto: %w", err)
	}
	if err := exigirProyectoActivo(proyecto); err != nil {
		return nil, err
	}

	email := domain.NormalizarEmail(entrada.Email)
	nombre := strings.TrimSpace(entrada.Nombre)

	integrante, err := s.integrantes.ObtenerPorEmail(ctx, email)
	switch {
	case err == nil:
		asociado, err := s.integrantes.EstaAsociado(ctx, proyectoID, integrante.ID)
		if err != nil {
			return nil, fmt.Errorf("verificando asociacion del integrante: %w", err)
		}
		if asociado {
			return nil, domain.ErrIntegranteYaAsociado
		}
		if err := s.integrantes.ActualizarNombre(ctx, integrante.ID, nombre); err != nil {
			return nil, fmt.Errorf("actualizando nombre del integrante: %w", err)
		}
		if err := s.integrantes.Vincular(ctx, proyectoID, integrante.ID); err != nil {
			return nil, fmt.Errorf("vinculando integrante: %w", err)
		}
		integrante.Nombre = nombre
		return integrante, nil

	case isNoEncontrado(err):
		integrante = &domain.Integrante{Nombre: nombre, Email: email}
		if err := s.integrantes.Crear(ctx, integrante); err != nil {
			return nil, fmt.Errorf("creando integrante: %w", err)
		}
		if err := s.integrantes.Vincular(ctx, proyectoID, integrante.ID); err != nil {
			return nil, fmt.Errorf("vinculando integrante: %w", err)
		}
		return integrante, nil

	default:
		return nil, fmt.Errorf("buscando integrante: %w", err)
	}
}

// QuitarIntegrante desasocia una persona del equipo del proyecto.
func (s *ProyectoService) QuitarIntegrante(ctx context.Context, proyectoID, integranteID uint) error {
	proyecto, err := s.proyectos.ObtenerPorID(ctx, proyectoID)
	if err != nil {
		return fmt.Errorf("obteniendo proyecto: %w", err)
	}
	if err := exigirProyectoActivo(proyecto); err != nil {
		return err
	}

	asociado, err := s.integrantes.EstaAsociado(ctx, proyectoID, integranteID)
	if err != nil {
		return fmt.Errorf("verificando asociacion del integrante: %w", err)
	}
	if !asociado {
		return domain.ErrIntegranteNoAsociado
	}

	if err := s.integrantes.Desvincular(ctx, proyectoID, integranteID); err != nil {
		return fmt.Errorf("desvinculando integrante: %w", err)
	}
	return nil
}

func (s *ProyectoService) resolverIntegrantes(ctx context.Context, entradas []domain.IntegranteInput) ([]domain.Integrante, error) {
	if len(entradas) == 0 {
		return []domain.Integrante{}, nil
	}

	tipos, err := validarIntegrantes(entradas)
	if err != nil {
		return nil, err
	}

	integrantes := make([]domain.Integrante, 0, len(entradas))
	for _, tipo := range tipos {
		integrante, err := s.integrantes.ObtenerPorEmail(ctx, tipo.email)
		switch {
		case err == nil:
			if err := s.integrantes.ActualizarNombre(ctx, integrante.ID, tipo.nombre); err != nil {
				return nil, fmt.Errorf("actualizando nombre del integrante: %w", err)
			}
			integrantes = append(integrantes, *integrante)
		case isNoEncontrado(err):
			integrantes = append(integrantes, domain.Integrante{Nombre: tipo.nombre, Email: tipo.email})
		default:
			return nil, fmt.Errorf("buscando integrante: %w", err)
		}
	}

	return integrantes, nil
}

type integranteNormalizado struct {
	nombre string
	email  string
}

// validarIntegrantes ejecuta RB-07 y RB-08 sobre la lista completa antes de
// tocar el repositorio, de modo que una entrada invalida no provoque escrituras
// parciales y el error sea determinista ante varias entradas invalidas.
func validarIntegrantes(entradas []domain.IntegranteInput) ([]integranteNormalizado, error) {
	vistos := make(map[string]struct{}, len(entradas))
	normalizados := make([]integranteNormalizado, 0, len(entradas))

	for _, entrada := range entradas {
		if err := validarIntegrante(entrada); err != nil {
			return nil, err
		}

		email := domain.NormalizarEmail(entrada.Email)
		if _, duplicado := vistos[email]; duplicado {
			return nil, domain.ErrIntegranteYaAsociado
		}
		vistos[email] = struct{}{}

		normalizados = append(normalizados, integranteNormalizado{
			nombre: strings.TrimSpace(entrada.Nombre),
			email:  email,
		})
	}

	return normalizados, nil
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

func validarFechas(inicio time.Time, fin *time.Time) error {
	if inicio.IsZero() {
		return domain.ErrFechaInicioObligatoria
	}
	if fin != nil && fin.Before(inicio) {
		return domain.ErrFechaFinAnteriorAlInicio
	}
	return nil
}

func validarIntegrante(entrada domain.IntegranteInput) error {
	recortado := strings.TrimSpace(entrada.Nombre)
	if len(recortado) == 0 {
		return domain.ErrIntegranteNombreVacio
	}
	if len([]rune(recortado)) > 100 {
		return domain.ErrIntegranteNombreLargo
	}
	if !domain.EmailValido(entrada.Email) {
		return domain.ErrIntegranteEmailInvalido
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
	return errors.Is(err, domain.ErrIntegranteNoEncontrado)
}
