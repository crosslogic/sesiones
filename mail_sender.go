package sesiones

import "github.com/go-mail/mail"

// DefaultMailSender es la implementación estandar del mail sender.
type DefaultMailSender struct {
	Dialer *mail.Dialer
}

// NewDefaultMailSender creau un MailSender.
func NewDefaultMailSender(host string, port int, user, pass string) (sender *DefaultMailSender) {

	d := mail.NewDialer(host, port, user, pass)
	sender = &DefaultMailSender{}
	sender.Dialer = d

	return
}

// Send envía mail con los datos ingresados
func (d *DefaultMailSender) Send(to, from, subject, body string) (err error) {
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Send the email to
	err = d.Dialer.DialAndSend(m)
	return

}
