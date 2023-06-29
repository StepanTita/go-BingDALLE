package config

type Authenticator interface {
	UCookie() string
}

type authenticator struct {
	uCookie string
}

func NewAuthenticator(uCookie string) Authenticator {
	return &authenticator{
		uCookie: uCookie,
	}
}

func (a authenticator) UCookie() string {
	return a.uCookie
}
