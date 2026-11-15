package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

func TestUsuarioServiceCrear(t *testing.T) {
	t.Parallel()

	t.Run("crea un usuario independiente", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuarios.On("Crear", mock.Anything, mock.Anything).Return(nil).Run(asignarIDUsuario(1))

		usuario, err := servicio.Crear(context.Background(), domain.UsuarioInput{
			Nombre: "Ana Pérez",
			Email:  "ANA@EXAMPLE.COM",
		})

		require.NoError(t, err)
		require.NotNil(t, usuario)
		assert.Equal(t, uint(1), usuario.ID)
		assert.Equal(t, "ana@example.com", usuario.Email)
		usuarios.AssertExpectations(t)
	})

	t.Run("rechaza email duplicado", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuarios.On("Crear", mock.Anything, mock.Anything).Return(domain.ErrUsuarioDuplicado)

		usuario, err := servicio.Crear(context.Background(), domain.UsuarioInput{
			Nombre: "Ana Pérez",
			Email:  "ana@example.com",
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUsuarioDuplicado)
		assert.Nil(t, usuario)
	})

	t.Run("rechaza email invalido", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuario, err := servicio.Crear(context.Background(), domain.UsuarioInput{
			Nombre: "Ana Pérez",
			Email:  "no-es-un-mail",
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUsuarioEmailInvalido)
		assert.Nil(t, usuario)
		usuarios.AssertNotCalled(t, "Crear", mock.Anything, mock.Anything)
	})
}

func TestUsuarioServiceListar(t *testing.T) {
	t.Parallel()

	t.Run("lista usuarios vacios", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuarios.On("Listar", mock.Anything).Return([]domain.Usuario{}, nil)

		lista, err := servicio.Listar(context.Background())
		require.NoError(t, err)
		require.NotNil(t, lista)
		assert.Empty(t, lista)
	})

	t.Run("lista usuarios", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuarios.On("Listar", mock.Anything).Return([]domain.Usuario{
			{ID: 1, Nombre: "Ana", Email: "ana@example.com"},
		}, nil)

		lista, err := servicio.Listar(context.Background())
		require.NoError(t, err)
		assert.Len(t, lista, 1)
	})
}

func TestUsuarioServiceObtenerPorID(t *testing.T) {
	t.Parallel()

	t.Run("devuelve usuario", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuarios.On("ObtenerPorID", mock.Anything, uint(1)).Return(&domain.Usuario{ID: 1, Nombre: "Ana", Email: "ana@example.com"}, nil)

		usuario, err := servicio.ObtenerPorID(context.Background(), 1)
		require.NoError(t, err)
		require.NotNil(t, usuario)
		assert.Equal(t, uint(1), usuario.ID)
	})

	t.Run("usuario inexistente", func(t *testing.T) {
		t.Parallel()
		usuarios := nuevoMockUsuarioRepository(t)
		servicio := NewUsuarioService(usuarios)

		usuarios.On("ObtenerPorID", mock.Anything, uint(999)).Return(nil, domain.ErrUsuarioNoEncontrado)

		usuario, err := servicio.ObtenerPorID(context.Background(), 999)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUsuarioNoEncontrado)
		assert.Nil(t, usuario)
	})
}
