package auth

type Credentials struct {
	AccessKey string
	SecretKey string
}

func GetNewCredentials() (Credentials, error) {
	return Credentials{}, nil
}
