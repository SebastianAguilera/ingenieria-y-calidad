package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidarDescripcionCriterio(t *testing.T) {
	t.Run("valida descripcion correcta", func(t *testing.T) {
		err := ValidarDescripcionCriterio("Debe validar el email")
		assert.NoError(t, err)
	})

	t.Run("rechaza descripcion vacia", func(t *testing.T) {
		err := ValidarDescripcionCriterio("   ")
		assert.ErrorIs(t, err, ErrDescripcionCriterioObligatoria)
	})

	t.Run("rechaza descripcion demasiado larga", func(t *testing.T) {
		larga := strings.Repeat("a", 501)
		err := ValidarDescripcionCriterio(larga)
		assert.ErrorIs(t, err, ErrDescripcionCriterioDemasiadoLarga)
	})
}
