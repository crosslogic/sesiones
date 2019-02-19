package sesiones

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplates(t *testing.T) {

	tpl, err := NewMailTemplate(defaultBlanqueoTemplate, "www.sweet.com.ar/#/auth/confirmar_blanqueo")
	assert.Nil(t, err)

	UsuarioNombre := "Ornelita"
	id := "51651651651651651"

	html, err := tpl.body(UsuarioNombre, id)
	assert.Nil(t, err)

	assert.True(t, len(html) > 300)

}
