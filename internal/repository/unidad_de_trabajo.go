package repository

import (
	"context"

	"gorm.io/gorm"

	"ingenieria-y-calidad/internal/domain"
)

// unidadDeTrabajo implementa domain.UnidadDeTrabajo con una transacción de
// GORM: los repositorios que recibe fn se construyen sobre el mismo *gorm.DB
// transaccional, de modo que todas sus escrituras se confirman o se revierten
// juntas (US-17 RB-06).
type unidadDeTrabajo struct {
	db *gorm.DB
}

// NuevaUnidadDeTrabajo construye la unidad de trabajo sobre la conexión dada.
func NuevaUnidadDeTrabajo(db *gorm.DB) *unidadDeTrabajo {
	return &unidadDeTrabajo{db: db}
}

var _ domain.UnidadDeTrabajo = (*unidadDeTrabajo)(nil)

// Ejecutar abre la transacción, ejecuta fn y hace commit si fn no devuelve
// error. Ante un error o un pánico de fn, GORM hace rollback.
func (u *unidadDeTrabajo) Ejecutar(ctx context.Context, fn func(repos domain.RepositoriosTransaccion) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(repositoriosTransaccion{tx: tx})
	})
}

// repositoriosTransaccion entrega los repositorios ligados a la transacción.
type repositoriosTransaccion struct {
	tx *gorm.DB
}

func (r repositoriosTransaccion) Historias() domain.HistoriaRepository {
	return NuevoHistoriaRepository(r.tx)
}

func (r repositoriosTransaccion) Poker() domain.SesionPokerRepository {
	return NuevoPokerRepository(r.tx)
}

func (r repositoriosTransaccion) Estimaciones() domain.StoryEstimateRepository {
	return NuevoStoryEstimateRepository(r.tx)
}
