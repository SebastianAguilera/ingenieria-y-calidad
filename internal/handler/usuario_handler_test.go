package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ingenieria-y-calidad/internal/domain"
)

type mockServicioUsuarios struct {
	mock.Mock
}

func (m *mockServicioUsuarios) Crear(ctx context.Context, entrada domain.UsuarioInput) (*domain.Usuario, error) {
	ret := m.Called(ctx, entrada)
	var usuario *domain.Usuario
	if ret.Get(0) != nil {
		usuario = ret.Get(0).(*domain.Usuario)
	}
	return usuario, ret.Error(1)
}

func (m *mockServicioUsuarios) Listar(ctx context.Context) ([]domain.Usuario, error) {
	ret := m.Called(ctx)
	var usuarios []domain.Usuario
	if ret.Get(0) != nil {
		usuarios = ret.Get(0).([]domain.Usuario)
	}
	return usuarios, ret.Error(1)
}

func (m *mockServicioUsuarios) ObtenerPorID(ctx context.Context, id uint) (*domain.Usuario, error) {
	ret := m.Called(ctx, id)
	var usuario *domain.Usuario
	if ret.Get(0) != nil {
		usuario = ret.Get(0).(*domain.Usuario)
	}
	return usuario, ret.Error(1)
}

func nuevoRouterUsuarios(t *testing.T) (*gin.Engine, *mockServicioUsuarios) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	servicio := new(mockServicioUsuarios)
	r := gin.New()
	RegistrarRutasUsuarios(r, servicio)
	return r, servicio
}

func ejecutarUsuarios(t *testing.T, router *gin.Engine, metodo, ruta string, cuerpo any) *httptest.ResponseRecorder {
	t.Helper()
	var lector *bytes.Reader
	if cuerpo == nil {
		lector = bytes.NewReader(nil)
	} else if s, ok := cuerpo.(string); ok {
		lector = bytes.NewReader([]byte(s))
	} else {
		datos, err := json.Marshal(cuerpo)
		require.NoError(t, err)
		lector = bytes.NewReader(datos)
	}

	req, err := http.NewRequest(metodo, ruta, lector)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCrearUsuarioE01(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterUsuarios(t)
	servicio.On("Crear", mock.Anything, domain.UsuarioInput{Nombre: "Ana Pérez", Email: "dev@example.com"}).
		Return(&domain.Usuario{ID: 1, Nombre: "Ana Pérez", Email: "dev@example.com"}, nil)

	w := ejecutarUsuarios(t, router, http.MethodPost, "/api/usuarios", altaUsuarioDTO{
		Nombre: "Ana Pérez",
		Email:  "dev@example.com",
	})

	require.Equal(t, http.StatusCreated, w.Code)

	var respuesta respuestaUsuarioDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respuesta))
	assert.Equal(t, uint(1), respuesta.ID)
	assert.Equal(t, "Ana Pérez", respuesta.Nombre)
	assert.Equal(t, "dev@example.com", respuesta.Email)
}

func TestCrearUsuarioE02EmailDuplicado(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterUsuarios(t)
	servicio.On("Crear", mock.Anything, mock.Anything).
		Return(nil, domain.ErrUsuarioDuplicado)

	w := ejecutarUsuarios(t, router, http.MethodPost, "/api/usuarios", altaUsuarioDTO{
		Nombre: "Otro",
		Email:  "dev@example.com",
	})

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var errResp respuestaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	assert.Equal(t, codigoUsuarioDuplicado, errResp.Error)
	assert.Equal(t, "email", errResp.Campo)
}

func TestListarUsuariosE03(t *testing.T) {
	t.Parallel()

	router, servicio := nuevoRouterUsuarios(t)
	servicio.On("Listar", mock.Anything).
		Return([]domain.Usuario{
			{ID: 1, Nombre: "Ana Pérez", Email: "dev@example.com"},
			{ID: 2, Nombre: "Bruno", Email: "bruno@utn.edu.ar"},
		}, nil)

	w := ejecutarUsuarios(t, router, http.MethodGet, "/api/usuarios", nil)

	require.Equal(t, http.StatusOK, w.Code)

	var lista []respuestaUsuarioDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &lista))
	require.Len(t, lista, 2)
	assert.Equal(t, "dev@example.com", lista[0].Email)
	assert.Equal(t, "bruno@utn.edu.ar", lista[1].Email)
}
