package sesiones

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	EstadoPendienteConfirmación = "Pendiente de confirmación"
	EstadoConfirmado            = "Confirmado"
)

// Usuario es cada usuario que ingresará al sistema
type Usuario struct {
	ID                            string
	Nombre                        string
	Apellido                      string
	Hash                          string
	BlanquearProximoIngreso       bool
	Estado                        string
	UltimaActualizacionContraseña time.Time
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
}

const (
	MotivoCreacion = "Creación"
	MotivoBlanqueo = "Blanqueo"
)

type UsuarioConfirmacion struct {
	ID                uuid.UUID
	UserID            string
	CreatedAt         time.Time
	Motivo            string
	Confirmada        bool
	FechaConfirmacion time.Time
}

// TableName devuelve el nombre de la tabla en la base de datos
func (u UsuarioConfirmacion) TableName() string {
	return "usuario_confirmaciones"
}

var (

	// ErrDirectivaPassword significa que la contraseña ingresada no cumplía con
	// las directivas establecidas.
	ErrDirectivaPassword = errors.New("error de validación")
)

// ErrAutenticacion significa que se analizaron los datos de usuario y
// contraseña suministrados por el usuario, pero alguno de ellos no era
// correcto.
type ErrAutenticacion struct{}

func (e ErrAutenticacion) Error() string {
	return "usuario o contraseña incorrecta"
}

// ErrCorrespondeBlanquear se da cuando los datos son correctos pero,
// el usuario tiene que cambiar la contraseña
type ErrCorrespondeBlanquear struct{}

func (e ErrCorrespondeBlanquear) Error() string {
	return "Se debe blanquear la contraseña"
}

// Borrar borra el usuario de la tabla de usuarios.
func (h *Handler) Borrar(u Usuario) error {
	return h.db.Delete(&u).Error
}

// blanquearPassword le pone la nueva contraseña al usuario. No realiza control
func (h *Handler) blanquearPassword(usuarioID, nuevaContraseña string, blanquearLuego bool) (err error) {
	usuario := Usuario{}

	// Traigo el usuario de la base de datos
	usuario.ID = usuarioID
	err = h.db.Where("id = ?", usuarioID).First(&usuario).Error
	if err != nil {
		return errors.Wrap(err, "buscando usuario")
	}

	// Coinciden
	usuario.Hash = calcularHash(nuevaContraseña)
	usuario.UltimaActualizacionContraseña = time.Now()
	usuario.BlanquearProximoIngreso = blanquearLuego
	fmt.Println("Blanqueando usuario... ", usuario)
	// Persisto
	err = h.db.Save(&usuario).Error
	if err != nil {
		return errors.Wrap(err, "al intentar blanquear password")
	}

	return nil
}

// EstaVigentePassword corrobora que la antiguedad de la clave sea menor
// al máximo de dias permitidos.
func (h *Handler) estaVigentePassword(usuarioID string) (ok bool, err error) {

	// Traigo el usuario de la base de datos
	usuario := Usuario{}
	usuario.ID = usuarioID
	err = h.db.Where("id = ?", usuarioID).First(&usuario).Error
	if err != nil {
		return false, errors.Wrap(err, "buscando usuario")
	}

	// Si no caduca nunca
	if h.PassValidez == 0 {
		return true, nil
	}

	vencimiento := usuario.UltimaActualizacionContraseña.Add(h.PassValidez)
	if vencimiento.After(time.Now()) {
		return true, nil
	}

	return false, nil
}

// calcularHash genera un hash en base al string del password.
func calcularHash(password string) (hash string) {
	enBytes := sha256.Sum256([]byte(password))
	return hex.EncodeToString(enBytes[:])
}

// ExisteUsuario corrobora si el id de usuario ingresado se encuentra en la base de datos.
func (h *Handler) existeUsuario(userID string) (usuario Usuario, existe bool, err error) {

	err = h.db.Where("id = ?", userID).First(&usuario).Error
	if err == gorm.ErrRecordNotFound {
		return usuario, false, nil
	}

	return usuario, true, err

}

//Compara el string de la password con el hash de la base de datos.
func compararPaswords(password string, hashDB string) error {
	hash := calcularHash(password)
	if hash != hashDB {

		//fmt.Println("Comparando: ", hash)
		//fmt.Println("con       : ", hashDB)

		return errors.Wrap(ErrAutenticacion{}, "Al chequear la contraseña")
	}
	return nil
}

// checkPass prueba si la contraseña y el usuario son correctos. No hace ninguna acción.
func (h *Handler) checkPass(userID, password string) error {
	// Corroboro que exista el usuario
	usuario, existe, err := h.existeUsuario(userID)
	if err != nil {
		return err
	}

	if existe == false {
		return errors.Wrap(ErrAutenticacion{}, "no se encontró el usuario "+userID)
	}

	// Corroboro que coincida la password.
	// "4cf6829aa93728e8f3c97df913fb1bfa95fe5810e2933a05943f8312a98d9cf2",
	err = compararPaswords(password, usuario.Hash)
	if err != nil {
		return errors.Wrap(ErrAutenticacion{}, "contraseña incorrecta")
	}

	if usuario.BlanquearProximoIngreso {
		return ErrCorrespondeBlanquear{}
	}

	// ¿Está vencida la clave? =>
	vigente, err := h.estaVigentePassword(userID)
	if err != nil {
		return errors.Wrap(err, "no se pudo corroborar si la contraseña estaba vigente")
	}
	if !vigente {
		return errors.Wrap(ErrCorrespondeBlanquear{}, "la contraseña ha caducado")
	}

	return nil
}

func (h *Handler) GetByID(userID string) (u Usuario, err error) {

	err = h.db.Find(&u, "id = ?", userID).Error
	return
}
