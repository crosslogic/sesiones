package sesiones

import (
	"bytes"
	"html/template"

	"github.com/pkg/errors"
)

// MailTemplate es el encargado de generar el HTML de los mails que se le
// enviarán al usuario para blanquear contraseña o confirmar usuario.
type MailTemplate struct {
	// template es el template con el que se generará el mail
	template *template.Template
	// frontEndPath es el link que va en el mail. Lleva a una página, esa página hace
	// la llamada al backend
	frontEndPath string
}

// NewMailTemplate crea un nuevo template de mail.
//
// frontEndPath: es la URL a la que lleva el mail, por ejemplo la dirección
// www.sweet.com.ar/#/auth/confirmar_usuario
func NewMailTemplate(t, frontEndPath string) (mt *MailTemplate, err error) {
	mt = &MailTemplate{}

	// Leo el html del template
	tpl := template.New("")
	tpl, err = tpl.Parse(t)
	if err != nil {
		return mt, errors.Wrap(err, "parseando string del template")
	}

	mt.template = tpl
	mt.frontEndPath = frontEndPath
	return
}

func (mt *MailTemplate) body(nombre, idConfirmacion string) (html string, err error) {

	url := mt.frontEndPath + "/?id=" + idConfirmacion

	datos := struct {
		Nombre          string
		URLConfirmacion string
	}{
		Nombre:          nombre,
		URLConfirmacion: url,
	}

	out := &bytes.Buffer{}
	err = mt.template.Execute(out, datos)
	if err != nil {
		return html, errors.Wrap(err, "ejecutando template")
	}

	html = out.String()

	return
}

var defaultBlanqueoTemplate = `
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

var defaultConfirmacionUsuarioTemplate = `
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
        <p> Hola {{ .Nombre }}, gracias por sumarte a Sweet!</p>
        <p>
            Para confirmar tu alta como usuario, haz clic <a href='{{ .URLConfirmacion }}'>AQUÍ</a>.
        </p>

        <p id="rechazar">
            Si tú no has realizado la solicitud de blanqueo de contraseña haz clic aquí.
        </p>
    </div>

    <br>
</body>
</html>
`
