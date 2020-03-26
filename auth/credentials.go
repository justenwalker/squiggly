package auth

type Credentials struct {
	Username string
	Password string
	Realm    string
}

func (c Credentials) Credentials(realm string) (Credentials, error) {
	return c, nil
}

type CredentialStore interface {
	Credentials(realm string) (Credentials, error)
}
