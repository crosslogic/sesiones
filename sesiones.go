package sesiones

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// chequearToken determina si el token ingresado es válido, si es así le
// actualiza la hora de vencimiento.
func (h *Handler) chequearToken(token string) (tokenOut *jwt.Token, err error) {
	t1, err := h.parseToken(token)
	if err != nil {
		return tokenOut, errors.Wrap(err, "chequeando token")
	}

	// Infiero tipo
	claims := t1.Claims.(jwt.MapClaims)
	t2, err := h.newToken(claims["userID"].(string))
	if err != nil {
		return tokenOut, errors.Wrap(err, "creando nuevo token")
	}

	return t2, nil
}

// setToken pega la cookie a la response.
func (h *Handler) setToken(w http.ResponseWriter, token *jwt.Token) (err error) {
	cookie := http.Cookie{}
	//cookie.HttpOnly = true
	cookie.Name = "token"
	cookie.Path = "/"
	cookie.Value, err = token.SignedString(h.secretKey)
	if err != nil {
		return errors.Wrap(err, "firmando token")
	}
	cookie.Expires = time.Now().Add(h.DuracionSesion)
	http.SetCookie(w, &cookie)

	return
}

// setTokenVencido pega una cookie vencida a la response.
func (h *Handler) setTokenVencido(w http.ResponseWriter, token *jwt.Token) (err error) {

	cookie := http.Cookie{}
	//cookie.HttpOnly = true
	cookie.Name = "token"
	cookie.Path = "/"
	cookie.Value, err = token.SignedString(h.secretKey)
	if err != nil {
		return errors.Wrap(err, "firmando token")
	}

	cookie.MaxAge = -1

	http.SetCookie(w, &cookie)

	return
}

// parseToken transforma el string en un jwt.Token
func (h *Handler) parseToken(tokenString string) (token *jwt.Token, err error) {

	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return h.secretKey, nil
	})
	if err != nil {
		return token, errors.Wrap(err, "parseando JWT")
	}
	// Corroboro
	_, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return token, errors.New("token inválido")
	}
	return
}

// NewToken Handlers
func (h *Handler) newToken(userID string) (token *jwt.Token, err error) {
	// Create the token
	token = jwt.New(jwt.SigningMethodHS256)

	// Create a map to store our claims
	claims := token.Claims.(jwt.MapClaims)

	// Set token claims
	claims["userID"] = userID
	claims["exp"] = time.Now().Add(h.DuracionSesion).Unix()

	return
}

// usuarioID devuelve el campo Nombre para el usuario de la sesión
func (h *Handler) usuarioID(r *http.Request) (id string, err error) {
	tokenString, err := extraerToken(r)
	if err != nil {
		return id, errors.Wrap(err, "extrayendo token de request")
	}

	token, err := h.parseToken(tokenString)
	if err != nil {
		return id, errors.Wrap(err, "parseando token")
	}

	// Infiero tipo
	claims := token.Claims.(jwt.MapClaims)
	id = claims["userID"].(string)

	return id, err

}

// devuelve el token que está en la request
func extraerToken(r *http.Request) (token string, err error) {
	c, err := r.Cookie("token")
	if err != nil {
		return token, errors.Wrap(err, "buscando cookie")
	}

	return c.Value, nil
}
