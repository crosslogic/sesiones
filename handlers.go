package sesiones

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// MailSender implementa el envío de un mail.
type MailSender interface {
	Send(to, from, subject, body string) error
	// SenderAlias es el nombre que aparece en FROM
	SenderAlias() string
}

type Handler struct {
	db        *gorm.DB
	secretKey []byte

	PassMaxLength int
	PassMinLength int
	PassValidez   time.Duration

	DuracionSesion time.Duration

	MailBlanqueo            *MailTemplate
	MailConfirmacionUsuario *MailTemplate
	MailSender              MailSender
}

// New instancia un nuevo handler de sesiones.
func New(
	secretKey []byte,
	db *gorm.DB,
	blanqueoTpl, confirmacionTpl *MailTemplate,
	sender MailSender,
) (h *Handler, err error) {

	h = &Handler{}
	h.db = db
	h.secretKey = secretKey

	// Mail de blanqueo de contraseña
	if blanqueoTpl == nil {
		return nil, errors.New("no se ingresó template de blanqueo")
	}
	h.MailBlanqueo = blanqueoTpl

	// Mail de confirmación de usuario
	if confirmacionTpl == nil {
		return nil, errors.New("no se ingresó template de confirmación de usuario")
	}
	h.MailConfirmacionUsuario = confirmacionTpl

	h.MailSender = sender

	// Datos por defecto PASSWORD
	h.PassMaxLength = 40
	h.PassMinLength = 6
	h.PassValidez = 30 * time.Hour * 24

	// Datos por defecto SESION
	h.DuracionSesion = time.Minute * 30

	return
}

const (
	pathNuevoUsuario      = "nuevo_usuario"
	pathConfirmarUsuario  = "confirmar_usuario"
	pathSolicitarBlanqueo = "solicitar_blanqueo"
	pathConfirmarBlanqueo = "confirmar_blanqueo"
	pathCerrarSesion      = "cerrar_sesion"
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var route string

	// Saco los "/" iniciales y finales
	r.URL.Path = strings.Trim(r.URL.Path, "/")

	// Separo el path por "/"
	paths := strings.Split(r.URL.Path, "/")

	switch len(paths) {
	case 0:
		http.Error(w, "", http.StatusNotFound)
	case 1:
		route = paths[0]
	default:
		// Tomo el último
		route = paths[len(paths)-1]
	}

	switch route {
	case pathNuevoUsuario:
		h.NuevoUsuario()(w, r)
	case pathConfirmarUsuario:
		h.ConfirmarUsuario()(w, r)
	case pathSolicitarBlanqueo:
		h.SolicitarBlanqueo()(w, r)
	case pathConfirmarBlanqueo:
		h.ConfirmarBlanqueo()(w, r)
	case pathCerrarSesion:
		h.CerrarSesion()(w, r)
	default:
		http.Error(w, "", http.StatusNotFound)
	}

}

// NuevoUsuario da de alta un usuario, envía el mail y lo deja en estado
// "Pendiente de Confirmación".
func (h *Handler) NuevoUsuario() http.HandlerFunc {

	request := struct {
		Nombre   string
		Apellido string
		Mail     string
		Pass     string
	}{}

	return func(w http.ResponseWriter, r *http.Request) {

		// Leo el request
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			httpErr(w, errors.Wrap(err, ""), http.StatusBadRequest)
		}

		// Creo el struct Usuario
		u := Usuario{}
		u.ID = request.Mail
		u.Nombre = request.Nombre
		u.Apellido = request.Apellido

		//Que tenga id
		if u.ID == "" {
			httpErr(w, errors.New("debe ingresar un mail"), http.StatusBadRequest)
			return
		}

		//Que no exista en la base

		// Que tenga nombre
		if u.Nombre == "" {
			httpErr(w, errors.New("debe ingresar un nombre"), http.StatusBadRequest)
			return
		}

		// Que el id ingresado no exista.
		_, existe, err := h.existeUsuario(u.ID)
		if err != nil {
			httpErr(w, errors.Wrap(err, "corroborando existencia del usuario"), http.StatusInternalServerError)
			return
		}
		if existe == true {
			httpErr(w, errors.Errorf("ya existe un usuario con mail %v", u.ID), http.StatusInternalServerError)
			return
		}

		// Le pego el hash de la password.
		u.Hash = calcularHash(request.Pass)
		u.UltimaActualizacionContraseña = time.Now()
		u.BlanquearProximoIngreso = false
		u.Estado = EstadoPendienteConfirmación

		// Persisto
		err = h.db.Create(&u).Error
		if err != nil {
			httpErr(w, errors.Wrap(err, "persistiendo usuario en base de datos"), http.StatusInternalServerError)
			return
		}

		// Creo el registro con el codigo de confirmación.
		conf := UsuarioConfirmacion{}
		conf.ID, _ = uuid.NewV4()
		conf.UserID = u.ID
		conf.Motivo = MotivoCreacion
		tx := h.db.Begin()
		err = tx.Create(&conf).Error
		if err != nil {
			tx.Rollback()
			httpErr(w, errors.Wrap(err, "creando confiramación de usuario"), http.StatusInternalServerError)
			return
		}

		// Envío el mail con el link para confirmar usuario
		body, err := h.MailConfirmacionUsuario.body(u.Nombre, conf.ID.String())
		if err != nil {
			tx.Rollback()
			httpErr(w, errors.Wrap(err, "creando body de mail usuario"), http.StatusInternalServerError)
			return
		}
		h.MailSender.Send(u.ID, h.MailSender.SenderAlias(), "Confirmación de usuario", body)

		err = tx.Commit().Error
		if err != nil {
			httpErr(w, errors.Wrap(err, "confirmando la transacción"), http.StatusInternalServerError)
			return
		}

		return
	}
}

