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
	// Creo un token
	token, err := newToken("marcos", time.Minute*10)
	assert.Nil(t, err)

	tokenString, err := token.SignedString(secret)
	assert.Nil(t, err)

	// Chequeo
	_, err = chequearToken(tokenString, time.Minute*10)
	assert.Nil(t, err)

}

func TestTokenVencido(t *testing.T) {
	// Creo un token
	token, err := newToken("marcos", time.Millisecond)
	assert.Nil(t, err)

	tokenString, err := token.SignedString(secret)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	// Chequeo, debería devolverme un token expirado
	_, err = chequearToken(tokenString, time.Millisecond*10)
	assert.NotNil(t, err)

}

// TestRequest prueba el funcionamiento del token dentro del request.
func TestRequestSinToken(t *testing.T) {

	// Envio un request sin token
	r, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleControlaSesiones)
	handler.ServeHTTP(rec, r)

	// Debería devolver no autorizado
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

}

func TestRequestConTokenValido(t *testing.T) {
	// Request CON token válido
	r2, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	token, err := newToken("marcos", time.Second*10)
	assert.Nil(t, err)
	tokenString, err := token.SignedString(secret)
	assert.Nil(t, err)
	r2.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleControlaSesiones)
	handler.ServeHTTP(rec, r2)

	assert.Equal(t, http.StatusOK, rec.Code)
}
func TestRequestConTokenInválido(t *testing.T) {
	// Request CON token válido
	r2, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	token, err := newToken("marcos", time.Second*10)
	assert.Nil(t, err)
	tokenString, err := token.SignedString([]byte("token inválido"))
	assert.Nil(t, err)
	r2.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
	assert.Nil(t, err)

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleControlaSesiones)
	handler.ServeHTTP(rec, r2)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequestVencida(t *testing.T) {

	r, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"pido": "gancho"}`))
	assert.Nil(t, err)

	token, err := newToken("marcos", time.Millisecond)
	assert.Nil(t, err)
	tokenString, err := token.SignedString(secret)
	assert.Nil(t, err)
	r.AddCookie(&http.Cookie{Name: "token", Value: tokenString})
	assert.Nil(t, err)

	time.Sleep(time.Second)
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleControlaSesiones)
	handler.ServeHTTP(rec, r)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func HandleControlaSesiones(w http.ResponseWriter, r *http.Request) {

	// Extrae el token
	token, err := extraerToken(r)
	if err != nil {
		// Puede ser un error en el token, o que bien no esté presente la cookie
		http.Error(w, "error extrayendo token", http.StatusUnauthorized)
		return
	}

	// Chequea el token
	_, err = chequearToken(token, time.Minute)
	if err != nil {
		http.Error(w, "token no válido", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Ok"))

}
