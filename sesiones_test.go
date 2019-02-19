package sesiones

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	h := Handler{}
	h.secretKey = []byte("secreto")

	// Creo un token
	token, err := h.newToken("marcos")
	assert.Nil(t, err)

	tokenString, err := token.SignedString(h.secretKey)
	assert.Nil(t, err)

	// Chequeo
	_, err = h.chequearToken(tokenString)
	assert.Nil(t, err)

}

func TestTokenVencido(t *testing.T) {

	h := Handler{}
	h.secretKey = []byte("secreto")

	// Creo un token
	token, err := h.newToken("marcos")
	assert.Nil(t, err)

	tokenString, err := token.SignedString(h.secretKey)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	// Chequeo, debería devolverme un token expirado
	_, err = h.chequearToken(tokenString)
	assert.NotNil(t, err)

}

// TestRequest prueba el funcionamiento del token dentro del request.
func TestRequestSinToken(t *testing.T) {

	// Envio un request sin token
	r, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(handleControlaSesiones)
	handler.ServeHTTP(rec, r)

	// Debería devolver no autorizado
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

}

func TestRequestConTokenValido(t *testing.T) {
	h := Handler{}
	h.secretKey = []byte("secreto")

	// Request CON token válido
	r2, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	token, err := h.newToken("marcos")
	assert.Nil(t, err)
	tokenString, err := token.SignedString(h.secretKey)
	assert.Nil(t, err)
	r2.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(handleControlaSesiones)
	handler.ServeHTTP(rec, r2)

	assert.Equal(t, http.StatusOK, rec.Code)
}
func TestRequestConTokenInválido(t *testing.T) {

	h := Handler{}
	h.secretKey = []byte("secreto")

	// Request CON token válido
	r2, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	token, err := h.newToken("marcos")
	assert.Nil(t, err)
	tokenString, err := token.SignedString([]byte("token inválido"))
	assert.Nil(t, err)
	r2.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(handleControlaSesiones)
	handler.ServeHTTP(rec, r2)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequestVencida(t *testing.T) {
	h := Handler{}
	h.secretKey = []byte("secreto")

	r, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	token, err := h.newToken("marcos")
	assert.Nil(t, err)
	tokenString, err := token.SignedString(h.secretKey)
	assert.Nil(t, err)
	r.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
	assert.Nil(t, err)

	time.Sleep(time.Second)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(handleControlaSesiones)
	handler.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func handleControlaSesiones(w http.ResponseWriter, r *http.Request) {
	h := Handler{}
	h.secretKey = []byte("secreto")

	// Extrae el token
	token, err := extraerToken(r)
	if err != nil {
		// Puede ser un error en el token, o que bien no esté presente la cookie
		http.Error(w, "error extrayendo token", http.StatusUnauthorized)
		return
	}

	// Chequea el token
	_, err = h.chequearToken(token)
	if err != nil {
		http.Error(w, "token no válido", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Ok"))

}