// ConfirmarUsuario tilda el usuario como "Confirmado". Tiene que hacerlo con
// el link que le llega al mail.
func (h *Handler) ConfirmarUsuario() http.HandlerFunc {

	request := struct {
		ID uuid.UUID
	}{}
	return func(w http.ResponseWriter, r *http.Request) {

		// Leo el ID de la confirmación
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			httpErr(w, errors.Wrapf(err, "no se pudo leer el ID"), http.StatusBadRequest)
			return
		}

		// Busco que esté disponible esa confirmación
		c := UsuarioConfirmacion{}
		err = h.db.First(&c, "id = ? AND motivo = ?", request.ID, MotivoCreacion).Error
		if err != nil {
			httpErr(w, errors.Wrap(err, "no se pudo obtener el registro de confirmación"), http.StatusInternalServerError)
			return
		}

		if c.Confirmada {
			httpErr(w, errors.New("la cuenta ya estaba confirmada"), http.StatusInternalServerError)
			return
		}

		c.Confirmada = true
		c.FechaConfirmacion = time.Now()

		tx := h.db.Begin()

		// Grabo UsuarioConfirmación
		err = tx.Save(&c).Error
		if err != nil {
			tx.Rollback()
			httpErr(w, errors.Wrap(err, "actualizando estado de solicitud"), http.StatusInternalServerError)
			return
		}

		// Cambio el estado en Usuario
		err = tx.
			Model(&Usuario{}).
			Where("id = ?", c.UserID).
			Update(map[string]interface{}{"Estado": EstadoConfirmado}).
			Error
		if err != nil {
			tx.Rollback()
			httpErr(w, errors.Wrap(err, "persistiendo usuario"), http.StatusInternalServerError)
			return
		}

		// Confirmo tx
		err = tx.Commit().Error
		if err != nil {
			httpErr(w, errors.Wrap(err, "confirmando transaccion"), http.StatusInternalServerError)
			return
		}

	}
}

// SolicitarBlanqueo le envía un mail al usuario con un link desde el que
// puede entrar y poner sun nueva clave.
func (h *Handler) SolicitarBlanqueo() http.HandlerFunc {

	request := struct {
		UserID string
	}{}

	return func(w http.ResponseWriter, r *http.Request) {

		// Leo el ID de usuario
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			httpErr(w, errors.Wrap(err, "al leer JSON"), http.StatusBadRequest)
			return
		}

		// Que el id ingresado  exista.
		usuario, existe, err := h.existeUsuario(request.UserID)
		if err != nil {
			httpErr(w, errors.Wrap(err, "corroborando existencia del usuario"), http.StatusInternalServerError)
			return
		}
		if !existe {
			httpErr(w, errors.Errorf("no existe ningún usuario con el mail %v", request.UserID), http.StatusInternalServerError)
			return
		}

		// Creo el registro con el codigo de confirmación.
		conf := UsuarioConfirmacion{}
		conf.ID, _ = uuid.NewV4()
		conf.UserID = request.UserID
		conf.Motivo = MotivoBlanqueo
		err = h.db.Create(&conf).Error
		if err != nil {
			httpErr(w, errors.Wrap(err, "creando confirmación de usuario"), http.StatusInternalServerError)
			return
		}

		// Envio un mail con el link que lleva a una página donde puede ingresar la nueva contraseña
		body, err := h.MailBlanqueo.body(usuario.Nombre, conf.ID.String())
		if err != nil {
			httpErr(w, errors.Wrap(err, "generando el body del mail"), http.StatusInternalServerError)
			return
		}
		err = h.MailSender.Send(conf.UserID, h.MailSender.SenderAlias(), "Confirmacion de blanqueo de contraseña", body)
		if err != nil {
			httpErr(w, errors.Wrap(err, "enviando el mail de confirmación"), http.StatusInternalServerError)
			return
		}
	}
}

// ConfirmarBlanqueo se llama desde la página /blanquear
func (h *Handler) ConfirmarBlanqueo() http.HandlerFunc {

	request := struct {
		CodigoConfirmacion string
		Pass               string
	}{}

	return func(w http.ResponseWriter, r *http.Request) {

		// Leo el ID de la confirmación
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			httpErr(w, errors.Wrap(err, "al leer JSON"), http.StatusBadRequest)
			return
		}

		// Busco que esté disponible esa confirmación
		c := UsuarioConfirmacion{}
		err = h.db.First(&c, "id = ? AND motivo = ?", request.CodigoConfirmacion, MotivoBlanqueo).Error
		if err != nil {
			httpErr(w, errors.Wrap(err, "no se pudo obtener el registro de confirmación"), http.StatusInternalServerError)
			return
		}

		if c.Confirmada {
			httpErr(w, errors.New("el blanqueo ya se había realizado"), http.StatusInternalServerError)
			return
		}

		// Estamos ok, procedemos con el blanqueo

		// Grabo UsuarioConfirmación
		c.Confirmada = true
		c.FechaConfirmacion = time.Now()
		err = h.db.Save(&c).Error
		if err != nil {
			httpErr(w, errors.New("confirmando la confirmación"), http.StatusInternalServerError)
			return
		}

		// Cambio el hash de la constraseña
		err = h.blanquearPassword(c.UserID, request.Pass, false)
		if err != nil {
			httpErr(w, errors.New("blanqueando password"), http.StatusInternalServerError)
			return
		}

	}
}

func httpErr(w http.ResponseWriter, err error, errCode int, msg ...string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(errCode)
	m := ""
	if len(msg) == 1 {
		m = msg[0] + ":"
	}
	fmt.Fprintln(w, m, err)
}

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
