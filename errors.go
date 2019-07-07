package sesiones 

import (
	"fmt"
)

// ErrAutenticacion significa que se analizaron los datos de usuario y
// contraseña suministrados por el usuario, pero alguno de ellos no era
// correcto.
type ErrAutenticacion struct{
	Msg string
}
func (e ErrAutenticacion) Error() string {
	return fmt.Sprintf("error de autenticación: %v", e.Msg)
}

// ErrCorrespondeBlanquear se da cuando los datos son correctos pero,
// el usuario tiene que cambiar la contraseña
type ErrCorrespondeBlanquear struct{
	Msg string
}
func (e ErrCorrespondeBlanquear) Error() string {
	return fmt.Sprintf("Se debe cambiar la contraseña: %v", e.Msg)
}

// ErrCorrespondeConfirmarMail se da cuando un usuario se loggea, pero el mismo
// no tiene confirmada la dirección de correo electrónico.
type ErrCorrespondeConfirmarMail struct {} 
func (e ErrCorrespondeConfirmarMail) Error() string {
	return "Debe confirmar su dirección de correo electrónico"
} 	
