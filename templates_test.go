package sesiones

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplates(t *testing.T) {

	tpl, err := NewConfirmacionBlanqueo()
	assert.Nil(t, err)

	UsuarioNombre := "Ornelita"
	URLConfirmacion := "www.sweet.com.ar/confirmacion?id=51651651651651651"

	html, err := tpl.String(UsuarioNombre, URLConfirmacion)
	assert.Nil(t, err)

	fmt.Println(html)
}
