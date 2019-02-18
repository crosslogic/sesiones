package sesiones

import (
	"bytes"
	"html/template"

	"github.com/pkg/errors"
)

//
type ConfirmacionBlanqueo struct {
	template        *template.Template
	NombreUsuario   string
	URLConfirmacion string
}

func NewConfirmacionBlanqueo() (c *ConfirmacionBlanqueo, err error) {
	c = &ConfirmacionBlanqueo{}

	// Leo el html del template
	tpl := template.New("")
	tpl, err = tpl.Parse(tplString)
	if err != nil {
		return c, errors.Wrap(err, "parseando template ConfirmacionBlanqueo")
	}

	c.template = tpl

	return
}

func (c *ConfirmacionBlanqueo) Body(nombre, url string) (html string, err error) {

	datos := struct {
		Nombre          string
		URLConfirmacion string
	}{
		Nombre:          nombre,
		URLConfirmacion: url,
	}

	out := &bytes.Buffer{}
	err = c.template.Execute(out, datos)
	if err != nil {
		return html, errors.Wrap(err, "ejecutando template NewConfirmacionBlanqueo")
	}

	html = out.String()

	return
}

var tplString = `
<!DOCTYPE html>
<html>

<head>
    <link href="https://fonts.googleapis.com/css?family=Roboto:300,400" rel="stylesheet">
    <style>
        body {
            background-color: lightblue;

        }

        * {
            font-family: "Roboto", sans-serif;
            font-weight: 300;
        }

        div#main {
            max-width: 700px;
            background-color: white;
            margin: auto;
            padding: 80px;
            height: 95%;
        }

        p#rechazar {}
    </style>
</head>

<body>
    <div id='main'>
        <p> Hola {{ .Nombre }}!</p>
        <p>
            Para continuar con el proceso de blanqueo de contraseña, haz clic <a href='{{ .URLConfirmacion }}'>AQUÍ</a>.
        </p>

        <p id="rechazar">
            Si tú no has realizado la solicitud de blanqueo de contraseña haz clic aquí.
        </p>
    </div>

    <br>
</body>
</html>
`
